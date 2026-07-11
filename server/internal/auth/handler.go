package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/resend/resend-go/v3"

	"github.com/trnahnh/katana-id/internal/db/ent"
	"github.com/trnahnh/katana-id/internal/db/ent/otp"
	"github.com/trnahnh/katana-id/internal/db/ent/session"
	"github.com/trnahnh/katana-id/internal/db/ent/user"
	"github.com/trnahnh/katana-id/util"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type sendOTPRequest struct {
	Email string
}

type successResponse struct {
	Message string `json:"message"`
}

type meResponse struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

type Handler struct {
	DB          *ent.Client
	EmailClient *resend.Client

	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
	ServerURL          string
	FrontendURL        string

	// SecureCookies gates the Secure flag on auth cookies. It must be false in
	// local http dev (browsers drop Secure cookies over http://localhost, which
	// silently breaks the OAuth state round-trip) and true in https production.
	SecureCookies bool
}

const sessionCookie = "session"

// createSessionCookie issues a new session for email and sets it as the "session" cookie.
func (h *Handler) createSessionCookie(ctx context.Context, w http.ResponseWriter, email string) error {
	sess, err := h.DB.Session.Create().
		SetEmail(email).
		SetExpiresAt(time.Now().Add(7 * 24 * time.Hour)).
		Save(ctx)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    sess.Token,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   h.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		util.WriteJSON(w, http.StatusUnauthorized, util.ErrorResponse{Error: "Unauthorized"})
		return
	}

	ctx := r.Context()
	sess, err := h.DB.Session.Query().
		Where(session.Token(cookie.Value), session.ExpiresAtGT(time.Now())).
		Only(ctx)
	if err != nil {
		util.WriteJSON(w, http.StatusUnauthorized, util.ErrorResponse{Error: "Unauthorized"})
		return
	}

	u, err := h.DB.User.Query().Where(user.Email(sess.Email)).Only(ctx)
	if err != nil {
		util.WriteJSON(w, http.StatusUnauthorized, util.ErrorResponse{Error: "Unauthorized"})
		return
	}

	util.WriteJSON(w, http.StatusOK, meResponse{
		Email:    u.Email,
		Username: u.Username,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil && cookie.Value != "" {
		h.DB.Session.Delete().Where(session.Token(cookie.Value)).Exec(r.Context())
	}

	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookie,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	util.WriteJSON(w, http.StatusOK, successResponse{Message: "Logged out"})
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

	code, err := genOTP()
	if err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}

	ctx := r.Context()
	if _, err := h.DB.OTP.Create().
		SetEmail(req.Email).
		SetCode(code).
		SetExpiresAt(time.Now().Add(5 * time.Minute)).
		Save(ctx); err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}

	if err := sendOTP(h.EmailClient, req.Email, code); err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}

	util.WriteJSON(w, http.StatusOK, successResponse{Message: "OTP sent"})
}

type verifyOTPRequest struct {
	Email string
	OTP   string
}

func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteJSON(w, http.StatusBadRequest, util.ErrorResponse{Error: "Invalid request"})
		return
	}

	if !emailRegex.MatchString(req.Email) {
		util.WriteJSON(w, http.StatusBadRequest, util.ErrorResponse{Error: "Invalid email"})
		return
	}

	ctx := r.Context()

	otpRow, err := h.DB.OTP.Query().
		Where(otp.Email(req.Email), otp.ExpiresAtGT(time.Now())).
		Order(ent.Desc(otp.FieldExpiresAt)).
		First(ctx)
	if ent.IsNotFound(err) {
		util.WriteJSON(w, http.StatusUnauthorized, util.ErrorResponse{Error: "Invalid or expired OTP"})
		return
	}
	if err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}

	if otpRow.Code != req.OTP {
		util.WriteJSON(w, http.StatusUnauthorized, util.ErrorResponse{Error: "Invalid or expired OTP"})
		return
	}

	if _, err := h.DB.OTP.Delete().Where(otp.Email(req.Email)).Exec(ctx); err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}

	if err := h.ensureUser(ctx, req.Email); err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}

	if err := h.createSessionCookie(ctx, w, req.Email); err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}

	util.WriteJSON(w, http.StatusOK, successResponse{Message: "OTP verified"})
}

// ensureUser creates an account for email if one does not already exist,
// deriving the username from the local part of the address.
func (h *Handler) ensureUser(ctx context.Context, email string) error {
	exists, err := h.DB.User.Query().Where(user.Email(email)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	username := strings.Split(email, "@")[0]
	_, err = h.DB.User.Create().SetUsername(username).SetEmail(email).Save(ctx)
	return err
}
