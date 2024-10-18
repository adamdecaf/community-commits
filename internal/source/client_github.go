package source

import (
	"context"
	"fmt"
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

func (g *githubSource) GetRepository(ctx context.Context, owner, name string) (Repository, error) {
	return Repository{}, nil
}

func (g *githubSource) GetForks(ctx context.Context, owner, name string) ([]Repository, error) {
	opts := &github.RepositoryListForksOptions{
		ListOptions: github.ListOptions{ // TODO(adam): paginate
			Page:    0,
			PerPage: 100,
		},
	}
	repos, resp, err := g.client.Repositories.ListForks(ctx, owner, name, opts)
	if err != nil {
		return nil, fmt.Errorf("listing forks: %w", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	out := make([]Repository, len(repos))
	for i := range repos {
		// Skip forks that aren't usable
		if repos[i].GetArchived() || repos[i].GetDisabled() || repos[i].GetPrivate() {
			continue
		}

		out[i] = Repository{
			Owner: repos[i].GetOwner().GetLogin(),
			Name:  repos[i].GetName(),
		}
	}
	return out, nil
}

func (g *githubSource) ListBranches(ctx context.Context, repo Repository) ([]Branch, error) {
	return nil, nil
}

func (g *githubSource) ListCommits(ctx context.Context, repo Repository, branch Branch) ([]Commit, error) {
	return nil, nil
}
