package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/trnahnh/katana-id/internal/db/ent/session"
	"github.com/trnahnh/katana-id/util"
)

type contextKey string

const EmailKey contextKey = "email"

// RequireAuth validates the session cookie and injects the user's email into ctx.
func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err != nil {
			util.WriteJSON(w, http.StatusUnauthorized, util.ErrorResponse{Error: "Unauthorized"})
			return
		}

		sess, err := h.DB.Session.Query().
			Where(session.Token(cookie.Value), session.ExpiresAtGT(time.Now())).
			Only(r.Context())
		if err != nil {
			util.WriteJSON(w, http.StatusUnauthorized, util.ErrorResponse{Error: "Unauthorized"})
			return
		}

		ctx := context.WithValue(r.Context(), EmailKey, sess.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
