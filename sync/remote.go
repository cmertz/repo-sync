package sync

import "strings"

type remote string

func (r remote) url() string {
	return string(r)
}

func (r remote) basepath() string {
	path := strings.TrimSuffix(r.url(), ".git")

	for _, prefix := range []string{"git@", "git://", "ssh://", "https://", "http://"} {
		if strings.HasPrefix(path, prefix) {
			path = strings.TrimPrefix(path, prefix)
			break
		}
	}

	return strings.ReplaceAll(path, ":", "/")
}
