package check

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// CheckReddit checks whether a Reddit username or subreddit is available.
// Uses the public JSON API — no key needed.
func CheckReddit(ctx context.Context, name string) Result {
	res := Result{Name: name, Platform: Reddit}
	slug := strings.ToLower(name)

	// Check username
	userTaken, err := redditPathExists(ctx, fmt.Sprintf("https://www.reddit.com/user/%s/about.json", slug))
	if err != nil {
		res.Err = err.Error()
		return res
	}
	if userTaken {
		res.Available = false
		return res
	}

	// Check subreddit
	subTaken, err := redditPathExists(ctx, fmt.Sprintf("https://www.reddit.com/r/%s/about.json", slug))
	if err != nil {
		res.Err = err.Error()
		return res
	}
	res.Available = !subTaken
	return res
}

func redditPathExists(ctx context.Context, url string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}
	// Reddit blocks the default Go user-agent
	req.Header.Set("User-Agent", "katana-id-checker/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// TwitterChecker returns a CheckFn that checks Twitter/X handle availability.
// Requires a bearer token from the Twitter Developer Portal (free tier works).
// https://developer.twitter.com/en/docs/twitter-api/users/lookup/api-reference
func TwitterChecker(bearerToken string) CheckFn {
	return func(ctx context.Context, name string) Result {
		res := Result{Name: name, Platform: X}
		slug := strings.ToLower(name)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("https://api.twitter.com/2/users/by/username/%s", slug), nil)
		if err != nil {
			res.Err = err.Error()
			return res
		}
		req.Header.Set("Authorization", "Bearer "+bearerToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			res.Err = err.Error()
			return res
		}
		defer resp.Body.Close()

		res.Available = resp.StatusCode == http.StatusNotFound
		return res
	}
}

// InstagramChecker and TikTokChecker are not implemented.
// Both platforms require authenticated API access or are protected against scraping.
// Options: Rapid API (https://rapidapi.com) has third-party endpoints for both.
