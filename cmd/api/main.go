package main

import (
	"context"
	"log/slog"
	"net/http"

	"github/aniayoub/resilient-job-system/internal/httpapi"
	"github/aniayoub/resilient-job-system/internal/logging"
	"github/aniayoub/resilient-job-system/internal/store"
	"github/aniayoub/resilient-job-system/internal/worker"
)

func main() {
	ctx := context.Background()
	logger := logging.New()

	// Initialize an in-memory jobStore for jobs
	jobStore := store.NewStore()

	// Initialize a shared queue for workers
	queue := make(chan string, 100)

	// Initialize handler with the store and queue
	handler := httpapi.NewHandler(jobStore, queue, logger.With("component", "httpapi"))
	handler.RegisterRoutes()

	// Initialize and start a pool of workers
	workerCount := 5
	for i := 0; i < workerCount; i++ {
		worker := worker.NewWorker(i, jobStore, queue, logger.With("component", "worker", "worker_id", i))
		worker.Start(ctx)
	}

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: logging.WithRequestLogging(logger.With("component", "http"), http.DefaultServeMux),
	}

	// Listen to http requests and handle job creation, status retrieval, etc.
	logger.Info("starting server", slog.String("addr", server.Addr))

	err := server.ListenAndServe()
	if err != nil {
		logger.Error("server stopped", slog.Any("error", err))
	}
}
