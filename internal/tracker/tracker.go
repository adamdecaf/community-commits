package tracker

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/adamdecaf/community-commits/internal/source"
)

// TODO(adam):
//
// Grab repository and forks, traverse each branch collecting commits
// dedup commits, but grab new commits, show if they're ahead of mainline,
// can be merged, how big they are, summary, etc
//
// Besides forks, search for commit hashes in mainline across Github, Gitlab, etc
//
// Web UI with button to cherry-pick commit into PR on the mainline.

type Worker struct {
	conf   Config
	logger *slog.Logger
}

func NewWorker(logger *slog.Logger, conf Config) (*Worker, error) {
	w := &Worker{
		conf:   conf,
		logger: logger,
	}

	return w, nil
}

func (w *Worker) Start(ctx context.Context) error {
	events := make(map[string][]source.PushEvent)

	// For each repository grab their latest events and format
	for _, repo := range w.conf.Tracking.Repositories {
		evts, err := w.getLatestNetworkEvents(ctx, repo)
		if err != nil {
			return err
		}

		for _, event := range evts {
			// Skip moov-io commits (often on branches)
			if strings.HasPrefix(event.RepoSlug, "moov-io/") {
				continue
			}

			key := event.CreatedAt.Format("2006-01-02")
			events[key] = append(events[key], event)
		}
	}

	// Convert push events to HTML output
	var pushEvents []PushEventsTemplate
	for key, commits := range events {
		pushEvents = append(pushEvents, PushEventsTemplate{
			Date:    key,
			Commits: dedupCommits(commits),
		})
	}
	slices.SortFunc(pushEvents, func(e1, e2 PushEventsTemplate) int {
		return -1 * cmp.Compare(e1.Date, e2.Date)
	})

	for idx := range pushEvents {
		slices.SortFunc(pushEvents[idx].Commits, func(c1, c2 source.PushEvent) int {
			return cmp.Compare(c1.RepoSlug, c2.RepoSlug)
		})
	}

	// relative to the project's root dir
	fd, err := os.Create(w.conf.Tracking.OutputFilepath)
	if err != nil {
		return fmt.Errorf("creating index.html failed")
	}
	err = IndexTemplate.Execute(fd, IndexTemplateData{
		PushEvents: pushEvents,
	})
	if err != nil {
		return fmt.Errorf("rendering index.html: %w", err)
	}

	return nil
}

func dedupCommits(events []source.PushEvent) []source.PushEvent {
	byURL := make(map[string]source.PushEvent)
	for _, commits := range events {
		for _, commit := range commits.Commits {
			byURL[commit.CommitURL] = source.PushEvent{
				RepoSlug:  commits.RepoSlug,
				Commits:   []source.WebCommit{commit},
				CreatedAt: commits.CreatedAt,
			}
		}
	}

	var out2 []source.PushEvent
	for _, event := range byURL {
		var found bool
		for idx := range out2 {
			if out2[idx].RepoSlug == event.RepoSlug {
				out2[idx].Commits = append(out2[idx].Commits, event.Commits...)
			}
		}
		if !found {
			out2 = append(out2, event)
		}
	}

	slices.SortFunc(out2, func(e1, e2 source.PushEvent) int {
		return cmp.Compare(e1.RepoSlug, e2.RepoSlug)
	})

	return out2
}

func (w *Worker) getLatestNetworkEvents(ctx context.Context, repo Repository) ([]source.PushEvent, error) {
	sourceClient := w.getSourceClient(repo.Source)
	if sourceClient == nil {
		return nil, nil
	}

	pushEvents, err := sourceClient.ListNetworkPushEvents(ctx, source.Repository{
		Owner: repo.Owner,
		Name:  repo.Name,
	})
	w.logger.Info(fmt.Sprintf("found %d PushEvents for %s/%s", len(pushEvents), repo.Owner, repo.Name))

	if err != nil {
		w.logger.Error(fmt.Sprintf("problem getting latest network events: %v", err))
		return nil, err
	}

	return pushEvents, nil
}
