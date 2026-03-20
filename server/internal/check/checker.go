package check

import (
	"context"
	"sync"
)

type Platform string

const (
	// Domain TLDs use the pattern "domain.com", "domain.io", etc.
	// Use DomainPlatform() to construct them.
	GitHub    Platform = "github"
	Npm       Platform = "npm"
	Reddit    Platform = "reddit"
	X         Platform = "x"
	Instagram Platform = "instagram"
	TikTok    Platform = "tiktok"
)

func DomainPlatform(tld string) Platform {
	return Platform("domain." + tld)
}

type Result struct {
	Name      string
	Platform  Platform
	Available bool
	Meta      map[string]string // optional extra data (e.g. competitiveness score)
	Err       string
}

type CheckFn func(ctx context.Context, name string) Result

func Run(ctx context.Context, name string, checkers []CheckFn) <-chan Result {
	out := make(chan Result, len(checkers))
	var wg sync.WaitGroup

	for _, fn := range checkers {
		wg.Add(1)
		go func(fn CheckFn) {
			defer wg.Done()
			out <- fn(ctx, name)
		}(fn)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
