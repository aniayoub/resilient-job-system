package store

import (
	"context"
	"errors"
	"github/aniayoub/resilient-job-system/internal/job"
)

var ErrJobNotFound = errors.New("job not found")

type Store interface {
	GetStatusReport(ctx context.Context) (job.Stats, error)

	Create(ctx context.Context, payload string) (job.Job, error)
	Get(ctx context.Context, id string) (job.Job, error)

	MarkProcessing(ctx context.Context, jobID string) (int, int, error)
	MarkCompleted(ctx context.Context, jobID string, result string) error
	MarkFailed(ctx context.Context, jobID string, errMsg string) error
	MarkRetrying(ctx context.Context, jobID string, errMsg string) (int, error)
}
