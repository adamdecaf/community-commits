package tracker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/acaloiaro/neoq/jobs"
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

	forks, err := sourceClient.GetForks(ctx, owner, name)
	if err != nil {
		return fmt.Errorf("getting forks: %w", err)
	}

	logger.Info(fmt.Sprintf("found %d forks", len(forks)))

	// TODO(adam): will need to enqueueRepository with future time after job completes

	return nil
}

// rescanEvery := w.conf.Tracking.Queue.RecanEvery
// if rescanEvery < time.Second {
// 	rescanEvery = 24 * time.Hour
// }
// nextScan := time.Now().In(time.UTC).Add(rescanEvery)
