package sync

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

// Local is the path of a working copy of a git repository.
type Local string

// parentPerm restricts the directories that ensureParentPathExists creates
// to their owner, as the working copies below may hold private code.
const parentPerm = 0o700

func (l Local) ensureParentPathExists() error {
	err := os.MkdirAll(filepath.Dir(l.path()), parentPerm)
	if err != nil {
		return fmt.Errorf("creating parent directory of %s: %w", l, err)
	}

	return nil
}

// isEmpty reports whether nothing exists at the path or it is an empty
// directory.
func (l Local) isEmpty() (bool, error) {
	entries, err := os.ReadDir(l.path())
	if os.IsNotExist(err) {
		return true, nil
	}

	if err != nil {
		return false, fmt.Errorf("reading %s: %w", l, err)
	}

	return len(entries) == 0, nil
}

func (l Local) open() (*git.Repository, error) {
	repo, err := git.PlainOpen(l.path())
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", l, err)
	}

	return repo, nil
}

func (l Local) path() string {
	return string(l)
}
