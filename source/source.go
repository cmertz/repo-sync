// package source provides functionality to
// list accessible remote repositories for
// syncing
package source

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	sourceKindGitlab = "gitlab"
	sourceKindGithub = "github"
	sourceKindGitea  = "gitea"

	KeyToken    = "token"
	KeyUsername = "username"
	KeyURL      = "url"
)

var (
	ErrUnknownSourceKind = errors.New("unknown source kind")
	ErrMissingField      = errors.New("missing field")
)

// Source for remotes to sync from
type Source interface {
	List(ctx context.Context) (remoteURLs []string, err error)
}

// New creates a new Source from the provided values
func New(kind string, values map[string]string) (Source, error) {

	// TODO: replace this mess via reflection
	switch kind {
	case sourceKindGitlab:
		err := checkFieldsSet(values, KeyURL, KeyToken)
		if err != nil {
			return nil, err
		}

		return Gitlab{
			URL:   values[KeyURL],
			Token: values[KeyToken],
		}, nil

	case sourceKindGithub:
		err := checkFieldsSet(values, KeyUsername, KeyToken)
		if err != nil {
			return nil, err
		}

		return Github{
			Username: values[KeyUsername],
			Token:    values[KeyToken],
		}, nil

	case sourceKindGitea:
		err := checkFieldsSet(values, KeyURL, KeyToken)
		if err != nil {
			return nil, err
		}

		return Gitea{
			URL:   values[KeyURL],
			Token: values[KeyToken],
		}, nil

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownSourceKind, kind)
	}
}

func checkFieldsSet(values map[string]string, keys ...string) error {
	for _, k := range keys {
		v, ok := values[k]
		if !ok || len(strings.TrimSpace(v)) == 0 {
			return fmt.Errorf("%w: %s", ErrMissingField, k)
		}
	}

	return nil
}
