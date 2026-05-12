package httpapi

import (
	"encoding/json"
	"net/http"

	"github/aniayoub/resilient-job-system/internal/job"
	"github/aniayoub/resilient-job-system/internal/store"

	"github.com/google/uuid"
)

type Handler struct {
	store *store.Store
}

type CreateJobRequest struct {
	Payload string `json:"payload"`
}

func NewHander(store *store.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes() {
	http.HandleFunc("/jobs", h.CreateJob)
	http.HandleFunc("/jobs/", h.GetJob)
}

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/jobs/"):]
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	job, err := h.store.GetJob(id)
	if err != nil {
		if err == store.ErrJobNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(job)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req CreateJobRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Payload == "" {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(map[string]string{
			"error": "payload is required",
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	j := job.Job{
		ID:      uuid.New().String(),
		Status:  job.StatusPending,
		Payload: req.Payload,
	}

	h.store.CreateJob(j)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(j)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
