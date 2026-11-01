package source_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/cmertz/repo-sync/source"
)

func TestGitea_List(t *testing.T) {
	t.Run("follows the pages", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/version", func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{"version":"1.22.0"}`)
		})

		var srv *httptest.Server

		mux.HandleFunc("GET /api/v1/user/repos", func(w http.ResponseWriter, r *http.Request) {
			got := r.Header.Get("Authorization")
			if got != "token secret" {
				t.Errorf("Authorization = %q, want %q", got, "token secret")
			}

			switch r.URL.Query().Get("page") {
			case "1":
				w.Header().Set("Link", fmt.Sprintf(`<%s/api/v1/user/repos?page=2&limit=50>; rel="next"`, srv.URL))
				fmt.Fprint(w, `[{"ssh_url":"git@gitea.example.com:me/a.git"}]`)
			case "2":
				fmt.Fprint(w, `[{"ssh_url":"git@gitea.example.com:me/b.git"}]`)
			default:
				t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
			}
		})

		srv = httptest.NewServer(mux)
		t.Cleanup(srv.Close)

		got, err := source.Gitea{URL: srv.URL, Token: "secret"}.List(t.Context())
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		want := []string{"git@gitea.example.com:me/a.git", "git@gitea.example.com:me/b.git"}
		if !slices.Equal(got, want) {
			t.Errorf("List() = %q, want %q", got, want)
		}
	})

	t.Run("reports an error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, `{"message":"token is required"}`, http.StatusUnauthorized)
		}))
		t.Cleanup(srv.Close)

		_, err := source.Gitea{URL: srv.URL, Token: "secret"}.List(t.Context())
		if err == nil {
			t.Error("List() error = nil, want an error")
		}
	})
}
