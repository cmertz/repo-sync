package source

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// githubPageSize is the most search results the GitHub GraphQL API returns
// per page.
const githubPageSize = 100

// Github is a GitHub source for remotes to sync from.
type Github struct {
	Username string
	Token    string
}

// githubSearch is one page of the repositories a GitHub search finds.
type githubSearch struct {
	Search struct {
		PageInfo struct {
			EndCursor   githubv4.String
			HasNextPage githubv4.Boolean
		}
		Nodes []struct {
			Repository struct {
				SSHURL githubv4.String `graphql:"sshUrl"`
			} `graphql:"... on Repository"`
		}
	} `graphql:"search(query: $query, type: REPOSITORY, first: $first, after: $cursor)"`
}

// List returns the SSH URLs of the repositories that the user owns and the
// token can see. It uses the GitHub search, which leaves out forks and finds
// at most 1,000 repositories.
func (g Github) List(ctx context.Context) ([]string, error) {
	client := githubv4.NewClient(
		oauth2.NewClient(
			ctx,
			oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: g.Token,
			}),
		),
	)

	vars := map[string]any{
		"query":  githubv4.String("user:" + g.Username),
		"first":  githubv4.Int(githubPageSize),
		"cursor": (*githubv4.String)(nil),
	}

	var urls []string

	for {
		var page githubSearch

		err := client.Query(ctx, &page, vars)
		if err != nil {
			return nil, fmt.Errorf("github user %s: searching repositories: %w", g.Username, err)
		}

		for _, n := range page.Search.Nodes {
			urls = append(urls, string(n.Repository.SSHURL))
		}

		if !page.Search.PageInfo.HasNextPage {
			return urls, nil
		}

		vars["cursor"] = githubv4.NewString(page.Search.PageInfo.EndCursor)
	}
}
