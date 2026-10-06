// Package source lists the remote repositories to sync from a code hosting
// service.
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

	// KeyToken is the key of the API token in the values passed to New.
	KeyToken = "token"

	// KeyUsername is the key of the GitHub user name in the values passed to New.
	KeyUsername = "username"

	// KeyURL is the key of the server URL in the values passed to New.
	KeyURL = "url"
)

var (
	// ErrUnknownSourceKind reports a kind that New doesn't support.
	ErrUnknownSourceKind = errors.New("unknown source kind")
	// ErrMissingField reports a value that New requires but didn't get.
	ErrMissingField = errors.New("missing field")
)

// Source lists the URLs of remote repositories to sync.
type Source interface {
	List(ctx context.Context) (remoteURLs []string, err error)
}

// New creates a Source of the given kind ("gitea", "github", or "gitlab")
// from values, which hold the fields that kind requires.
func New(kind string, values map[string]string) (Source, error) {
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
		return nil, fmt.Errorf("%w: %q", ErrUnknownSourceKind, kind)
	}
}

func checkFieldsSet(values map[string]string, keys ...string) error {
	for _, k := range keys {
		if strings.TrimSpace(values[k]) == "" {
			return fmt.Errorf("%w: %s", ErrMissingField, k)
		}
	}

	return nil
}
