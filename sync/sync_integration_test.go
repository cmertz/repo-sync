//go:build integration

package sync_test

import (
	"path/filepath"
	"testing"

	"github.com/cmertz/repo-sync/sync"
)

const defaultRemoteURL = "https://github.com/cmertz/repo-sync"

// TestSync_Do_github clones this repository from GitHub over HTTPS, then
// pulls it: the network path that TestSync_Do leaves out.
func TestSync_Do_github(t *testing.T) {
	s := sync.Sync{
		Remote: sync.Remote(defaultRemoteURL),
		Local:  sync.Local(filepath.Join(t.TempDir(), "repo-sync")),
	}

	for range 2 {
		err := s.Do(t.Context())
		if err != nil {
			t.Fatalf("Do() error = %v", err)
		}
	}
}
