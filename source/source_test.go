package source_test

import (
	"errors"
	"testing"

	"github.com/cmertz/repo-sync/source"
)

func TestNew(t *testing.T) {
	cases := []struct {
		name    string
		kind    string
		values  map[string]string
		want    source.Source
		wantErr error
	}{
		{
			name:   "gitea",
			kind:   "gitea",
			values: map[string]string{"url": "https://gitea.example.com", "token": "t"},
			want:   source.Gitea{URL: "https://gitea.example.com", Token: "t"},
		},
		{
			name:   "github",
			kind:   "github",
			values: map[string]string{"username": "me", "token": "t"},
			want:   source.Github{Username: "me", Token: "t"},
		},
		{
			name:   "gitlab",
			kind:   "gitlab",
			values: map[string]string{"url": "https://gitlab.example.com", "token": "t"},
			want:   source.Gitlab{URL: "https://gitlab.example.com", Token: "t"},
		},
		{
			name:    "unknown kind",
			kind:    "svn",
			wantErr: source.ErrUnknownSourceKind,
		},
		{
			name:    "missing field",
			kind:    "gitea",
			values:  map[string]string{"url": "https://gitea.example.com"},
			wantErr: source.ErrMissingField,
		},
		{
			name:    "blank field",
			kind:    "github",
			values:  map[string]string{"username": " ", "token": "t"},
			wantErr: source.ErrMissingField,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := source.New(c.kind, c.values)
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("New() error = %v, want %v", err, c.wantErr)
			}

			if got != c.want {
				t.Errorf("New() = %#v, want %#v", got, c.want)
			}
		})
	}
}
