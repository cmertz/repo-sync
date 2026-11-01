package sync_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/cmertz/repo-sync/sync"
)

// commitFile commits a file called name to the repository at dir.
func commitFile(t *testing.T, dir, name string) {
	t.Helper()

	repo, err := git.PlainOpen(dir)
	if err != nil {
		t.Fatal(err)
	}

	w, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, name), []byte(name), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = w.Add(name)
	if err != nil {
		t.Fatal(err)
	}

	_, err = w.Commit("add "+name, &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com", When: time.Now()},
	})
	if err != nil {
		t.Fatal(err)
	}
}

// initRepo creates an empty repository in a temporary directory.
func initRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	_, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestSync_Do(t *testing.T) {
	// go-git serves local remotes through git-upload-pack.
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("needs the git binary: %v", err)
	}

	t.Run("clones, then pulls", func(t *testing.T) {
		upstream := initRepo(t)
		commitFile(t, upstream, "a")

		local := filepath.Join(t.TempDir(), "parent", "copy")
		s := sync.Sync{Remote: sync.Remote(upstream), Local: sync.Local(local)}

		err := s.Do(t.Context())
		if err != nil {
			t.Fatalf("Do() clone error = %v", err)
		}

		_, err = os.Stat(filepath.Join(local, "a"))
		if err != nil {
			t.Fatalf("after clone: %v", err)
		}

		commitFile(t, upstream, "b")

		err = s.Do(t.Context())
		if err != nil {
			t.Fatalf("Do() pull error = %v", err)
		}

		_, err = os.Stat(filepath.Join(local, "b"))
		if err != nil {
			t.Fatalf("after pull: %v", err)
		}

		err = s.Do(t.Context())
		if err != nil {
			t.Fatalf("Do() up to date error = %v", err)
		}
	})

	t.Run("skips an empty remote", func(t *testing.T) {
		local := filepath.Join(t.TempDir(), "copy")
		s := sync.Sync{Remote: sync.Remote(initRepo(t)), Local: sync.Local(local)}

		err := s.Do(t.Context())
		if err != nil {
			t.Fatalf("Do() error = %v", err)
		}

		_, err = os.Stat(local)
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("local copy of an empty remote: %v, want it not to exist", err)
		}
	})

	t.Run("refuses a directory with files", func(t *testing.T) {
		upstream := initRepo(t)
		commitFile(t, upstream, "a")

		local := t.TempDir()

		err := os.WriteFile(filepath.Join(local, "a"), []byte("keep me"), 0o600)
		if err != nil {
			t.Fatal(err)
		}

		s := sync.Sync{Remote: sync.Remote(upstream), Local: sync.Local(local)}

		err = s.Do(t.Context())
		if !errors.Is(err, sync.ErrNotRepository) {
			t.Errorf("Do() error = %v, want %v", err, sync.ErrNotRepository)
		}
	})
}
