package source

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"
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

func (g *githubSource) ListNetworkPushEvents(ctx context.Context, repo Repository) ([]PushEvent, error) {
	var pushEvents []PushEvent

	collector := func(event *github.Event) {
		if strings.EqualFold("PushEvent", event.GetType()) {
			payload, err := event.ParsePayload()
			if err != nil {
				return
			}
			pushEvent, ok := payload.(*github.PushEvent)
			if !ok {
				return
			}

			parts := strings.Split(event.Repo.GetName(), "/")
			if len(parts) != 2 {
				fmt.Printf("ERROR: unexpected org/repo name: %v", event.Repo.GetName())
			}

			owner, name := parts[0], parts[1]

			comparison, _, err := g.client.Repositories.CompareCommits(ctx, owner, name, *pushEvent.Before, *pushEvent.Head, nil)
			if err != nil {
				fmt.Printf("CompareCommits error for %s/%s (before=%s, head=%s): %v\n", owner, name, *pushEvent.Before, *pushEvent.Head, err)
				return
			}

			if len(comparison.Commits) > 0 {
				repoSlug := fmt.Sprintf("%s/%s", owner, name)
				commits := makeWebCommitsFromComparison(comparison.Commits)

				if len(commits) > 0 {
					evt := PushEvent{
						RepoSlug:  repoSlug,
						Commits:   commits,
						CreatedAt: event.GetCreatedAt().Time,
					}
					pushEvents = append(pushEvents, evt)
				}
			}
		}
	}

	currentListOptions := &github.ListOptions{
		Page: -1,
	}
	maxListOptions := &github.ListOptions{
		Page:    3,
		PerPage: 100,
	}
	err := g.listNetworkEvents(ctx, repo, collector, currentListOptions, maxListOptions)
	if err != nil {
		return nil, fmt.Errorf("listing network commits: %w", err)
	}
	return pushEvents, nil
}

func makeRepoSlug(commit *github.HeadCommit) string {
	// convert https://api.github.com/repos/andrewjones-bond/ach/commits/f2b925a4f758140544cf1ccecbf4bf113a9123b0
	parts := strings.Split(strings.TrimPrefix(*commit.URL, "https://"), "/")
	return fmt.Sprintf("%s/%s", parts[2], parts[3])
}

var (
	skippableMessageSlugs = []string{
		"renovate[bot]",
		"dependabot[bot]",
		"snyk.io",
		"github@users.noreply.github.com",
		"Adam Shannon",
	}
)

func makeWebCommitsFromComparison(commits []*github.RepositoryCommit) []WebCommit {
	var out []WebCommit
	for _, commit := range commits {
		if commit.Commit.Message == nil {
			continue
		}

		message := *commit.Commit.Message
		for idx := range skippableMessageSlugs {
			if strings.Contains(message, skippableMessageSlugs[idx]) {
				continue
			}
		}

		// Transform API URL to web URL (same as before)
		url := strings.TrimPrefix(*commit.URL, "https://api.github.com/repos/")
		url = fmt.Sprintf("https://github.com/%s", url)

		out = append(out, WebCommit{
			CommitURL: url,
			Message:   message,
		})
	}
	return out
}

func (g *githubSource) listNetworkEvents(
	ctx context.Context,
	repo Repository,
	collector func(event *github.Event),
	currentListOptions *github.ListOptions,
	maxListOptions *github.ListOptions,
) error {
	// Have we reached our max?
	if currentListOptions.Page >= maxListOptions.Page && currentListOptions.PerPage >= maxListOptions.PerPage {
		return nil
	}
	// Increment
	if currentListOptions.Page < maxListOptions.Page {
		currentListOptions.Page += 1
	}
	if currentListOptions.PerPage < maxListOptions.PerPage {
		currentListOptions.PerPage = maxListOptions.PerPage
	}

	// Pull evenets
	events, resp, err := g.client.Activity.ListEventsForRepoNetwork(ctx, repo.Owner, repo.Name, currentListOptions)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return nil
		}
		return fmt.Errorf("listing commits: %w", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	for _, event := range events {
		collector(event)
	}

	return g.listNetworkEvents(ctx, repo, collector, currentListOptions, maxListOptions)
}
