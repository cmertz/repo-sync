package sync

import (
	"context"
	"errors"
	"testing"
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
	tpl1 := Template{
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
	if s1[0].local != "/src/a" {
		t.Errorf("expected /src/a, actual %s", s1[0].local)
	}
	if s1[1].local != "/src/b" {
		t.Errorf("expected /src/b, actual %s", s1[1].local)
	}

	tpl2 := Template{
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
