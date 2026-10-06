package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cmertz/repo-sync/source"
)

func TestParseTemplates(t *testing.T) {
	t.Run("reads every source", func(t *testing.T) {
		tpls, err := parseTemplates([]byte(`
- kind: github
  username: me
  token: t1
  prefix: /src
- kind: gitlab
  url: https://gitlab.example.com
  token: t2
- kind: gitea
  url: https://gitea.example.com
  token: t3
  prefix: src
`))
		if err != nil {
			t.Fatalf("parseTemplates() error = %v", err)
		}

		want := []struct {
			source source.Source
			prefix string
		}{
			{source.Github{Username: "me", Token: "t1"}, "/src"},
			{source.Gitlab{URL: "https://gitlab.example.com", Token: "t2"}, ""},
			{source.Gitea{URL: "https://gitea.example.com", Token: "t3"}, "src"},
		}

		if len(tpls) != len(want) {
			t.Fatalf("parseTemplates() returned %d templates, want %d", len(tpls), len(want))
		}

		for i, w := range want {
			if tpls[i].RemoteSource != w.source || tpls[i].LocalPrefix != w.prefix {
				t.Errorf("template %d = %+v, want source %+v and prefix %q", i, tpls[i], w.source, w.prefix)
			}
		}
	})

	t.Run("rejects an unknown kind", func(t *testing.T) {
		_, err := parseTemplates([]byte("- kind: svn\n"))
		if !errors.Is(err, source.ErrUnknownSourceKind) {
			t.Errorf("parseTemplates() error = %v, want %v", err, source.ErrUnknownSourceKind)
		}
	})

	t.Run("rejects a missing field", func(t *testing.T) {
		_, err := parseTemplates([]byte("- kind: github\n  username: me\n"))
		if !errors.Is(err, source.ErrMissingField) {
			t.Errorf("parseTemplates() error = %v, want %v", err, source.ErrMissingField)
		}
	})

	t.Run("rejects a map", func(t *testing.T) {
		_, err := parseTemplates([]byte("kind: github\n"))
		if err == nil {
			t.Error("parseTemplates() error = nil, want an error")
		}
	})
}

func TestReadTemplates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	err := os.WriteFile(path, []byte("- kind: github\n  username: me\n  token: t\n"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	tpls, err := readTemplates(path)
	if err != nil {
		t.Fatalf("readTemplates() error = %v", err)
	}

	if len(tpls) != 1 {
		t.Errorf("readTemplates() returned %d templates, want 1", len(tpls))
	}

	_, err = readTemplates(filepath.Join(t.TempDir(), "missing.yaml"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("readTemplates() of a missing file: error = %v, want %v", err, os.ErrNotExist)
	}
}
