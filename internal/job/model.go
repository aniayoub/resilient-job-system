package job

import "time"

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "processing"
	StatusCompleted Status = "done"
	StatusFailed    Status = "failed"
)

type Job struct {
	ID        string    `json:"id"`
	Status    Status    `json:"status"`
	Payload   string    `json:"payload"`
	Result    string    `json:"result,omitempty"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
