package sync

import "strings"

// Remote git repository
type Remote string

func (r Remote) url() string {
	return string(r)
}

func (r Remote) local(prefix string) Local {
	p := strings.TrimSuffix(r.url(), ".git")

	for _, prefix := range []string{"git@", "git://", "ssh://", "https://", "http://"} {
		// strip only the first prefix we find, everything else is odd
		if strings.HasPrefix(p, prefix) {
			p = strings.TrimPrefix(p, prefix)

			break
		}
	}

	p = strings.ReplaceAll(p, ":", "/")

	if len(prefix) != 0 {
		p = prefix + "/" + p
	}

	return Local(p)
}
