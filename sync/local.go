package sync

import (
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
)

// Local working copy of a git repository
type Local string

func (l Local) ensureParentPathExists() error {
	const sep = string(os.PathSeparator)
	elems := strings.Split(l.path(), sep)

	parent := strings.Join(elems[:len(elems)-1], sep)

	return os.MkdirAll(parent, 0700)
}

func (l Local) open() (*git.Repository, error) {
	return git.PlainOpen(l.path())
}

func (l Local) exists() bool {
	_, err := l.open()

	return err == nil
}

func (l Local) path() string {
	return string(l)
}
