package tracker

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/acaloiaro/neoq"
	"github.com/acaloiaro/neoq/backends/postgres"
	"github.com/acaloiaro/neoq/handler"
	"github.com/acaloiaro/neoq/jobs"
	"github.com/acaloiaro/neoq/logging"
)

func (w *Worker) setupNeoq() error {
	ctx := context.Background()

	queueConf := w.conf.Tracking.Queue
	checkInterval := cmp.Or(queueConf.WorkerInterval, time.Second)

	nq, err := neoq.New(ctx,
		neoq.WithJobCheckInterval(checkInterval),
		neoq.WithBackend(postgres.Backend),
		neoq.WithLogLevel(logging.LogLevelDebug),
		postgres.WithConnectionString(queueConf.ConnectionString),
	)
	if err != nil {
		return fmt.Errorf("neoq postgres connect: %w", err)
	}

	nq.SetLogger(w.logger)

	w.queue = nq

	return nil
}

func (w *Worker) startProcessingJobs(ctx context.Context) error {
	if w == nil {
		return nil
	}

	err := w.queue.Start(ctx, w.handleJobs())
	if err != nil {
		return fmt.Errorf("starting neoq processing: %w", err)
	}

	return nil
}

func (w *Worker) stopProcessingJobs() error {
	if w == nil {
		return nil
	}

	w.queue.Shutdown(context.Background())
	return nil
}

func (w *Worker) handleJobs() handler.Handler {
	queueName := w.conf.Tracking.Queue.QueueName

	return handler.New(queueName, func(ctx context.Context) error {
		job, err := jobs.FromContext(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("%#v\n", job)

		jobType, ok := job.Payload["type"].(string)
		if !ok {
			return fmt.Errorf("job=%d unexpected type: %T", job.ID, job.Payload["type"])
		}

		logger := w.logger.With(
			slog.String("job_id", fmt.Sprintf("%v", job.ID)),
			slog.String("job_type", jobType),
		)
		logger.Info("handling job")

		switch strings.ToLower(jobType) {
		case "repository":
			err = w.handleRepositoryJob(logger, job)
			if err != nil {
				logger.Error("problem with job", slog.String("error", err.Error()))
				return err
			}
		}

		logger.Info("finished job")
		return nil
	},
		handler.Concurrency(1),
	)
}

func (w *Worker) enqueueRepository(r Repository, nextScan time.Time) error {
	if w == nil {
		return errors.New("missing queue")
	}

	ctx := context.Background()
	jobID, err := w.queue.Enqueue(ctx, &jobs.Job{
		Queue: w.conf.Tracking.Queue.QueueName,
		Payload: map[string]interface{}{
			"type":   "repository",
			"source": r.Source,
			"owner":  r.Owner,
			"name":   r.Name,
		},
		RunAfter: nextScan,
	})
	if err != nil {
		// Skip if we've already enqueued this repository
		if errors.Is(err, postgres.ErrDuplicateJob) {
			return nil
		}
		return fmt.Errorf("enqueue of %s failed: %w", r.ID(), err)
	}

	w.logger.Info("enqueued repository",
		slog.String("job_id", jobID),
		slog.String("repository", r.ID()),
	)

	return nil
}
