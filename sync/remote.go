package sync

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// ErrInvalidRemote reports a remote URL that maps to no path below the local
// prefix.
var ErrInvalidRemote = errors.New("invalid remote")

// Remote is the URL of a remote git repository.
type Remote string

func (r Remote) url() string {
	return string(r)
}

// local returns the path of the working copy of r below prefix: the host and
// path of the URL, without scheme, user, port, and ".git" suffix. It rejects
// a URL whose path would escape prefix.
func (r Remote) local(prefix string) (Local, error) {
	host, path, err := r.hostAndPath()
	if err != nil {
		return "", fmt.Errorf("%w %q: %w", ErrInvalidRemote, r, err)
	}

	path = strings.TrimSuffix(strings.TrimLeft(path, "/"), ".git")

	rel := filepath.Join(host, filepath.FromSlash(path))
	if rel == "." || !filepath.IsLocal(rel) {
		return "", fmt.Errorf("%w %q: no path below the prefix", ErrInvalidRemote, r)
	}

	return Local(filepath.Join(prefix, rel)), nil
}

// hostAndPath splits r into host and path. It understands URLs with a scheme
// ("ssh://git@example.com:2222/user/repo.git"), the scp-like syntax of SSH
// ("git@example.com:user/repo.git"), and plain paths.
func (r Remote) hostAndPath() (host, path string, err error) {
	s := r.url()

	if strings.Contains(s, "://") {
		u, err := url.Parse(s)
		if err != nil {
			return "", "", fmt.Errorf("parsing url: %w", err)
		}

		return u.Hostname(), u.Path, nil
	}

	// As with git, a colon before the first slash makes it scp-like.
	userHost, path, found := strings.Cut(s, ":")
	if found && !strings.Contains(userHost, "/") {
		// Drop the user, if any; LastIndex is -1 without one.
		host = userHost[strings.LastIndex(userHost, "@")+1:]

		return host, path, nil
	}

	return "", s, nil
}
