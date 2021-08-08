package sync_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/cmertz/repo-sync/sync"
)

var ErrTestDummy = errors.New("test")

type dummySyncer struct {
	calls int32
}

func (d *dummySyncer) Do(context.Context) error {
	atomic.AddInt32(&d.calls, 1)

	return nil
}

type dummyErrorSyncer struct{ error }

func (d dummyErrorSyncer) Do(context.Context) error {
	return error(d)
}

type dummyBlockingSyncer struct {
	latch chan struct{}
}

func (d dummyBlockingSyncer) Do(context.Context) error {
	d.latch <- struct{}{}
	<-d.latch

	return nil
}

func TestRun(t *testing.T) {
	const syncCount = 10

	var syncs []sync.Syncer

	d := dummySyncer{}

	for i := 0; i < syncCount; i++ {
		syncs = append(syncs, &d)
	}

	sync.Run(context.Background(), syncs, 100, func(error) {})

	if d.calls != syncCount {
		t.Errorf("expected %d calls, actual %d calls", syncCount, d.calls)
	}
}

func TestRun_errorHandler(t *testing.T) {
	syncs := []sync.Syncer{&dummyErrorSyncer{ErrTestDummy}}

	var err error

	sync.Run(context.Background(), syncs, 1, func(e error) { err = e })

	if err.Error() != ErrTestDummy.Error() {
		t.Errorf("expected error %s, actual error %s", ErrTestDummy, err)
	}
}

// nolint: wsl
func TestRun_cancelation(t *testing.T) {
	d := dummyBlockingSyncer{
		latch: make(chan struct{}),
	}

	done := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())

	// the test relies on `sync.Run` with a single
	// goroutine
	const concurrency = 1

	go func() {
		sync.Run(ctx, []sync.Syncer{d, d}, concurrency, func(error) {})
		done <- struct{}{}
	}()

	<-d.latch
	cancel()
	d.latch <- struct{}{}
	<-done

	// there is no need for expectations, in case
	// of failure, the test will hang and time out
}
