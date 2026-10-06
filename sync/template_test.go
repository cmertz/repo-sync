package sync_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/cmertz/repo-sync/sync"
)

type dummyRemoteSource []string

func (d dummyRemoteSource) List(context.Context) ([]string, error) {
	return d, nil
}

type errorRemoteSource struct{ error }

func (e errorRemoteSource) List(context.Context) ([]string, error) {
	return nil, e.error
}

func TestTemplate_Syncs(t *testing.T) {
	t.Run("maps remotes below the prefix", func(t *testing.T) {
		tpl := sync.Template{
			RemoteSource: dummyRemoteSource{"a", "git@example.com:user/b.git"},
			LocalPrefix:  "/src",
		}

		got, err := tpl.Syncs(t.Context())
		if err != nil {
			t.Fatalf("Syncs() error = %v", err)
		}

		want := []sync.Sync{
			{Remote: "a", Local: sync.Local(filepath.FromSlash("/src/a"))},
			{Remote: "git@example.com:user/b.git", Local: sync.Local(filepath.FromSlash("/src/example.com/user/b"))},
		}

		if len(got) != len(want) {
			t.Fatalf("Syncs() = %v, want %v", got, want)
		}

		for i := range want {
			if got[i] != want[i] {
				t.Errorf("Syncs()[%d] = %v, want %v", i, got[i], want[i])
			}
		}
	})

	t.Run("reports a failed listing", func(t *testing.T) {
		errList := errors.New("woops")
		tpl := sync.Template{RemoteSource: errorRemoteSource{errList}, LocalPrefix: "/src"}

		got, err := tpl.Syncs(t.Context())
		if !errors.Is(err, errList) {
			t.Errorf("Syncs() error = %v, want %v", err, errList)
		}

		if got != nil {
			t.Errorf("Syncs() = %v, want nil", got)
		}
	})

	t.Run("rejects a remote outside the prefix", func(t *testing.T) {
		tpl := sync.Template{RemoteSource: dummyRemoteSource{"git@example.com:../../etc"}, LocalPrefix: "/src"}

		_, err := tpl.Syncs(t.Context())
		if !errors.Is(err, sync.ErrInvalidRemote) {
			t.Errorf("Syncs() error = %v, want %v", err, sync.ErrInvalidRemote)
		}
	})
}
