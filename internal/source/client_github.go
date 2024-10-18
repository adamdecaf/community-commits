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

	var out []Repository
	for _, repo := range repos {
		// Skip forks that aren't usable
		if repo.GetArchived() || repo.GetDisabled() || repo.GetPrivate() {
			continue
		}

		out = append(out, Repository{
			Owner: repo.GetOwner().GetLogin(),
			Name:  repo.GetName(),
		})
	}
	return out, nil
}

func (g *githubSource) ListBranches(ctx context.Context, repo Repository) ([]Branch, error) {
	opts := &github.BranchListOptions{
		ListOptions: github.ListOptions{ // TODO(adam): paginate
			Page:    0,
			PerPage: 100,
		},
	}
	branches, resp, err := g.client.Repositories.ListBranches(ctx, repo.Owner, repo.Name, opts)
	if err != nil {
		return nil, fmt.Errorf("listing branches: %w", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	var out []Branch
	for _, branch := range branches {
		b := Branch{
			Name:       branch.GetName(),
			Repository: repo,
		}

		authoredAt := branch.GetCommit().GetCommit().GetAuthor().GetDate()
		committedAt := branch.GetCommit().GetCommit().GetCommitter().GetDate()
		if authoredAt.GetTime() != nil {
			if !authoredAt.Time.IsZero() {
				b.LastCommitTimestamp = authoredAt.Time
			}
		}
		if committedAt.GetTime() != nil {
			if !committedAt.Time.IsZero() {
				if committedAt.Time.After(b.LastCommitTimestamp) {
					b.LastCommitTimestamp = committedAt.Time
				}
			}
		}

		out = append(out, b)
	}
	return out, nil
}

func (g *githubSource) ListCommits(ctx context.Context, repo Repository, branch Branch) ([]Commit, error) {
	return nil, nil
}
