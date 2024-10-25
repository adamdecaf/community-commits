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
	checkInterval := cmp.Or(queueConf.JobInterval, time.Second)

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

		fmt.Printf("%#v\n", job)

		jobType, ok := job.Payload["type"].(string)
		if !ok {
			return fmt.Errorf("job=%d unexpected type: %T", job.ID, job.Payload["type"])
		}

		w.logger.Info(fmt.Sprintf("handling job %d", job.ID),
			slog.String("job_type", jobType),
		)

		switch strings.ToLower(jobType) {
		case "repository":
			return w.handleRepositoryJob(job)
		}

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

	ctx := context.Background()
	jobID, err := w.queue.Enqueue(ctx, &jobs.Job{
		Queue: w.conf.Tracking.Queue.QueueName,
		Payload: map[string]interface{}{
			"type":   "repository",
			"source": r.Source,
			"owner":  r.Owner,
			"name":   r.Name,
		},
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
