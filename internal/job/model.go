package job

import "time"

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "processing"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

type Job struct {
	ID          string    `json:"id"`
	Status      Status    `json:"status"`
	Payload     string    `json:"payload"`
	Result      string    `json:"result,omitempty"`
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	RetryCount  int       `json:"retry_count"`
	MaxRetries  int       `json:"max_retries"`
	LastError   string    `json:"last_error,omitempty"`
}
