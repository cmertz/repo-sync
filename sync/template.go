package sync

import (
	"context"
	"fmt"
)

// Template generates a Sync for each repository of a remote source.
type Template struct {
	// RemoteSource lists the URLs of the remote repositories.
	RemoteSource interface {
		List(ctx context.Context) (remoteURLs []string, err error)
	}
	// LocalPrefix is the directory that holds the working copies.
	LocalPrefix string
}

// Syncs lists the remotes of RemoteSource and returns a Sync for each. The
// working copy of a remote is LocalPrefix/<host>/<path>.
func (t Template) Syncs(ctx context.Context) ([]Syncer, error) {
	remotes, err := t.RemoteSource.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing remotes: %w", err)
	}

	res := make([]Syncer, 0, len(remotes))

	for _, url := range remotes {
		r := Remote(url)

		l, err := r.local(t.LocalPrefix)
		if err != nil {
			return nil, err
		}

		res = append(res, Sync{Remote: r, Local: l})
	}

	return res, nil
}
