package store

import (
	"context"
	"github/aniayoub/resilient-job-system/internal/job"
)

type Store interface {
	Create(ctx context.Context, payload string) (job.Job, error)
	Get(ctx context.Context, id string) (job.Job, error)

	MarkProcessing(ctx context.Context, jobID string) (int, int, error)
	MarkDone(ctx context.Context, jobID string, result string) error
	MarkFailed(ctx context.Context, jobID string, errMsg string) error
	MarkRetrying(ctx context.Context, jobID string, errMsg string) (int, error)
}
