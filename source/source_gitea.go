package source

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
)

const defaultGiteaMaxPageSize = 50

// Gitea is a gitea source for remotes to sync from
type Gitea struct {
	URL   string
	Token string
}

// List remote repository urls
func (g Gitea) List(ctx context.Context) ([]string, error) {
	client, err := gitea.NewClient(
		g.URL,
		gitea.SetContext(ctx),
		gitea.SetToken(g.Token),
	)
	if err != nil {
		return nil, fmt.Errorf("List: %w", err)
	}

	// TODO: pagination
	// TODO: check if `ListMyRepos` is to broad
	repos, resp, err := client.ListMyRepos(gitea.ListReposOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: defaultGiteaMaxPageSize,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("List: %w", err)
	}
	defer resp.Body.Close()

	var res []string
	for _, r := range repos {
		res = append(res, r.SSHURL)
	}

	return res, nil
}
