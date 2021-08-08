package sync

import (
	"context"
	"sync"
)

type Syncer interface {
	Do(ctx context.Context) error
}

func Run(ctx context.Context, syncs []Syncer, concurrency int, errorHandler func(error)) {
	gitSyncs := make(chan Syncer)
	errs := make(chan error)
	done := make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(concurrency)

	go func() {
	outer:
		for _, s := range syncs {
			select {
			case <-ctx.Done():
				break outer
			case gitSyncs <- s:
			}
		}

		close(gitSyncs)
	}()

	for i := 0; i < concurrency; i++ {
		go func() {
			for s := range gitSyncs {
				err := s.Do(ctx)
				if err != nil {
					errs <- err
				}
			}

			wg.Done()
		}()
	}

	go func() {
		for err := range errs {
			errorHandler(err)
		}

		done <- struct{}{}
	}()

	wg.Wait()

	close(errs)

	<-done
}
