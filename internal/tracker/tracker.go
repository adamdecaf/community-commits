package tracker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/adamdecaf/community-commits/internal/forks"

	"github.com/acaloiaro/neoq"
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

	queue neoq.Neoq

	forkRepository forks.Repository
}

func NewWorker(logger *slog.Logger, conf Config, forkRepository forks.Repository) (*Worker, error) {
	w := &Worker{
		conf:   conf,
		logger: logger,

		forkRepository: forkRepository,
	}

	err := w.setupNeoq()
	if err != nil {
		return nil, fmt.Errorf("setting up neoq: %w", err)
	}

	return w, nil
}

func (w *Worker) Sync() error {
	// We want to scan repositories on startup
	nextScan := time.Now().In(time.UTC)

	// For each repository grab the forks and insert each as an item to crawl
	for _, repo := range w.conf.Tracking.Repositories {
		err := w.enqueueRepository(repo, nextScan)
		if err != nil {
			return fmt.Errorf("enqueue %v failed: %w", repo.ID(), err)
		}
	}
	return nil
}

func (w *Worker) Start(ctx context.Context) error {
	err := w.startProcessingJobs(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (w *Worker) Stop() error {
	err := w.stopProcessingJobs()
	if err != nil {
		return err
	}

	return nil
}
