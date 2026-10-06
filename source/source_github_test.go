package source_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"

	"golang.org/x/oauth2"

	"github.com/cmertz/repo-sync/source"
)

// redirectTransport sends every request to target, so that a test server can
// stand in for api.github.com.
type redirectTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (rt redirectTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.URL.Scheme = rt.target.Scheme
	r.URL.Host = rt.target.Host

	return rt.base.RoundTrip(r) //nolint:wrapcheck // a transport passes errors through
}

// withServer returns a context that makes the oauth2 client, and with it
// source.Github, talk to srv.
func withServer(t *testing.T, srv *httptest.Server) context.Context {
	t.Helper()

	target, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{Transport: redirectTransport{target: target, base: srv.Client().Transport}}

	return context.WithValue(t.Context(), oauth2.HTTPClient, client)
}

func TestGithub_List(t *testing.T) {
	t.Run("follows the pages", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := r.Header.Get("Authorization")
			if got != "Bearer secret" {
				t.Errorf("Authorization = %q, want %q", got, "Bearer secret")
			}

			var req struct {
				Variables struct {
					Query  string  `json:"query"`
					Cursor *string `json:"cursor"`
				} `json:"variables"`
			}

			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				t.Errorf("decoding request: %v", err)
			}

			if req.Variables.Query != "user:me" {
				t.Errorf("query = %q, want %q", req.Variables.Query, "user:me")
			}

			w.Header().Set("Content-Type", "application/json")

			switch {
			case req.Variables.Cursor == nil:
				fmt.Fprint(w, `{"data":{"search":{
					"pageInfo":{"endCursor":"c1","hasNextPage":true},
					"nodes":[{"sshUrl":"git@github.com:me/a.git"}]}}}`)
			case *req.Variables.Cursor == "c1":
				fmt.Fprint(w, `{"data":{"search":{
					"pageInfo":{"endCursor":"c2","hasNextPage":false},
					"nodes":[{"sshUrl":"git@github.com:me/b.git"}]}}}`)
			default:
				t.Errorf("unexpected cursor %q", *req.Variables.Cursor)
			}
		}))
		t.Cleanup(srv.Close)

		got, err := source.Github{Username: "me", Token: "secret"}.List(withServer(t, srv))
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		want := []string{"git@github.com:me/a.git", "git@github.com:me/b.git"}
		if !slices.Equal(got, want) {
			t.Errorf("List() = %q, want %q", got, want)
		}
	})

	t.Run("reports an error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, `{"message":"Bad credentials"}`, http.StatusUnauthorized)
		}))
		t.Cleanup(srv.Close)

		_, err := source.Github{Username: "me", Token: "secret"}.List(withServer(t, srv))
		if err == nil {
			t.Error("List() error = nil, want an error")
		}
	})
}
