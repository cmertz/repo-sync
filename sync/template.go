package sync

import (
	"context"
	"fmt"
)

// Template is a template for generating `Sync`s
// TODO: find a better name
type Template struct {
	RemoteSource interface {
		List(ctx context.Context) (remoteURLs []string, err error)
	}
	LocalPrefix string
}

func (t Template) Syncs(ctx context.Context) ([]Syncer, error) {
	remotes, err := t.RemoteSource.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("Syncs: %w", err)
	}

	var res []Syncer

	for _, rem := range remotes {
		r := Remote(rem)

		res = append(res, Sync{
			r,
			r.local(t.LocalPrefix),
		})
	}

	return res, nil
}
