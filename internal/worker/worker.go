package worker

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github/aniayoub/resilient-job-system/internal/store"
)

type Worker struct {
	id     int
	store  *store.Store
	queue  chan string // Workers should only read from the queue, so we use a receive-only channel
	logger *slog.Logger
}

func NewWorker(id int, s *store.Store, queue chan string, logger *slog.Logger) *Worker {
	return &Worker{
		id:     id,
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

	retryCount, maxRetries, err := w.store.MarkProcessing(jobID)
	if err != nil {
		w.logger.Error("failed to mark job as processing", "job_id", jobID, "error", err)
		return
	}

	w.logger.Info("job marked as processing", "job_id", jobID, "retry_count", retryCount, "max_retries", maxRetries)

	// Simulate job processing time
	time.Sleep(3 * time.Second)

	// Fail randomly
	if rand.Intn(10) < 5 { // 50% chance of failure
		w.handleFailure(jobID, retryCount, maxRetries, "simulated failure")
		return
	}

	err = w.store.MarkDone(jobID, "job completed")
	if err != nil {
		w.logger.Error("failed to mark job as done", "job_id", jobID, "error", err)
		return
	}

	w.logger.Info("job processing completed", "job_id", jobID, "result", "job completed")
}

func (w *Worker) handleFailure(jobID string, retryCount int, maxRetries int, errMsg string) {
	if retryCount >= maxRetries {
		w.logger.Error("job failed after max retries", "job_id", jobID, "retry_count", retryCount)
		err := w.store.MarkFailed(jobID, "job failed after max retries")
		if err != nil {
			w.logger.Error("failed to mark job as failed", "job_id", jobID, "error", err)
		}
		return
	}

	w.logger.Error("job processing failed", "job_id", jobID, "error", errMsg)
	retryCount, err := w.store.MarkRetrying(jobID, errMsg)
	if err != nil {
		w.logger.Error("failed to mark job as retrying", "job_id", jobID, "error", err)

		// For simplicity, if we fail to mark the job as retrying, we won't re-queue it. In a real system, you might want to handle this case more robustly.
		return
	}

	// Re-queue with delay to avoid immediate retry storms
	// Note 1: No need for go routine here since time.AfterFunc will handle the asynchronous execution
	// NOTE 2: using time.AfterFunc scales badly. In real production, maybe another queue with delayed processing or a scheduler would be better(?)
	time.AfterFunc(time.Duration(retryCount)*time.Second, func() {
		w.queue <- jobID
		w.logger.Info("job re-queued for retry", "job_id", jobID, "retry_count", retryCount)
	})

}
