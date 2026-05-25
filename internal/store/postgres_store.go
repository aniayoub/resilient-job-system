package store

import (
	"context"
	"database/sql"
	"fmt"
	"github/aniayoub/resilient-job-system/internal/job"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Create(ctx context.Context, payload string) (job.Job, error) {
	var j job.Job
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO jobs (
			status,
			payload,
			max_retries,
			retry_count,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, status, payload, max_retries, retry_count, created_at, updated_at
	`, job.StatusPending, payload, 5, 0).Scan(
		&j.ID,
		&j.Status,
		&j.Payload,
		&j.MaxRetries,
		&j.RetryCount,
		&j.CreatedAt,
		&j.UpdatedAt,
	)

	if err != nil {
		return job.Job{}, err
	}

	return j, nil
}

func (s *PostgresStore) Get(ctx context.Context, id string) (job.Job, error) {
	var job job.Job

	err := s.db.QueryRowContext(ctx,
		`
		SELECT id, status, payload, max_retries, retry_count, created_at, updated_at
		FROM jobs
		WHERE id = $1
	`, id).Scan(
		&job.ID,
		&job.Status,
		&job.Payload,
		&job.MaxRetries,
		&job.RetryCount,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		return job, err
	}

	return job, nil
}

func (s *PostgresStore) MarkProcessing(ctx context.Context, jobID string) (int, int, error) {
	var retryCount, maxRetries int

	err := s.db.QueryRowContext(ctx, `
		UPDATE jobs
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND status = $3
		RETURNING retry_count, max_retries
	`,
		job.StatusRunning, jobID, job.StatusPending).Scan(
		&retryCount,
		&maxRetries,
	)
	if err != nil {
		return 0, 0, err
	}

	// Get retry count and max retries for the job
	return retryCount, maxRetries, nil
}

func (s *PostgresStore) MarkDone(ctx context.Context, jobID string, result string) error {
	dbResult, err := s.db.ExecContext(ctx, `
		UPDATE jobs
		SET status = $1, updated_at = NOW(), result = $3, error_message = NULL
		WHERE id = $2 AND status = $4
	`, job.StatusCompleted, jobID, result, job.StatusRunning)
	if err != nil {
		return err
	}

	rowsAffected, err := dbResult.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return fmt.Errorf("expected to update 1 row, but updated %d", rowsAffected)
	}
	return nil
}

func (s *PostgresStore) MarkFailed(ctx context.Context, jobID string, errMsg string) error {
	dbResult, err := s.db.ExecContext(ctx, `
		UPDATE jobs
		SET status = $1, updated_at = NOW(), error_message = $3, result = NULL
		WHERE id = $2 AND status = $4
	`, job.StatusFailed, jobID, errMsg, job.StatusRunning)
	if err != nil {
		return err
	}

	rowsAffected, err := dbResult.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return fmt.Errorf("expected to update 1 row, but updated %d", rowsAffected)
	}
	return nil
}

func (s *PostgresStore) MarkRetrying(ctx context.Context, jobID string, errMsg string) (int, error) {
	var retryCount int
	err := s.db.QueryRowContext(ctx, `
		UPDATE jobs
		SET retry_count = retry_count + 1, status = $1, updated_at = NOW(), error_message = $3
		WHERE id = $2 AND status = $4
		RETURNING retry_count
	`, job.StatusPending, jobID, errMsg, job.StatusRunning).Scan(&retryCount)

	if err != nil {
		return 0, err
	}
	return retryCount, nil
}
