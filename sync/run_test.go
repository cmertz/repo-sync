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
