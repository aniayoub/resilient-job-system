package store

import (
	"errors"
	"sync"
	"time"

	"github/aniayoub/resilient-job-system/internal/job"

	"github.com/google/uuid"
)

var ErrJobNotFound = errors.New("job not found")

type Store struct {
	mu   sync.RWMutex
	jobs map[string]job.Job
}

func NewStore() *Store {
	return &Store{
		jobs: make(map[string]job.Job),
	}
}

func (s *Store) CreateJob(payload string) job.Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := job.Job{
		ID:         uuid.New().String(),
		Status:     job.StatusPending,
		Payload:    payload,
		MaxRetries: 5,
		RetryCount: 0,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	s.jobs[j.ID] = j
	return j
}

func (s *Store) GetJob(id string) (job.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, exists := s.jobs[id]
	if !exists {
		return job.Job{}, ErrJobNotFound
	}
	return j, nil
}

func (s *Store) MarkProcessing(jobID string) (int, int, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	j, exists := s.jobs[jobID]
	if !exists {
		return 0, 0, ErrJobNotFound
	}

	if j.Status != job.StatusPending {
		return 0, 0, errors.New("job is not in pending state")
	}
	j.Status = job.StatusRunning
	j.UpdatedAt = time.Now().UTC()

	s.jobs[jobID] = j

	return j.RetryCount, j.MaxRetries, nil
}

func (s *Store) MarkFailed(jobID string, errMsg string) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	j, exists := s.jobs[jobID]
	if !exists {
		return ErrJobNotFound
	}

	if j.Status != job.StatusRunning {
		return errors.New("job is not in processing state")
	}

	j.Status = job.StatusFailed
	j.Result = errMsg
	j.UpdatedAt = time.Now().UTC()
	s.jobs[jobID] = j

	return nil
}

func (s *Store) MarkDone(jobID string, result string) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	j, exists := s.jobs[jobID]
	if !exists {
		return ErrJobNotFound
	}

	if j.Status != job.StatusRunning {
		return errors.New("job is not in processing state")
	}

	j.Status = job.StatusCompleted
	j.Result = result
	j.UpdatedAt = time.Now().UTC()
	s.jobs[jobID] = j

	return nil
}

func (s *Store) MarkRetrying(jobID string, errMsg string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	j, exists := s.jobs[jobID]
	if !exists {
		return 0, ErrJobNotFound
	}

	if j.Status != job.StatusRunning {
		return 0, errors.New("job is not in processing state")
	}

	j.Status = job.StatusPending
	j.LastError = errMsg
	j.UpdatedAt = time.Now().UTC()
	j.RetryCount++

	s.jobs[jobID] = j

	return j.RetryCount, nil
}
