package source

import (
	"bytes"
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
			rawPayload := event.GetRawPayload()

			if bytes.Contains(rawPayload, []byte("renovate[bot]")) || bytes.Contains(rawPayload, []byte("dependabot[bot]")) {
				return
			}
			if bytes.Contains(rawPayload, []byte("snyk.io")) {
				return
			}
			if bytes.Contains(rawPayload, []byte("github@users.noreply.github.com")) {
				return
			}

			// TODO(adam): read from config
			if bytes.Contains(rawPayload, []byte("Adam Shannon")) {
				return
			}

			payload, err := event.ParsePayload()
			if err != nil {
				return
			}
			pushEvent, ok := payload.(*github.PushEvent)
			if !ok {
				return
			}

			if len(pushEvent.Commits) > 0 {
				pushEvents = append(pushEvents, PushEvent{
					RepoSlug:  makeRepoSlug(pushEvent.Commits[0]),
					Commits:   makeWebCommits(pushEvent.Commits),
					CreatedAt: event.GetCreatedAt().Time,
				})
			}
		}
	}

	currentListOptions := &github.ListOptions{}
	maxListOptions := &github.ListOptions{
		Page:    5,
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

func makeWebCommits(commits []*github.HeadCommit) []WebCommit {
	var out []WebCommit
	for _, commit := range commits {
		// convert https://api.github.com/repos/andrewjones-bond/ach/commits/f2b925a4f758140544cf1ccecbf4bf113a9123b0
		// into    https://    github.com/      andrewjones-bond/ach/commits/f2b925a4f758140544cf1ccecbf4bf113a9123b0
		url := strings.TrimPrefix(*commit.URL, "https://api.github.com/repos/")
		url = fmt.Sprintf("https://github.com/%s", url)

		out = append(out, WebCommit{
			CommitURL: url,
			Message:   *commit.Message,
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
	if currentListOptions.Page > maxListOptions.Page && currentListOptions.PerPage > maxListOptions.PerPage {
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
