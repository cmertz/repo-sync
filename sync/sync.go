// Package sync synchronizes remote git repositories to local working copies.
package sync

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// ErrNotRepository reports a local path that holds files but no git
// repository, so cloning into it could overwrite them.
var ErrNotRepository = errors.New("not empty and not a git repository")

// Sync represents a synchronization from a remote git repository to a local
// working copy.
type Sync struct {
	Remote Remote
	Local  Local
}

// Do clones the remote repository when the local working copy doesn't exist
// yet, and pulls into the working copy otherwise. Pulls only fast-forward.
// Do skips an empty remote repository, since there is nothing to clone.
func (s Sync) Do(ctx context.Context) error {
	repo, err := s.Local.open()
	if errors.Is(err, git.ErrRepositoryNotExists) {
		return s.clone(ctx)
	}

	if err != nil {
		return fmt.Errorf("syncing %s: %w", s.Remote, err)
	}

	return s.pull(ctx, repo)
}

func (s Sync) clone(ctx context.Context) error {
	empty, err := s.Local.isEmpty()
	if err != nil {
		return fmt.Errorf("cloning %s: %w", s.Remote, err)
	}

	if !empty {
		return fmt.Errorf("cloning %s: %s: %w", s.Remote, s.Local, ErrNotRepository)
	}

	err = s.Local.ensureParentPathExists()
	if err != nil {
		return fmt.Errorf("cloning %s: %w", s.Remote, err)
	}

	_, err = git.PlainCloneContext(ctx, s.Local.path(), false, &git.CloneOptions{
		URL: s.Remote.url(),
	})

	// go-git removes what it created, so the next run tries again.
	if errors.Is(err, transport.ErrEmptyRemoteRepository) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("cloning %s into %s: %w", s.Remote, s.Local, err)
	}

	return nil
}

func (s Sync) pull(ctx context.Context, repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("pulling %s into %s: %w", s.Remote, s.Local, err)
	}

	err = w.PullContext(ctx, &git.PullOptions{})
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("pulling %s into %s: %w", s.Remote, s.Local, err)
	}

	return nil
}
