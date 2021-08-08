package source

import (
	"context"
	"fmt"

	"github.com/xanzy/go-gitlab"
)

// Gitlab is a gitlab source for remotes to sync from
type Gitlab struct {
	URL   string
	Token string
}

// List remote repository urls
func (g Gitlab) List(ctx context.Context) ([]string, error) {
	client, err := gitlab.NewClient(
		g.Token,
		gitlab.WithBaseURL(g.URL),
	)
	if err != nil {
		return nil, fmt.Errorf("List: %w", err)
	}

	// TODO: pagination
	repos, resp, err := client.Projects.ListProjects(
		&gitlab.ListProjectsOptions{
			Visibility: gitlab.Visibility(gitlab.PrivateVisibility),
		},
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("List: %w", err)
	}
	defer resp.Body.Close()

	var res []string
	for _, r := range repos {
		res = append(res, r.SSHURLToRepo)
	}

	return res, nil
}
