package source

import (
	"context"
	"net/http"
	"time"

	"github.com/google/go-github/v62/github"
)

type githubSource struct {
	client *github.Client
}

func newGithubClient(config GithubConfig) (Client, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	c := github.NewClient(httpClient).WithAuthToken(config.AuthToken)
	return &githubSource{client: c}, nil
}

// https://pkg.go.dev/github.com/google/go-github/v62/github#RepositoriesService.ListForks

func (g *githubSource) GetRepository(ctx context.Context, owner, name string) (Repository, error) {
	return Repository{}, nil
}
func (g *githubSource) GetForks(ctx context.Context, owner, name string) ([]Repository, error) {
	return nil, nil
}

func (g *githubSource) ListBranches(ctx context.Context, repo Repository) ([]Branch, error) {
	return nil, nil
}

func (g *githubSource) ListCommits(ctx context.Context, repo Repository, branch Branch) ([]Commit, error) {
	return nil, nil
}
