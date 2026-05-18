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

func (s *Store) MarkProcessing(jobID string) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	j, exists := s.jobs[jobID]
	if !exists {
		return ErrJobNotFound
	}

	if j.Status != job.StatusPending {
		return errors.New("job is not in pending state")
	}
	j.Status = job.StatusRunning
	j.UpdatedAt = time.Now().UTC()

	s.jobs[jobID] = j

	return nil
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

func NewStore() *Store {
	return &Store{
		jobs: make(map[string]job.Job),
	}
}

func (s *Store) CreateJob(payload string) job.Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := job.Job{
		ID:        uuid.New().String(),
		Status:    job.StatusPending,
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
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
