package source

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// Github is a github source for remotes to sync from
type Github struct {
	Username string
	Token    string
}

// List remote repository urls
func (g Github) List(ctx context.Context) ([]string, error) {
	client := githubv4.NewClient(
		oauth2.NewClient(
			ctx,
			oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: g.Token,
			}),
		),
	)

	// TODO: pagination
	var query struct {
		Search struct {
			RepositoryCount githubv4.Int
			PageInfo        struct {
				EndCursor   githubv4.String
				StartCursor githubv4.String
			}
			Edges []struct {
				Node struct {
					Repository struct {
						SSHURL githubv4.String
					} `graphql:"... on Repository"`
				}
			}
		} `graphql:"search(query: $query, type: REPOSITORY, first: 100)"`
	}

	err := client.Query(ctx, &query, map[string]interface{}{"query": githubv4.String("user:" + g.Username)})
	if err != nil {
		return nil, fmt.Errorf("List: %w", err)
	}

	var res []string
	for _, edge := range query.Search.Edges {
		res = append(res, string(edge.Node.Repository.SSHURL))
	}

	return res, nil
}
