package store

import (
	"errors"
	"sync"

	"github/aniayoub/resilient-job-system/internal/job"
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

func (s *Store) CreateJob(j job.Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[j.ID] = j
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
