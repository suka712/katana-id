package check

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// userAgent is sent on every outbound check. Several registries (crates.io,
// Reddit, ...) reject requests that use Go's default user agent.
const userAgent = "katana-id-checker/1.0 (+https://katanaid.com)"

// httpClient is shared by all checkers so connections are pooled across the
// fan-out. Each individual check is still bounded by the request context.
var httpClient = &http.Client{Timeout: 12 * time.Second}

// getStatus issues a single GET and returns the response status code.
func getStatus(ctx context.Context, url string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

// statusChecker builds a CheckFn for services that answer 200 when a name is
// taken and 404 when it is free. urlFmt must contain a single %s for the slug.
func statusChecker(platform Platform, urlFmt string) CheckFn {
	return func(ctx context.Context, name string) Result {
		res := Result{Name: name, Platform: platform}
		status, err := getStatus(ctx, fmt.Sprintf(urlFmt, strings.ToLower(name)))
		if err != nil {
			res.Err = err.Error()
			return res
		}
		switch status {
		case http.StatusOK:
			res.Available = false
		case http.StatusNotFound:
			res.Available = true
		default:
			res.Err = fmt.Sprintf("unexpected status %d", status)
		}
		return res
	}
}

// gitlabChecker queries the GitLab users API, which returns a (possibly empty)
// JSON array. An empty array means the username is free.
func gitlabChecker(ctx context.Context, name string) Result {
	res := Result{Name: name, Platform: GitLab}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://gitlab.com/api/v4/users?username=%s", strings.ToLower(name)), nil)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	defer resp.Body.Close()

	var users []struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		res.Err = err.Error()
		return res
	}
	res.Available = len(users) == 0
	return res
}

// keybaseChecker uses Keybase's public lookup API. them is null/empty when no
// account owns the username.
func keybaseChecker(ctx context.Context, name string) Result {
	res := Result{Name: name, Platform: Keybase}
	status, err := getStatus(ctx,
		fmt.Sprintf("https://keybase.io/_/api/1.0/user/lookup.json?username=%s&fields=basics",
			strings.ToLower(name)))
	if err != nil {
		res.Err = err.Error()
		return res
	}
	// Keybase answers 200 for existing users and 404 for unknown ones.
	switch status {
	case http.StatusOK:
		res.Available = false
	case http.StatusNotFound:
		res.Available = true
	default:
		res.Err = fmt.Sprintf("unexpected status %d", status)
	}
	return res
}

// Options configures which optional, key-gated checkers are included.
type Options struct {
	GitHubToken  string
	TwitterToken string
	BraveAPIKey  string
}

// DefaultCheckers assembles the full fan-out: domain TLDs + GitHub + every
// keyless registry/handle checker, plus any optional keyed providers. With the
// defaults this is 21 external integrations per name.
func DefaultCheckers(opts Options) []CheckFn {
	checkers := DomainCheckers()
	checkers = append(checkers, GitHubChecker(opts.GitHubToken))
	checkers = append(checkers, RegistryCheckers()...)
	if opts.TwitterToken != "" {
		checkers = append(checkers, TwitterChecker(opts.TwitterToken))
	}
	if opts.BraveAPIKey != "" {
		checkers = append(checkers, SearchChecker(opts.BraveAPIKey))
	}
	return checkers
}

// RegistryCheckers returns the keyless registry/handle checkers. Combined with
// the domain fan-out and the optional keyed providers (GitHub token, Twitter,
// Brave), a single generation orchestrates well over 19 external API calls.
func RegistryCheckers() []CheckFn {
	return []CheckFn{
		CheckNpm,
		statusChecker(PyPI, "https://pypi.org/pypi/%s/json"),
		statusChecker(RubyGems, "https://rubygems.org/api/v1/gems/%s.json"),
		statusChecker(Crates, "https://crates.io/api/v1/crates/%s"),
		statusChecker(DockerHub, "https://hub.docker.com/v2/users/%s/"),
		statusChecker(Homebrew, "https://formulae.brew.sh/api/formula/%s.json"),
		statusChecker(DevTo, "https://dev.to/api/users/by_username?url=%s"),
		CheckReddit,
		gitlabChecker,
		keybaseChecker,
	}
}
