// Package brand orchestrates the end-to-end flow: turn a prompt into a Gemini
// brand concept, fan out availability checks across every platform via
// goroutines, stream results over SSE, persist the kit, and render a PDF.
package brand

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/trnahnh/katana-id/internal/auth"
	"github.com/trnahnh/katana-id/internal/check"
	"github.com/trnahnh/katana-id/internal/db/ent"
	"github.com/trnahnh/katana-id/internal/db/ent/user"
	"github.com/trnahnh/katana-id/internal/gemini"
	"github.com/trnahnh/katana-id/internal/model"
	"github.com/trnahnh/katana-id/internal/report"
	"github.com/trnahnh/katana-id/internal/trust"
	"github.com/trnahnh/katana-id/util"
)

// maxCheckNames caps how many candidate names get the full platform fan-out, so
// a single generation stays fast while still orchestrating dozens of API calls.
const maxCheckNames = 4

type Handler struct {
	DB        *ent.Client
	Gemini    *gemini.Client // nil disables AI; a local generator is used instead
	Store     *check.Store
	CheckOpts check.Options
}

type generateRequest struct {
	Prompt      string            `json:"prompt"`
	Fingerprint string            `json:"fingerprint"`
	Components  map[string]string `json:"components"`
}

type generateResponse struct {
	ID      string             `json:"id"`
	Concept model.BrandConcept `json:"concept"`
	Names   []string           `json:"names"`
	Total   int                `json:"total"`
	Trust   trust.Score        `json:"trust"`
}

// Generate handles POST /generate: it produces a brand concept, persists a kit,
// kicks off the availability fan-out, and returns an ID for streaming + PDF.
func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Prompt) == 0 {
		util.WriteJSON(w, http.StatusBadRequest, util.ErrorResponse{Error: "prompt is required"})
		return
	}
	if len(req.Prompt) > 500 {
		req.Prompt = req.Prompt[:500]
	}

	ctx := r.Context()
	email, _ := ctx.Value(auth.EmailKey).(string)

	score := trust.Evaluate(trust.Signals{
		Fingerprint:   req.Fingerprint,
		Components:    req.Components,
		UserAgent:     r.UserAgent(),
		Authenticated: email != "",
		IP:            trust.ClientIP(r),
	})

	concept := h.generateConcept(ctx, req.Prompt)

	names := concept.Names
	if len(names) > maxCheckNames {
		names = names[:maxCheckNames]
	}

	checkers := check.DefaultCheckers(h.CheckOpts)
	total := len(names) * len(checkers)

	create := h.DB.BrandKit.Create().
		SetPrompt(req.Prompt).
		SetConcept(concept).
		SetTrustScore(float64(score.Value)).
		SetFingerprint(req.Fingerprint)
	if email != "" {
		if u, err := h.DB.User.Query().Where(user.Email(email)).Only(ctx); err == nil {
			create = create.SetOwner(u)
		}
	}
	kit, err := create.Save(ctx)
	if err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "could not save brand kit"})
		return
	}

	id := strconv.Itoa(kit.ID)
	_, sess := h.Store.CreateWithID(id, total)

	go h.runChecks(kit.ID, names, checkers, sess)

	// Reclaim the session if the client never opens the stream.
	go func() {
		time.Sleep(5 * time.Minute)
		h.Store.Delete(id)
	}()

	util.WriteJSON(w, http.StatusCreated, generateResponse{
		ID:      id,
		Concept: concept,
		Names:   names,
		Total:   total,
		Trust:   score,
	})
}

// runChecks fans out every checker across every name, streams each result into
// the session as it lands, and persists the aggregate onto the kit when done.
func (h *Handler) runChecks(kitID int, names []string, checkers []check.CheckFn, sess *check.Session) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	var mu sync.Mutex
	collected := make([]model.Availability, 0, len(names)*len(checkers))

	var wg sync.WaitGroup
	for _, name := range names {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			for res := range check.Run(ctx, name, checkers) {
				a := toAvailability(res)
				mu.Lock()
				collected = append(collected, a)
				mu.Unlock()
				sess.Results <- res
			}
		}(name)
	}
	wg.Wait()
	close(sess.Results)

	if err := h.DB.BrandKit.UpdateOneID(kitID).SetResults(collected).Exec(ctx); err != nil {
		log.Printf("brand: persist results for kit %d: %v", kitID, err)
	}
}

// Stream handles GET /generate/{id}/stream as an SSE endpoint.
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
	w.Header().Set("X-Accel-Buffering", "no")

	for {
		select {
		case res, open := <-sess.Results:
			if !open {
				w.Write([]byte("data: {\"done\":true}\n\n"))
				flusher.Flush()
				return
			}
			data, _ := json.Marshal(toAvailability(res))
			w.Write([]byte("data: "))
			w.Write(data)
			w.Write([]byte("\n\n"))
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// PDF handles GET /kits/{id}/pdf: renders the persisted kit into a PDF report.
func (h *Handler) PDF(w http.ResponseWriter, r *http.Request) {
	kitID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	kit, err := h.DB.BrandKit.Get(r.Context(), kitID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	pdf, err := report.Render(report.Kit{
		Prompt:     kit.Prompt,
		Concept:    kit.Concept,
		Results:    kit.Results,
		TrustScore: kit.TrustScore,
		CreatedAt:  kit.CreatedAt,
	})
	if err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "could not render report"})
		return
	}

	name := "brand"
	if len(kit.Concept.Names) > 0 {
		name = kit.Concept.Names[0]
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"katanaid-"+name+".pdf\"")
	w.Write(pdf)
}

// generateConcept uses Gemini when configured and falls back to a local
// generator so the product keeps working without an API key.
func (h *Handler) generateConcept(ctx context.Context, prompt string) model.BrandConcept {
	if h.Gemini != nil {
		concept, err := h.Gemini.GenerateConcept(ctx, prompt)
		if err == nil {
			return concept
		}
		log.Printf("brand: gemini fell back to local generator: %v", err)
	}
	return fallbackConcept(prompt)
}

func toAvailability(r check.Result) model.Availability {
	return model.Availability{
		Name:      r.Name,
		Platform:  string(r.Platform),
		Available: r.Available,
		Meta:      r.Meta,
		Err:       r.Err,
	}
}
