// package sync provides functionality to synchronize
// remote git repositories to local working copies.
package sync

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// Sync represents a synchronization from a remote
// git repository to a local working copy
type Sync struct {
	Remote Remote
	Local  Local
}

func (s Sync) clone(ctx context.Context) error {
	_, err := git.PlainCloneContext(ctx, s.Local.path(), false, &git.CloneOptions{
		URL: s.Remote.url(),
	})

	// ignore empty remotes
	if errors.Is(err, transport.ErrEmptyRemoteRepository) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("clone: %w", err)
	}

	return nil
}

func (s Sync) pull(ctx context.Context) error {
	repo, err := s.Local.open()
	if err != nil {
		return fmt.Errorf("pull: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("pull: %w", err)
	}

	err = w.PullContext(ctx, &git.PullOptions{})

	// ignore if there's nothing to update
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("pull: %w", err)
	}

	return nil
}

// Do a sync from a remote repository to a local working copy
func (s Sync) Do(ctx context.Context) error {
	if !s.Local.exists() {
		err := s.Local.ensureParentPathExists()
		if err != nil {
			return fmt.Errorf("Do: %w", err)
		}

		err = s.clone(ctx)
		if err != nil {
			return fmt.Errorf("Do: %w", err)
		}
	}

	err := s.pull(ctx)
	if err != nil {
		return fmt.Errorf("Do: %w", err)
	}

	return nil
}
