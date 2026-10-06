//nolint:testpackage // tests an unexported method
package sync

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestRemote_local(t *testing.T) {
	cases := []struct {
		remote string
		prefix string
		want   string
	}{
		{"example.com", "", "example.com"},
		{"example.com", "src", "src/example.com"},
		{"git@example.com:user/1", "src", "src/example.com/user/1"},
		{"ssh://example.com/user/2", "src", "src/example.com/user/2"},
		{"git://example.com/user/3", "src", "src/example.com/user/3"},
		{"https://example.com/user/4", "src", "src/example.com/user/4"},
		{"http://example.com/user/5", "src", "src/example.com/user/5"},
		{"ssh://git@example.com:2222/user/6.git", "src", "src/example.com/user/6"},
		{"https://example.com/user/7.git", "/src", "/src/example.com/user/7"},
		{"git@example.com:user/8.git", "", "example.com/user/8"},
		{"example.com:user/9", "src", "src/example.com/user/9"},
		{"file:///srv/git/10.git", "src", "src/srv/git/10"},
		{"./dir:with/colon", "src", "src/dir:with/colon"},
	}

	for _, c := range cases {
		t.Run(c.remote, func(t *testing.T) {
			got, err := Remote(c.remote).local(c.prefix)
			if err != nil {
				t.Fatalf("local(%q) error = %v", c.prefix, err)
			}

			want := Local(filepath.FromSlash(c.want))
			if got != want {
				t.Errorf("local(%q) = %q, want %q", c.prefix, got, want)
			}
		})
	}
}

func TestRemote_local_invalid(t *testing.T) {
	for _, remote := range []string{
		"",
		"git@example.com:../../etc",
		"https://example.com/../../etc",
		"ssh://example.com/%zz",
	} {
		t.Run(remote, func(t *testing.T) {
			got, err := Remote(remote).local("src")
			if !errors.Is(err, ErrInvalidRemote) {
				t.Errorf("local() = %q, %v, want error %v", got, err, ErrInvalidRemote)
			}
		})
	}
}
