package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/trnahnh/katana-id/internal/db/generated"
)

const (
	oauthStateCookie    = "oauth_state"
	oauthRedirectCookie = "oauth_redirect"
)

// beginOAuth issues a CSRF state cookie, optionally stashes a post-login
// redirect path, and sends the browser to the provider's authorize endpoint.
func (h *Handler) beginOAuth(w http.ResponseWriter, r *http.Request, authEndpoint string, params url.Values) {
	state, err := randomToken()
	if err != nil {
		h.redirectWithError(w, r)
		return
	}
	params.Set("state", state)

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/",
		MaxAge:   10 * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	if redirect := r.URL.Query().Get("redirect"); strings.HasPrefix(redirect, "/") {
		http.SetCookie(w, &http.Cookie{
			Name:     oauthRedirectCookie,
			Value:    redirect,
			Path:     "/",
			MaxAge:   10 * 60,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	http.Redirect(w, r, authEndpoint+"?"+params.Encode(), http.StatusFound)
}

// consumeState validates the state param against the CSRF cookie, clearing the cookie either way.
func (h *Handler) consumeState(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie(oauthStateCookie)
	clearCookie(w, oauthStateCookie)
	if err != nil || cookie.Value == "" {
		return false
	}
	return cookie.Value == r.URL.Query().Get("state")
}

// popRedirect reads and clears the stashed post-login path, defaulting to "/".
func (h *Handler) popRedirect(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(oauthRedirectCookie)
	clearCookie(w, oauthRedirectCookie)
	if err != nil || !strings.HasPrefix(cookie.Value, "/") {
		return "/"
	}
	return cookie.Value
}

func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: "/", MaxAge: -1})
}

func (h *Handler) redirectWithError(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, h.FrontendURL+"/signin?error=oauth_failed", http.StatusFound)
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// resolveUserAndLinkProvider ensures a user exists for email and records the
// provider link, matching identity the same way the OTP flow does (by email).
func (h *Handler) resolveUserAndLinkProvider(ctx context.Context, providerName, providerAccountID, email string) error {
	user, err := h.Queries.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		username := strings.Split(email, "@")[0]
		user, err = h.Queries.CreateUser(ctx, gendb.CreateUserParams{Username: username, Email: email})
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return h.Queries.UpsertProvider(ctx, gendb.UpsertProviderParams{
		UserID:            user.ID,
		ProviderName:      providerName,
		ProviderAccountID: providerAccountID,
	})
}

// ─── Google ─────────────────────────────────────────────────────────────────

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.GoogleClientID == "" || h.GoogleClientSecret == "" {
		h.redirectWithError(w, r)
		return
	}

	params := url.Values{
		"client_id":     {h.GoogleClientID},
		"redirect_uri":  {h.ServerURL + "/auth/google/callback"},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"access_type":   {"online"},
		"prompt":        {"select_account"},
	}
	h.beginOAuth(w, r, "https://accounts.google.com/o/oauth2/v2/auth", params)
}

func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.GoogleClientID == "" || h.GoogleClientSecret == "" {
		h.redirectWithError(w, r)
		return
	}
	if !h.consumeState(w, r) {
		h.redirectWithError(w, r)
		return
	}
	redirectPath := h.popRedirect(w, r)

	code := r.URL.Query().Get("code")
	if code == "" {
		h.redirectWithError(w, r)
		return
	}

	ctx := r.Context()

	form := url.Values{
		"client_id":     {h.GoogleClientID},
		"client_secret": {h.GoogleClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {h.ServerURL + "/auth/google/callback"},
	}

	tokRes, err := http.PostForm("https://oauth2.googleapis.com/token", form)
	if err != nil {
		h.redirectWithError(w, r)
		return
	}
	defer tokRes.Body.Close()

	var tok googleTokenResponse
	if tokRes.StatusCode != http.StatusOK || json.NewDecoder(tokRes.Body).Decode(&tok) != nil || tok.AccessToken == "" {
		h.redirectWithError(w, r)
		return
	}

	infoReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		h.redirectWithError(w, r)
		return
	}
	infoReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)

	infoRes, err := http.DefaultClient.Do(infoReq)
	if err != nil {
		h.redirectWithError(w, r)
		return
	}
	defer infoRes.Body.Close()

	var info googleUserInfo
	if infoRes.StatusCode != http.StatusOK || json.NewDecoder(infoRes.Body).Decode(&info) != nil || info.Email == "" || !info.EmailVerified {
		h.redirectWithError(w, r)
		return
	}

	if err := h.resolveUserAndLinkProvider(ctx, "google", info.Sub, info.Email); err != nil {
		h.redirectWithError(w, r)
		return
	}
	if err := h.createSessionCookie(ctx, w, info.Email); err != nil {
		h.redirectWithError(w, r)
		return
	}

	http.Redirect(w, r, h.FrontendURL+redirectPath, http.StatusFound)
}

// ─── GitHub ─────────────────────────────────────────────────────────────────

type githubTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type githubUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (h *Handler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	if h.GitHubClientID == "" || h.GitHubClientSecret == "" {
		h.redirectWithError(w, r)
		return
	}

	params := url.Values{
		"client_id":    {h.GitHubClientID},
		"redirect_uri": {h.ServerURL + "/auth/github/callback"},
		"scope":        {"read:user user:email"},
	}
	h.beginOAuth(w, r, "https://github.com/login/oauth/authorize", params)
}

func (h *Handler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	if h.GitHubClientID == "" || h.GitHubClientSecret == "" {
		h.redirectWithError(w, r)
		return
	}
	if !h.consumeState(w, r) {
		h.redirectWithError(w, r)
		return
	}
	redirectPath := h.popRedirect(w, r)

	code := r.URL.Query().Get("code")
	if code == "" {
		h.redirectWithError(w, r)
		return
	}

	ctx := r.Context()

	form := url.Values{
		"client_id":     {h.GitHubClientID},
		"client_secret": {h.GitHubClientSecret},
		"code":          {code},
		"redirect_uri":  {h.ServerURL + "/auth/github/callback"},
	}

	tokReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		h.redirectWithError(w, r)
		return
	}
	tokReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokReq.Header.Set("Accept", "application/json")

	tokRes, err := http.DefaultClient.Do(tokReq)
	if err != nil {
		h.redirectWithError(w, r)
		return
	}
	defer tokRes.Body.Close()

	var tok githubTokenResponse
	if tokRes.StatusCode != http.StatusOK || json.NewDecoder(tokRes.Body).Decode(&tok) != nil || tok.AccessToken == "" {
		h.redirectWithError(w, r)
		return
	}

	userReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		h.redirectWithError(w, r)
		return
	}
	userReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	userReq.Header.Set("Accept", "application/vnd.github+json")

	userRes, err := http.DefaultClient.Do(userReq)
	if err != nil {
		h.redirectWithError(w, r)
		return
	}
	defer userRes.Body.Close()

	var ghUser githubUser
	if userRes.StatusCode != http.StatusOK || json.NewDecoder(userRes.Body).Decode(&ghUser) != nil || ghUser.ID == 0 {
		h.redirectWithError(w, r)
		return
	}

	email := ghUser.Email
	if email == "" {
		email = h.fetchGitHubPrimaryEmail(ctx, tok.AccessToken)
	}
	if email == "" {
		h.redirectWithError(w, r)
		return
	}

	if err := h.resolveUserAndLinkProvider(ctx, "github", strconv.FormatInt(ghUser.ID, 10), email); err != nil {
		h.redirectWithError(w, r)
		return
	}
	if err := h.createSessionCookie(ctx, w, email); err != nil {
		h.redirectWithError(w, r)
		return
	}

	http.Redirect(w, r, h.FrontendURL+redirectPath, http.StatusFound)
}

// fetchGitHubPrimaryEmail looks up the verified primary email for users whose
// profile email is private. Returns "" if none is found.
func (h *Handler) fetchGitHubPrimaryEmail(ctx context.Context, accessToken string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer res.Body.Close()

	var emails []githubEmail
	if json.NewDecoder(res.Body).Decode(&emails) != nil {
		return ""
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email
		}
	}
	return ""
}
