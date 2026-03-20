package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trnahnh/katana-id/util"
)

type contextKey string

const EmailKey contextKey = "email"

// RequireAuth validates the session cookie and injects the user's email into ctx.
func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			util.WriteJSON(w, http.StatusUnauthorized, util.ErrorResponse{Error: "Unauthorized"})
			return
		}

		token, err := uuid.Parse(cookie.Value)
		if err != nil {
			util.WriteJSON(w, http.StatusUnauthorized, util.ErrorResponse{Error: "Unauthorized"})
			return
		}

		session, err := h.Queries.GetSession(r.Context(), pgtype.UUID{Bytes: token, Valid: true})
		if err != nil {
			util.WriteJSON(w, http.StatusUnauthorized, util.ErrorResponse{Error: "Unauthorized"})
			return
		}

		ctx := context.WithValue(r.Context(), EmailKey, session.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
