# repo-sync

`repo-sync` clones and updates local working copies of the repositories you can
access on GitHub, GitLab, and Gitea. Run it on a schedule to keep a backup, or
an up-to-date checkout of every project, below one directory.

For each source in its configuration file, `repo-sync` lists the repositories
through the API of the service, then syncs each one over SSH:

- A repository without a working copy gets cloned.
- A working copy gets a pull, which only fast-forwards.

The working copy of a remote lives at `<prefix>/<host>/<path>`, so
`git@github.com:cmertz/repo-sync.git` with the prefix `/home/cmertz/src` ends
up in `/home/cmertz/src/github.com/cmertz/repo-sync`.

## Install

`repo-sync` needs Go 1.26 or newer:

```sh
go install github.com/cmertz/repo-sync@latest
```

## Configure

The configuration file is a YAML list of sources. Every source has a `kind`, a
`token` for the API, and an optional `prefix`, the directory that holds its
working copies. Without a prefix, the paths start in the current directory, and
`repo-sync` doesn't expand `~` or environment variables.

```yaml
- kind: github
  username: cmertz
  token: ghp_...
  prefix: /home/cmertz/src

- kind: gitlab
  url: https://gitlab.com
  token: glpat-...
  prefix: /home/cmertz/src

- kind: gitea
  url: https://gitea.example.com
  token: 0123abcd...
  prefix: /home/cmertz/src
```

| Kind     | Fields                       | Repositories                                                            |
|----------|------------------------------|-------------------------------------------------------------------------|
| `github` | `username`, `token`          | Owned by `username`, through the search, which leaves out forks and finds at most 1,000 |
| `gitlab` | `url`, `token`               | Private projects that the token can access                              |
| `gitea`  | `url`, `token`               | Those that Gitea lists for the token's user at `/user/repos`            |

Each token needs read access to the repositories it should list; on GitHub, a
classic token lists private repositories only with the `repo` scope. The file
holds the tokens in plain text, so make it readable for you alone:
`chmod 600 config.yaml`.

## Run

```sh
repo-sync -config config.yaml
```

| Flag          | Default            | Meaning                                    |
|---------------|--------------------|--------------------------------------------|
| `-config`     | none, required     | Path to the configuration file             |
| `-concurrent` | `GOMAXPROCS`       | Number of repositories to sync in parallel |

Cloning and pulling use SSH: `repo-sync` takes the keys from your SSH agent
(`SSH_AUTH_SOCK`) and checks hosts against `~/.ssh/known_hosts`.

`repo-sync` logs each failure to standard error and carries on with the other
sources and repositories. It exits with status 1 when anything failed or an
interrupt stopped it, and with status 2 on bad flags. It skips empty remote
repositories, and refuses to clone into a directory that holds files but no
git repository.

## Develop

```sh
go test -race ./...
go test -tags integration ./sync/   # clones this repository from GitHub
golangci-lint run ./...
```

The unit tests of the `sync` package need the `git` binary, through which
go-git serves local repositories.
