package check

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// SearchChecker returns a CheckFn that estimates search competitiveness for a name.
// Uses the Brave Search API (free tier: 2000 queries/month).
// Get a key at https://brave.com/search/api/
//
// Result.Available is always false (not applicable for search).
// Result.Meta contains:
//   - "competitiveness": "low" | "medium" | "high"
//   - "total_results":   approximate result count as a string
func SearchChecker(apiKey string) CheckFn {
	return func(ctx context.Context, name string) Result {
		res := Result{Name: name, Platform: "search"}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("https://api.search.brave.com/res/v1/web/search?q=%s&count=1",
				url.QueryEscape(strings.ToLower(name))), nil)
		if err != nil {
			res.Err = err.Error()
			return res
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("X-Subscription-Token", apiKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			res.Err = err.Error()
			return res
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			res.Err = fmt.Sprintf("search API returned %d", resp.StatusCode)
			return res
		}

		var payload struct {
			Web struct {
				TotalCount int `json:"total_count"`
			} `json:"web"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			res.Err = err.Error()
			return res
		}

		count := payload.Web.TotalCount
		res.Meta = map[string]string{
			"total_results":   strconv.Itoa(count),
			"competitiveness": competitiveness(count),
		}
		return res
	}
}

// competitiveness buckets result count into low / medium / high.
// Thresholds are rough estimates and can be tuned.
func competitiveness(count int) string {
	switch {
	case count < 100_000:
		return "low"
	case count < 10_000_000:
		return "medium"
	default:
		return "high"
	}
}
