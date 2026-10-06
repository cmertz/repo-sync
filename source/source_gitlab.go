package source

import (
	"context"
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
)

// gitlabPageSize is the most projects the GitLab API returns per page.
const gitlabPageSize = 100

// Gitlab is a GitLab source for remotes to sync from.
type Gitlab struct {
	URL   string
	Token string
}

// List returns the SSH URLs of all private projects the owner of the token
// can access.
func (g Gitlab) List(ctx context.Context) ([]string, error) {
	client, err := gitlab.NewClient(
		g.Token,
		gitlab.WithBaseURL(g.URL),
	)
	if err != nil {
		return nil, fmt.Errorf("gitlab %s: %w", g.URL, err)
	}

	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: gitlabPageSize,
		},
		Visibility: gitlab.Ptr(gitlab.PrivateVisibility),
	}

	var urls []string

	for {
		projects, resp, err := client.Projects.ListProjects(opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("gitlab %s: listing projects: %w", g.URL, err)
		}

		for _, p := range projects {
			urls = append(urls, p.SSHURLToRepo)
		}

		// Stop at the last page, and at a next page that doesn't advance.
		if resp.NextPage <= opt.Page {
			return urls, nil
		}

		opt.Page = resp.NextPage
	}
}
