# Resilient Job System

This is a Go learning project built in phases.

Current phase: a basic resilient job system with an HTTP API, PostgreSQL persistence, lifecycle reporting, a worker pool, retry handling, and graceful shutdown support.

Small Go service for submitting background jobs, polling their status, and viewing aggregate job counts over HTTP.

The project exposes a simple API that persists jobs in PostgreSQL, queues them for asynchronous processing, and updates each job as it moves through its lifecycle.

## What It Does

- Accepts job creation requests over HTTP.
- Queues jobs for background processing.
- Persists job state through a store abstraction.
- Processes jobs with a pool of workers.
- Retries failed jobs up to a configured limit.
- Exposes an aggregate status report endpoint.
- Returns a temporary error when the queue is full.
- Cancels work cleanly during server shutdown.

## Project Layout

- `cmd/api`: starts the HTTP API server on `localhost:8080`.
- `cmd/flood`: sends many job creation requests to the API for quick load testing.
- `internal/config`: loads runtime configuration from environment variables.
- `internal/httpapi`: HTTP handlers and route registration.
- `internal/worker`: worker pool, retry behavior, and timeout-aware job execution.
- `internal/store`: store interface and PostgreSQL implementation.
- `internal/job`: job model, statuses, and aggregate stats model.
- `migrations`: SQL schema for the PostgreSQL-backed job store.

## Run

The API currently uses PostgreSQL directly and loads its runtime settings from environment variables.

Create the database and apply the migration before starting the API:

```bash
createdb -h localhost -U user resilient-job-system
psql -h localhost -U user -d resilient-job-system -f migrations/001_create_jobs.sql
psql -h localhost -U user -d resilient-job-system -f migrations/002_jobs_completed_extension.sql
```

Set the required configuration before starting the API:

```bash
export DATABASE_URL='postgres://user:password@localhost:5432/resilient-job-system?sslmode=disable'
export WORKER_COUNT=5
export QUEUE_SIZE=100
```

`DATABASE_URL` is required. `WORKER_COUNT` and `QUEUE_SIZE` are optional and default to `5` and `100`.

Start the API server:

```bash
go run ./cmd/api
```

In another terminal, create a job:

```bash
curl -X POST http://localhost:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{"payload":"process this job"}'
```

Fetch a job by ID:

```bash
curl http://localhost:8080/jobs/<job-id>
```

Fetch an aggregate job status report:

```bash
curl http://localhost:8080/status
```

Run the flood client to submit multiple jobs concurrently:

```bash
go run ./cmd/flood
```

## API

### `POST /jobs`

Creates a new job.

Request body:

```json
{
  "payload": "process this job"
}
```

Successful response returns `201 Created` with the new job:

```json
{
  "id": "2e6f7d48-2f8d-4d07-b967-0f6ef40e3f8f",
  "status": "pending",
  "payload": "process this job",
  "retry_count": 0,
  "max_retries": 5,
  "created_at": "2026-05-18T12:00:00Z",
  "updated_at": "2026-05-18T12:00:00Z"
}
```

If the queue is full, the API returns `503 Service Unavailable`:

```json
{
  "error": "job queue is full, please try again later"
}
```

### `GET /jobs/{id}`

Returns the current state of a job.

Possible job statuses:

- `pending`
- `processing`
- `completed`
- `failed`

Job responses also include retry metadata:

- `retry_count`: how many retry attempts have already been used.
- `max_retries`: maximum retries allowed for the job.
- `last_error`: the latest processing error before a retry or final failure.
- `completed_at`: when the job was marked as completed.

Completed jobs include a `result` field. Failed jobs include the final failure information.

### `GET /status`

Returns aggregate job counts by lifecycle state.

Example response:

```json
{
  "pending": 3,
  "running": 2,
  "completed": 18,
  "failed": 1
}
```

## Notes

- The API entry point currently uses the PostgreSQL store.
- Runtime configuration is loaded from environment variables.
- The `completed_at` column is added by the second migration.
- Processing is simulated with a fixed delay of about 3 seconds.
- Each job runs with a 5-second context timeout.
- The API defaults to 5 workers and a queue size of 100 unless overridden by config.
- The server shuts down gracefully on `SIGINT` and `SIGTERM`.
- The worker intentionally fails some jobs to exercise retry and failure paths during development.

## Documentation Note

This README was drafted with AI assistance and reviewed by the developer.