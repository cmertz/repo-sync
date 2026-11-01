package sync_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/cmertz/repo-sync/sync"
)

var errTestDummy = errors.New("test")

type dummySyncer struct {
	calls atomic.Int32
}

func (d *dummySyncer) Do(context.Context) error {
	d.calls.Add(1)

	return nil
}

type dummyErrorSyncer struct{ error }

func (d dummyErrorSyncer) Do(context.Context) error {
	return d.error
}

func TestRun(t *testing.T) {
	t.Run("runs every sync", func(t *testing.T) {
		const syncCount = 10

		var (
			d     dummySyncer
			syncs []sync.Syncer
		)

		for range syncCount {
			syncs = append(syncs, &d)
		}

		sync.Run(t.Context(), syncs, 3, func(err error) {
			t.Errorf("unexpected error: %v", err)
		})

		got := d.calls.Load()
		if got != syncCount {
			t.Errorf("got %d calls, want %d", got, syncCount)
		}
	})

	t.Run("passes every error to the handler", func(t *testing.T) {
		syncs := []sync.Syncer{
			dummyErrorSyncer{errTestDummy},
			&dummySyncer{},
			dummyErrorSyncer{errTestDummy},
		}

		var errs []error

		sync.Run(t.Context(), syncs, 2, func(err error) { errs = append(errs, err) })

		if len(errs) != 2 {
			t.Fatalf("got %d errors, want 2: %v", len(errs), errs)
		}

		for _, err := range errs {
			if !errors.Is(err, errTestDummy) {
				t.Errorf("got error %v, want %v", err, errTestDummy)
			}
		}
	})

	t.Run("panics without workers", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("Run() with concurrency 0 didn't panic")
			}
		}()

		sync.Run(t.Context(), nil, 0, func(error) {})
	})
}
