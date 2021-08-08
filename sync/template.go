package sync

import "context"

// Template is a template for generating `Sync`s
// TODO: find a better name
type Template struct {
	RemoteSource interface {
		List(ctx context.Context) (remoteURLs []string, err error)
	}
	LocalPrefix string
}

func (t Template) Syncs(ctx context.Context) ([]Sync, error) {
	remotes, err := t.RemoteSource.List(ctx)
	if err != nil {
		return nil, err
	}

	var res []Sync
	for _, rem := range remotes {
		r := remote(rem)

		res = append(res, Sync{
			r,
			r.local(t.LocalPrefix),
		})
	}

	return res, nil
}
