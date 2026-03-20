package check

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// GitHubChecker returns a CheckFn that checks whether a GitHub org or user
// with the given name exists. token is optional but raises the rate limit
// from 60 to 5000 requests/hour.
func GitHubChecker(token string) CheckFn {
	return func(ctx context.Context, name string) Result {
		slug := strings.ToLower(name)
		res := Result{Name: name, Platform: GitHub}

		taken, err := githubSlugExists(ctx, token, "orgs", slug)
		if err != nil {
			res.Err = err.Error()
			return res
		}
		if taken {
			res.Available = false
			return res
		}

		taken, err = githubSlugExists(ctx, token, "users", slug)
		if err != nil {
			res.Err = err.Error()
			return res
		}
		res.Available = !taken
		return res
	}
}

func githubSlugExists(ctx context.Context, token, kind, slug string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://api.github.com/%s/%s", kind, slug), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
