package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/resend/resend-go/v3"
	"github.com/trnahnh/katana-id/internal/db/generated"
	"github.com/trnahnh/katana-id/util"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type sendOTPRequest struct {
	Email string
}

type Handler struct {
	Queries     *gendb.Queries
	EmailClient *resend.Client
}

func (h *Handler) SendOTP(w http.ResponseWriter, r *http.Request) {
	var req sendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteJSON(w, http.StatusBadRequest, util.ErrorResponse{Error: "Invalid request"})
		return
	}

	if !emailRegex.MatchString(req.Email) {
		util.WriteJSON(w, http.StatusBadRequest, util.ErrorResponse{Error: "Invalid email"})
		return
	}

	otp, err := genOTP()
	if err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}

  expires := pgtype.Timestamptz{
    Time: time.Now().Add(5 * time.Minute),
    Valid: true,
  }

	if err := h.Queries.CreateOTP(context.Background(), gendb.CreateOTPParams{
		Email: req.Email,
		Otp:   otp,
    ExpiresAt: expires,
	}); err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
    return
	}

	if err := sendOTP(h.EmailClient, req.Email, otp); err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]string{"message": "OTP sent"})
}

func (h *Handler) verifyOTP(w http.ResponseWriter, r *http.Request) {

} 