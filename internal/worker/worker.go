package worker

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github/aniayoub/resilient-job-system/internal/store"
)

type Worker struct {
	store  *store.Store
	queue  <-chan string // Workers should only read from the queue, so we use a receive-only channel
	logger *slog.Logger
}

func NewWorker(s *store.Store, queue <-chan string, logger *slog.Logger) *Worker {
	return &Worker{
		store:  s,
		queue:  queue,
		logger: logger,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("worker started")

	go func() {
		for {
			select {
			case <-ctx.Done():
				w.logger.Info("worker stopped", "reason", ctx.Err())
				return

			case jobID, ok := <-w.queue:
				if !ok {
					w.logger.Info("worker stopped", "reason", "queue closed")
					return
				}
				w.process(jobID)
			}
		}
	}()
}

func (w *Worker) process(jobID string) {
	w.logger.Info("job processing started", "job_id", jobID)

	err := w.store.MarkProcessing(jobID)
	if err != nil {
		w.logger.Error("failed to mark job as processing", "job_id", jobID, "error", err)
		return
	}

	// Simulate job processing time
	time.Sleep(3 * time.Second)

	// Fail randomly
	if rand.Intn(10) < 3 { // 30% chance of failure
		w.logger.Error("job processing failed", "job_id", jobID, "error", "simulated failure")
		w.store.MarkFailed(jobID, "simulated failure")
		return
	}

	err = w.store.MarkDone(jobID, "job completed")
	if err != nil {
		w.logger.Error("failed to mark job as done", "job_id", jobID, "error", err)
		return
	}

	w.logger.Info("job processing completed", "job_id", jobID, "result", "job completed")
}
