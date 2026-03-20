package check

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/trnahnh/katana-id/util"
)

type Handler struct {
	Store        *Store
	GitHubToken  string // optional — raises GH rate limit from 60 to 5000 req/hr
	BraveAPIKey  string // optional — enables search presence checks
	TwitterToken string // optional — enables Twitter/X checks
}

type checkRequest struct {
	Query string `json:"query"`
}

type checkResponse struct {
	ID    string   `json:"id"`
	Names []string `json:"names"`
	Total int      `json:"total"`
}

// Check handles POST /check.
// Extracts a name from the query, fans out goroutines, returns a session ID
// the client uses to open the SSE stream.
func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	var req checkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		util.WriteJSON(w, http.StatusBadRequest, util.ErrorResponse{Error: "query is required"})
		return
	}

	name := extractName(req.Query)
	if name == "" {
		util.WriteJSON(w, http.StatusBadRequest, util.ErrorResponse{Error: "could not extract a name — try 'my app called X'"})
		return
	}

	checkers := h.buildCheckers()
	id, sess := h.Store.Create(len(checkers))

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		results := Run(ctx, name, checkers)
		for result := range results {
			sess.Results <- result
		}
		close(sess.Results)
	}()

	// Auto-cleanup if nobody connects within 5 minutes
	go func() {
		time.Sleep(5 * time.Minute)
		h.Store.Delete(id)
	}()

	util.WriteJSON(w, http.StatusCreated, checkResponse{
		ID:    id,
		Names: []string{name},
		Total: len(checkers),
	})
}

// Stream handles GET /check/{id} as an SSE endpoint.
// Drains the session channel and writes one event per result.
func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, ok := h.Store.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	defer h.Store.Delete(id)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering

	for {
		select {
		case result, open := <-sess.Results:
			if !open {
				fmt.Fprintf(w, "data: {\"done\":true}\n\n")
				flusher.Flush()
				return
			}
			data, _ := json.Marshal(result)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// buildCheckers composes the full list of check functions based on available config.
func (h *Handler) buildCheckers() []CheckFn {
	checkers := DomainCheckers()
	checkers = append(checkers, GitHubChecker(h.GitHubToken))
	checkers = append(checkers, CheckNpm)
	checkers = append(checkers, CheckReddit)
	if h.TwitterToken != "" {
		checkers = append(checkers, TwitterChecker(h.TwitterToken))
	}
	if h.BraveAPIKey != "" {
		checkers = append(checkers, SearchChecker(h.BraveAPIKey))
	}
	return checkers
}

// extractName pulls a usable slug from a free-form query.
//
//	"ruffle"                         → "ruffle"
//	"tinder for dogs called Ruffle"  → "ruffle"
//	"tinder for dogs"                → "tinder"  (first word fallback)
func extractName(query string) string {
	q := strings.TrimSpace(query)
	patterns := []string{
		`(?i)called\s+"([^"]+)"`,
		`(?i)named\s+"([^"]+)"`,
		`(?i)called\s+([\w-]+)`,
		`(?i)named\s+([\w-]+)`,
	}
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		if m := re.FindStringSubmatch(q); len(m) > 1 {
			return sanitize(m[1])
		}
	}
	words := strings.Fields(q)
	if len(words) == 0 {
		return ""
	}
	return sanitize(words[0])
}

func sanitize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
