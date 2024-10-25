package tracker

import (
	"context"
	"errors"
	"fmt"

	"github.com/acaloiaro/neoq"
	"github.com/acaloiaro/neoq/backends/postgres"
	"github.com/acaloiaro/neoq/handler"
	"github.com/acaloiaro/neoq/jobs"
)

func (w *Worker) setupNeoq() error {
	ctx := context.Background()

	nq, err := neoq.New(ctx,
		postgres.WithConnectionString(w.conf.Tracking.Queue.ConnectionString),
		neoq.WithBackend(postgres.Backend),
	)
	if err != nil {
		return fmt.Errorf("neoq postgres connect: %w", err)
	}

	w.queue = nq

	return nil
}

func (w *Worker) startProcessingJobs() error {
	if w == nil {
		return nil
	}

	ctx := context.Background()
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

		w.logger.Info(fmt.Sprintf("job: %#v", job))

		return nil
	})
}

func (w *Worker) enqueueRepository(r Repository) error {
	if w == nil {
		return errors.New("missing queue")
	}

	payload := make(map[string]interface{})
	payload["repository"] = r

	ctx := context.Background()
	jobID, err := w.queue.Enqueue(ctx, &jobs.Job{
		Queue:   w.conf.Tracking.Queue.QueueName,
		Payload: payload,
	})
	if err != nil {
		return fmt.Errorf("enqueue of %s failed: %w", r.ID(), err)
	}

	w.logger.Info("enqueued repository",
		"job_id", jobID,
		"repository", r.ID(),
	)

	return nil
}
