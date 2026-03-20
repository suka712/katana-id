package check

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// CheckNpm checks whether an npm package name is available.
// A 404 from the registry means the package does not exist.
func CheckNpm(ctx context.Context, name string) Result {
	res := Result{Name: name, Platform: Npm}
	slug := strings.ToLower(name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://registry.npmjs.org/%s", slug), nil)
	if err != nil {
		res.Err = err.Error()
		return res
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	defer resp.Body.Close()

	res.Available = resp.StatusCode == http.StatusNotFound
	return res
}
