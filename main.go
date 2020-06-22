package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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
	kindGitlab  = "gitlab"
	kindGitea   = "gitea"
	kindGithub  = "github"
	kindGitRepo = "git"
)

type repo struct {
	remote string
	path   string
}

func (r repo) sync(ctx context.Context) error {
	parts := strings.Split(r.path, "/")
	parentDir := strings.Join(parts[:len(parts)-1], "/")

	err := os.MkdirAll(parentDir, 0700)
	if err != nil {
		return err
	}

	repo, err := git.PlainOpen(r.path)
	if err != nil {
		repo, err = git.PlainCloneContext(ctx, r.path, false, &git.CloneOptions{
			URL: r.remote,
		})

		// ignore empty remotes
		if err != nil && err != transport.ErrEmptyRemoteRepository {
			return err
		}
	}

	var w *git.Worktree
	w, err = repo.Worktree()
	if err != nil {
		return err
	}

	err = w.PullContext(ctx, &git.PullOptions{RemoteName: "origin"})

	// ignore if there's nothing to update
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}

type repoSource struct {
	Kind         string
	URL          string
	Username     string
	TokenCommand string `yaml:"tokenCommand"`

	// either one of those needs to be provided
	Prefix string
	Path   string
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

func (r repoSource) repos(ctx context.Context) ([]repo, error) {
	token, err := r.token(ctx)
	if err != nil {
		return nil, err
	}

	switch r.Kind {
	case kindGitlab:
		return reposGitlab(ctx, r.URL, token, r.Prefix)
	case kindGithub:
		return reposGithub(ctx, token, r.Username, r.Prefix)
	case kindGitea:
		return reposGitea(r.URL, token, r.Username, r.Prefix)
	case kindGitRepo:
		return []repo{repo{remote: r.URL, path: r.Path}}, nil
	default:
		return []repo{}, fmt.Errorf("unknown provider kind '%s'", r.Kind)
	}
}

func path(prefix, url string) string {
	if strings.HasPrefix(url, "git@") {
		url = strings.TrimPrefix(url, "git@")
	}

	if strings.HasSuffix(url, ".git") {
		url = strings.TrimSuffix(url, ".git")
	}

	return prefix + "/" + strings.Replace(url, ":", "/", -1)
}

func reposGitea(url, token, username, prefix string) ([]repo, error) {
	c := gitea.NewClient(url, token)

	repos, err := c.ListMyRepos()
	if err != nil {
		return nil, err
	}

	var res []repo
	for _, r := range repos {
		res = append(res, repo{
			remote: r.SSHURL,
			path:   path(prefix, r.SSHURL),
		})
	}
	return res, nil
}

func reposGithub(ctx context.Context, token, username, prefix string) ([]repo, error) {
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

	var res []repo
	for _, edge := range query.Search.Edges {
		repoURL := string(edge.Node.Repository.SSHURL)
		res = append(res, repo{
			remote: repoURL,
			path:   path(prefix, repoURL),
		})
	}

	return res, nil
}

func reposGitlab(ctx context.Context, url, token, prefix string) ([]repo, error) {
	c := gitlab.NewClient(nil, token)
	c.SetBaseURL(url)

	repos, _, err := c.Projects.ListProjects(&gitlab.ListProjectsOptions{Visibility: gitlab.Visibility(gitlab.PrivateVisibility)}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	var res []repo
	for _, r := range repos {
		res = append(res, repo{
			remote: r.SSHURLToRepo,
			path:   path(prefix, r.SSHURLToRepo),
		})
	}

	return res, nil
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	confPath := home + "/.config/reposync.yaml"

	flag.StringVar(&confPath, "config", confPath, "TODO")
	flag.Parse()

	var b []byte
	b, err = ioutil.ReadFile(confPath)
	if err != nil {
		log.Fatalln(err)
	}

	var sources []repoSource
	err = yaml.Unmarshal(b, &sources)
	if err != nil {
		log.Fatalln(err)
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

	for _, s := range sources {
		repos, err := s.repos(ctx)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, repo := range repos {
			err = repo.sync(ctx)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}
