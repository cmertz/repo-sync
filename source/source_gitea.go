package source

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
)

// giteaPageSize is the number of repositories to request per page. Gitea
// returns at most its MAX_RESPONSE_ITEMS setting, 50 by default, and the
// Link header of the response points to the next page either way.
const giteaPageSize = 50

// Gitea is a Gitea source for remotes to sync from.
type Gitea struct {
	URL   string
	Token string
}

// List returns the SSH URLs of the repositories that Gitea lists for the
// owner of the token at /user/repos.
func (g Gitea) List(ctx context.Context) ([]string, error) {
	client, err := gitea.NewClient(
		g.URL,
		gitea.SetContext(ctx),
		gitea.SetToken(g.Token),
	)
	if err != nil {
		return nil, fmt.Errorf("gitea %s: %w", g.URL, err)
	}

	opt := gitea.ListReposOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: giteaPageSize,
		},
	}

	var urls []string

	for {
		repos, resp, err := client.ListMyRepos(opt)
		if err != nil {
			return nil, fmt.Errorf("gitea %s: listing repositories: %w", g.URL, err)
		}

		for _, r := range repos {
			urls = append(urls, r.SSHURL)
		}

		// Stop at the last page, and at a next page that doesn't advance.
		if resp.NextPage <= opt.Page {
			return urls, nil
		}

		opt.Page = resp.NextPage
	}
}
