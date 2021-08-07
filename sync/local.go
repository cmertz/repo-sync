package sync

import (
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
)

// local represents a local working copy of a git repository
type local string

func (l local) ensureParentPathExists() error {
	elems := strings.Split(string(l), string(os.PathSeparator))
	parent := strings.Join(elems[:len(elems)-1], "/")

	return os.MkdirAll(parent, 0700)
}

func (l local) open() (*git.Repository, error) {
	return git.PlainOpen(string(l))
}

func (l local) exists() bool {
	_, err := l.open()

	return err == nil
}

func (l local) path() string {
	return string(l)
}
