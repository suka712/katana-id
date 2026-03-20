package check

import (
	"context"
	"sync"
)

type Platform string

const (
	Domain Platform = "domain"
	X      Platform = "x"
)

type Result struct {
	Name      string
	Platform  Platform
	Available bool
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
