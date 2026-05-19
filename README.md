# Resilient Job System

This is a Go learning project built in phases.

Current phase: a basic resilient job system with an HTTP API, an in-memory store, a shared queue, a worker pool, retry handling, and graceful shutdown support.

Small Go service for submitting background jobs and polling their status over HTTP.

The project exposes a simple API that stores jobs in memory, queues them for asynchronous processing, and updates each job as it moves through its lifecycle.

## What It Does

- Accepts job creation requests over HTTP.
- Queues jobs for background processing.
- Tracks job state in memory.
- Processes jobs with a pool of workers.
- Retries failed jobs up to a configured limit.
- Returns a temporary error when the queue is full.
- Cancels work cleanly during server shutdown.

## Project Layout

- `cmd/api`: starts the HTTP API server on `localhost:8080`.
- `cmd/flood`: sends many job creation requests to the API for quick load testing.
- `internal/httpapi`: HTTP handlers and route registration.
- `internal/worker`: worker pool, retry behavior, and timeout-aware job execution.
- `internal/store`: in-memory job storage and retry tracking.
- `internal/job`: job model and statuses.

## Run

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
- `done`
- `failed`

Job responses also include retry metadata:

- `retry_count`: how many retry attempts have already been used.
- `max_retries`: maximum retries allowed for the job.
- `last_error`: the latest processing error before a retry or final failure.

Completed jobs include a `result` field. Failed jobs include the final failure information.

## Notes

- Storage is in memory only, so all jobs are lost when the API process stops.
- Processing is simulated with a fixed delay of about 3 seconds.
- Each job runs with a 5-second context timeout.
- The API currently starts 5 workers.
- The server shuts down gracefully on `SIGINT` and `SIGTERM`.
- The worker intentionally fails some jobs to exercise retry and failure paths during development.

## Documentation Note

This README was drafted with AI assistance and reviewed by the developer.