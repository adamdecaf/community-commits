package tracker

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

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
			key := event.CreatedAt.Format("2006-01-02")
			events[key] = append(events[key], event)
		}
	}

	// Convert push events to HTML output
	var pushEvents []PushEventsTemplate
	for key, commits := range events {
		pushEvents = append(pushEvents, PushEventsTemplate{
			Date:    key,
			Commits: commits,
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
	fd, err := os.Create(filepath.Join("docs", "networks", "index.html"))
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

func (w *Worker) getLatestNetworkEvents(ctx context.Context, repo Repository) ([]source.PushEvent, error) {
	sourceClient := w.getSourceClient(repo.Source)
	if sourceClient == nil {
		return nil, nil
	}
	pushEvents, err := sourceClient.ListNetworkPushEvents(ctx, source.Repository{
		Owner: repo.Owner,
		Name:  repo.Name,
	})
	if err != nil {
		return nil, err
	}
	return pushEvents, nil
}
