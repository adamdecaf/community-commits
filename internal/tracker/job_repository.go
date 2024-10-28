package tracker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/acaloiaro/neoq/jobs"
	"github.com/adamdecaf/community-commits/internal/source"
)

func (w *Worker) handleRepositoryJob(logger *slog.Logger, job *jobs.Job) error {
	ctx := context.Background()

	sourceType := job.Payload["source"].(string)
	if sourceType == "" {
		return errors.New("missing source type")
	}
	sourceClient := w.getSourceClient(sourceType)
	if sourceClient == nil {
		return fmt.Errorf("no source.Client found for %s", sourceType)
	}

	// Get forks
	owner := job.Payload["owner"].(string)
	name := job.Payload["name"].(string)

	// Scan the fork for commits
	err := w.updateForks(ctx, logger, sourceType, sourceClient, owner, name)
	if err != nil {
		return fmt.Errorf("updating forks: %w", err)
	}

	return nil
}

// rescanEvery := w.conf.Tracking.Queue.RecanEvery
// if rescanEvery < time.Second {
// 	rescanEvery = 24 * time.Hour
// }
// nextScan := time.Now().In(time.UTC).Add(rescanEvery)

func (w *Worker) updateForks(ctx context.Context, logger *slog.Logger, sourceType string, sourceClient source.Client, owner, name string) error {
	logger = logger.With(
		slog.String("owner", owner),
		slog.String("name", name),
	)

	forks, err := sourceClient.GetForks(ctx, owner, name)
	if err != nil {
		return fmt.Errorf("getting forks: %w", err)
	}

	logger.Info(fmt.Sprintf("found %d forks", len(forks)))

	// Enqueue each fork to be scanned
	for idx, fork := range forks {
		nextScan := time.Now().Add(time.Duration(idx) * time.Minute)
		repo := Repository{
			Source: sourceType,
			Owner:  fork.Owner,
			Name:   fork.Name,
		}

		err = w.saveNewerCommits(ctx, logger, sourceClient, repo)
		if err != nil {
			return fmt.Errorf("saving newer commits from %s: %w", repo.ID(), err)
		}

		err = w.enqueueRepository(repo, nextScan)
		if err != nil {
			return fmt.Errorf("queue of fork %s failed: %w", repo.ID(), err)
		}
	}

	return nil
}

func (w *Worker) saveNewerCommits(ctx context.Context, logger *slog.Logger, sourceClient source.Client, repo Repository) error {
	sourceRepository := source.Repository{
		Owner: repo.Owner,
		Name:  repo.Name,
	}
	branches, err := sourceClient.ListBranches(ctx, sourceRepository)
	if err != nil {
		return fmt.Errorf("listing branches: %w", err)
	}

	for _, branch := range branches {
		commits, err := sourceClient.ListCommits(ctx, sourceRepository, branch)
		if err != nil {
			return fmt.Errorf("listing %s commits from %s: %w", branch.Name, repo.ID(), err)
		}

		err = w.forkRepository.SaveCommits(ctx, sourceRepository, branch, commits)
		if err != nil {
			return fmt.Errorf("saving commits: %w", err)
		}
	}

	return nil
}
