package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github/aniayoub/resilient-job-system/internal/logging"
	"github/aniayoub/resilient-job-system/internal/store"
)

type Handler struct {
	store  store.Store
	queue  chan<- string // Handlers should only write to the queue, so we use a send-only channel
	logger *slog.Logger
}

type CreateJobRequest struct {
	Payload string `json:"payload"`
}

func NewHandler(store store.Store, queue chan<- string, logger *slog.Logger) *Handler {
	return &Handler{store: store, queue: queue, logger: logger}
}

func (h *Handler) RegisterRoutes() {
	http.HandleFunc("/jobs", h.CreateJob)
	http.HandleFunc("/jobs/", h.GetJob)
	http.HandleFunc("/status", h.GetStatus)
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With("request_id", logging.RequestIDFromContext(r.Context()))

	if r.Method != http.MethodGet {
		logger.Warn("method not allowed", "method", r.Method, "path", r.URL.Path)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.store.GetStatusReport(r.Context())
	if err != nil {
		logger.Error("failed to get status report", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(stats)
	if err != nil {
		logger.Error("failed to encode status report", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: Consider removing this log or changing it to debug level in production, as it may be too verbose, and might be considered sensitive information in some contexts.
	logger.Info("status report fetched", "pending", stats.Pending, "running", stats.Running, "completed", stats.Completed, "failed", stats.Failed)
}

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With("request_id", logging.RequestIDFromContext(r.Context()))

	if r.Method != http.MethodGet {
		logger.Warn("method not allowed", "method", r.Method, "path", r.URL.Path)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/jobs/"):]
	if id == "" {
		logger.Warn("missing job id", "path", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	job, err := h.store.Get(r.Context(), id)
	if err != nil {
		if err == store.ErrJobNotFound {
			logger.Warn("job not found", "job_id", id)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		logger.Error("failed to fetch job", "job_id", id, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(job)
	if err != nil {
		logger.Error("failed to encode job", "job_id", job.ID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Info("job fetched", "job_id", job.ID, "status", job.Status)
}

func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With("request_id", logging.RequestIDFromContext(r.Context()))

	// Only allow POST method for creating jobs
	if r.Method != http.MethodPost {
		logger.Warn("method not allowed", "method", r.Method, "path", r.URL.Path)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req CreateJobRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		logger.Warn("invalid create job payload", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Basic validation to ensure payload is provided
	if req.Payload == "" {
		logger.Warn("missing payload in create job request")
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(map[string]string{
			"error": "payload is required",
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Create a new job in the store
	j, err := h.store.Create(r.Context(), req.Payload)
	if err != nil {
		logger.Error("failed to create job", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger.Info("job created", "job_id", j.ID, "status", j.Status, "payload_size", len(req.Payload))

	// Send the job ID to the queue for processing
	select {
	case h.queue <- j.ID:
		// Job ID successfully sent to the queue
		logger.Info("job queued", "job_id", j.ID)

		// Send the created job back in the response
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(j)
		if err != nil {
			logger.Error("failed to encode created job", "job_id", j.ID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		// Queue is full, mark job as failed and return an error response
		logger.Error("job queue is full, cannot process job", "job_id", j.ID)

		w.WriteHeader(http.StatusServiceUnavailable)
		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(map[string]string{
			"error": "job queue is full, please try again later",
		})
		if err != nil {
			logger.Error("failed to encode queue full response", "job_id", j.ID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

}
