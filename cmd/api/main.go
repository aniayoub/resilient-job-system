package main

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github/aniayoub/resilient-job-system/internal/config"
	"github/aniayoub/resilient-job-system/internal/httpapi"
	"github/aniayoub/resilient-job-system/internal/logging"
	"github/aniayoub/resilient-job-system/internal/store"
	"github/aniayoub/resilient-job-system/internal/worker"

	_ "github.com/lib/pq"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// Initialize structured logger
	logger := logging.New()

	cfg, err := config.Load()

	if err != nil {
		logger.Error("failed to load config", slog.Any("error", err))
		return
	}

	// Initialize the job store
	jobStore, dbCloser, err := initStore(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		logger.Error("failed to initialize store", slog.Any("error", err))
		return
	}
	defer dbCloser.Close()

	// Initialize a shared queue for workers
	queue := make(chan string, cfg.QueueSize)

	// Initialize handler with the store and queue
	handler := httpapi.NewHandler(jobStore, queue, logger.With("component", "httpapi"))
	handler.RegisterRoutes()

	// Initialize and start a pool of workers
	workerCount := cfg.WorkerCount
	for i := 0; i < workerCount; i++ {
		worker := worker.NewWorker(i, jobStore, queue, logger.With("component", "worker", "worker_id", i))
		worker.Start(ctx)
	}

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: logging.WithRequestLogging(logger.With("component", "http"), http.DefaultServeMux),
	}

	go func() {
		// Listen to http requests and handle job creation, status retrieval, etc.
		logger.Info("starting server", slog.String("addr", server.Addr))

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(
				"server crashed",
				slog.Any("error", err),
			)
		}
	}()

	<-ctx.Done()

	logger.Info("shutting down server")

	// Gracefully shutdown the server with a timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		logger.Error("server shutdown error", slog.Any("error", err))
	} else {
		logger.Info("server shutdown complete")
	}
}

func initStore(ctx context.Context, databaseURL string, logger *slog.Logger) (store.Store, io.Closer, error) {
	//databaseURL := "postgres://user:password@localhost:5432/resilient-job-system?sslmode=disable"

	// Open a connection to the PostgreSQL database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Error("failed to connect to database", slog.Any("error", err))
		return nil, nil, err
	}

	logger.Info("connected to database")

	// Verify the database connection
	err = db.PingContext(ctx)
	if err != nil {
		logger.Error("failed to ping database", slog.Any("error", err))
		return nil, nil, err
	}

	logger.Info("database connection verified")
	return store.NewPostgresStore(db), db, nil
}
