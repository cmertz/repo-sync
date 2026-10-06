package sync

import (
	"context"
	"sync"
)

// Syncer performs one synchronization.
type Syncer interface {
	Do(ctx context.Context) error
}

// Run calls Do on each of syncs, with up to concurrency calls at a time, and
// passes every error to onError. It calls onError from its own goroutine, one
// error at a time, and returns after the last call. Once ctx is done, Run
// starts no more syncs. It panics if concurrency is less than 1.
func Run(ctx context.Context, syncs []Syncer, concurrency int, onError func(error)) {
	if concurrency < 1 {
		panic("sync.Run: concurrency must be at least 1")
	}

	pending := make(chan Syncer)
	errs := make(chan error)

	go func() {
		defer close(pending)

		for _, s := range syncs {
			select {
			case <-ctx.Done():
				return
			case pending <- s:
			}
		}
	}()

	var workers sync.WaitGroup

	for range concurrency {
		workers.Go(func() {
			for s := range pending {
				err := s.Do(ctx)
				if err != nil {
					errs <- err
				}
			}
		})
	}

	go func() {
		workers.Wait()
		close(errs)
	}()

	for err := range errs {
		onError(err)
	}
}
