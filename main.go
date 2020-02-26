package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"code.gitea.io/sdk/gitea"
	"github.com/shurcooL/githubv4"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/yaml.v2"
)

const (
	kindGitlab = "gitlab"
	kindGitea  = "gitea"
	kindGithub = "github"
)

type repoURL string

func (r repoURL) path(prefix string) string {
	u := string(r)
	if strings.HasPrefix(u, "git@") {
		u = strings.TrimPrefix(u, "git@")
	}

	if strings.HasSuffix(u, ".git") {
		u = strings.TrimSuffix(u, ".git")
	}

	return prefix + "/" + strings.Replace(u, ":", "/", -1)
}

func (r repoURL) ensureParentDirsExist(prefix string) {
	parts := strings.Split(r.path(prefix), "/")
	parentDir := strings.Join(parts[:len(parts)-1], "/")

	if err := os.MkdirAll(parentDir, 0700); err != nil {
		panic(err)
	}
}

func (r repoURL) clone(ctx context.Context, prefix string) error {
	r.ensureParentDirsExist(prefix)

	_, err := git.PlainCloneContext(ctx, r.path(prefix), false, &git.CloneOptions{
		URL: string(r),
	})

	if err == transport.ErrEmptyRemoteRepository {
		return nil
	}

	return err
}

func (r repoURL) pull(ctx context.Context, prefix string) error {
	gitRepo, err := git.PlainOpen(r.path(prefix))
	if err != nil {
		return err
	}

	var w *git.Worktree
	w, err = gitRepo.Worktree()
	if err != nil {
		return err
	}

	err = w.PullContext(ctx, &git.PullOptions{RemoteName: "origin"})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (r repoURL) hasLocalCopy(prefix string) bool {
	_, err := git.PlainOpen(r.path(prefix))
	return err == nil
}

func (r repoURL) sync(ctx context.Context, prefix string) error {
	if !r.hasLocalCopy(prefix) {
		return r.clone(ctx, prefix)
	}

	return r.pull(ctx, prefix)
}

type repoSource struct {
	Kind         string
	URL          string
	Username     string
	TokenCommand string `yaml:"tokenCommand"`
}

func (r repoSource) token(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", r.TokenCommand)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out.String()), nil
}

func (r repoSource) repos(ctx context.Context) ([]repoURL, error) {
	token, err := r.token(ctx)
	if err != nil {
		return nil, err
	}

	switch r.Kind {
	case kindGitlab:
		return reposGitlab(ctx, r.URL, token)
	case kindGithub:
		return reposGithub(ctx, token, r.Username)
	case kindGitea:
		return reposGitea(r.URL, token, r.Username)
	default:
		panic(fmt.Sprintf("unknown provider kind '%s'", r.Kind))
	}
}

func reposGitea(url, token, username string) ([]repoURL, error) {
	c := gitea.NewClient(url, token)

	repos, err := c.ListMyRepos()
	if err != nil {
		return nil, err
	}

	var res []repoURL
	for _, repo := range repos {
		res = append(res, repoURL(repo.SSHURL))
	}
	return res, nil
}

func reposGithub(ctx context.Context, token, username string) ([]repoURL, error) {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
	client := githubv4.NewClient(httpClient)

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

	err := client.Query(ctx, &query, map[string]interface{}{"query": githubv4.String("user:" + username)})
	if err != nil {
		return nil, err
	}

	var res []repoURL
	for _, edge := range query.Search.Edges {
		res = append(res, repoURL(edge.Node.Repository.SSHURL))
	}

	return res, nil
}

func reposGitlab(ctx context.Context, url, token string) ([]repoURL, error) {
	c := gitlab.NewClient(nil, token)
	c.SetBaseURL(url)

	repos, _, err := c.Projects.ListProjects(&gitlab.ListProjectsOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	var res []repoURL
	for _, repo := range repos {
		res = append(res, repoURL(repo.SSHURLToRepo))
	}

	return res, nil
}

type repoSources struct {
	Sources []repoSource
	Prefix  string
}

func (r repoSources) sync(ctx context.Context) error {
	for _, s := range r.Sources {
		repos, err := s.repos(ctx)
		if err != nil {
			return err
		}

		for _, repo := range repos {
			err = repo.sync(ctx, r.Prefix)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	var b []byte
	b, err = ioutil.ReadFile(home + "/.reposync.yml")
	if err != nil {
		panic(err)
	}

	var cfg repoSources
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		panic(err)
	}

	var sig = make(chan os.Signal)
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
		case <-sig:
			cancel()
		}
	}()

	err = cfg.sync(ctx)
	if err != nil {
		panic(err)
	}
}
