package tracker

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/acaloiaro/neoq"
	"github.com/acaloiaro/neoq/backends/postgres"
	"github.com/acaloiaro/neoq/handler"
	"github.com/acaloiaro/neoq/jobs"
)

func (w *Worker) setupNeoq() error {
	ctx := context.Background()

	queueConf := w.conf.Tracking.Queue
	checkInterval := cmp.Or(queueConf.JobInterval, time.Second)

	nq, err := neoq.New(ctx,
		neoq.WithJobCheckInterval(checkInterval),
		neoq.WithBackend(postgres.Backend),
		postgres.WithConnectionString(queueConf.ConnectionString),
	)
	if err != nil {
		return fmt.Errorf("neoq postgres connect: %w", err)
	}

	w.queue = nq

	return nil
}

func (w *Worker) startProcessingJobs(ctx context.Context) error {
	if w == nil {
		return nil
	}

	jobInterval := w.conf.Tracking.Queue.JobInterval
	if jobInterval <= time.Second {
		jobInterval = time.Hour
	}

	schedule := formatCrontabSchedule(jobInterval)
	err := w.queue.StartCron(ctx, schedule, w.handleJobs())
	if err != nil {
		return fmt.Errorf("starting neoq processing: %w", err)
	}

	return nil
}

func formatCrontabSchedule(interval time.Duration) string {
	s, m, h, d := "0", "0", "*", "*"

	days := int(interval.Truncate(24*time.Hour).Hours()) / 24
	if days >= 1 {
		days += 1 // account for Truncate

		d = fmt.Sprintf("1/%d", days)
	}

	hours := int(interval.Truncate(time.Hour).Hours()) % 24
	if hours >= 1 {
		if days >= 1 {
			h = fmt.Sprintf("%d", hours)
		} else {
			h = fmt.Sprintf("1/%d", hours)
		}
	}

	mins := int(interval.Truncate(time.Minute).Minutes()) % 60
	if mins >= 1 {
		if hours > 1 {
			m = fmt.Sprintf("%d", mins)
		} else {
			m = fmt.Sprintf("1/%d", mins)
		}
	}

	secs := int(interval.Seconds()) % 60
	if secs >= 1 {
		s = fmt.Sprintf("1/%d", secs)
	}

	return fmt.Sprintf("%s %s %s %s * *", s, m, h, d)
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

	return handler.NewPeriodic(func(ctx context.Context) error {
		job, err := jobs.FromContext(ctx)
		if err != nil {
			return err
		}

		w.logger.Info(fmt.Sprintf("job: %#v", job))

		return nil
	},
		handler.Concurrency(1),
		handler.Queue(queueName),
	)
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
		// Skip if we've already enqueued this repository
		if errors.Is(err, postgres.ErrDuplicateJob) {
			return nil
		}
		return fmt.Errorf("enqueue of %s failed: %w", r.ID(), err)
	}

	w.logger.Info("enqueued repository",
		"job_id", jobID,
		"repository", r.ID(),
	)

	return nil
}
