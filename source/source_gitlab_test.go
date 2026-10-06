package source_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/cmertz/repo-sync/source"
)

func TestGitlab_List(t *testing.T) {
	t.Run("follows the pages", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v4/projects" {
				t.Errorf("unexpected path %q", r.URL.Path)
			}

			token := r.Header.Get("PRIVATE-TOKEN")
			if token != "secret" {
				t.Errorf("PRIVATE-TOKEN = %q, want %q", token, "secret")
			}

			visibility := r.URL.Query().Get("visibility")
			if visibility != "private" {
				t.Errorf("visibility = %q, want %q", visibility, "private")
			}

			w.Header().Set("Content-Type", "application/json")

			switch r.URL.Query().Get("page") {
			case "1":
				w.Header().Set("X-Next-Page", "2")
				fmt.Fprint(w, `[{"ssh_url_to_repo":"git@gitlab.example.com:me/a.git"}]`)
			case "2":
				fmt.Fprint(w, `[{"ssh_url_to_repo":"git@gitlab.example.com:me/b.git"}]`)
			default:
				t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
			}
		}))
		t.Cleanup(srv.Close)

		got, err := source.Gitlab{URL: srv.URL, Token: "secret"}.List(t.Context())
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		want := []string{"git@gitlab.example.com:me/a.git", "git@gitlab.example.com:me/b.git"}
		if !slices.Equal(got, want) {
			t.Errorf("List() = %q, want %q", got, want)
		}
	})

	t.Run("reports an error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, `{"message":"401 Unauthorized"}`, http.StatusUnauthorized)
		}))
		t.Cleanup(srv.Close)

		_, err := source.Gitlab{URL: srv.URL, Token: "secret"}.List(t.Context())
		if err == nil {
			t.Error("List() error = nil, want an error")
		}
	})
}
