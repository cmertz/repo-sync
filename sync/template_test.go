package sync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cmertz/repo-sync/sync"
)

type dummyRemoteSource []string

func (d dummyRemoteSource) List(context.Context) ([]string, error) {
	return []string(d), nil
}

type errorRemoteSource struct{ error }

func (e errorRemoteSource) List(context.Context) ([]string, error) {
	return nil, error(e)
}

func TestTemplate_Syncs(t *testing.T) {
	tpl1 := sync.Template{
		RemoteSource: dummyRemoteSource([]string{"a", "b"}),
		LocalPrefix:  "/src",
	}

	s1, err := tpl1.Syncs(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(s1) != 2 {
		t.Errorf("expected 2 results, actual %d results", len(s1))
	}

	if s2, _ := s1[0].(sync.Sync); s2.Local != "/src/a" {
		t.Errorf("expected /src/a, actual %s", s2.Local)
	}

	if s2, _ := s1[1].(sync.Sync); s2.Local != "/src/b" {
		t.Errorf("expected /src/b, actual %s", s2.Local)
	}

	tpl2 := sync.Template{
		// nolint: goerr113
		RemoteSource: errorRemoteSource{errors.New("woops")},
		LocalPrefix:  "/src",
	}

	s2, err := tpl2.Syncs(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}

	if s2 != nil {
		t.Errorf("expected nil result, actual %v", s2)
	}
}
