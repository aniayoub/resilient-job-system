# Resilient Job System

This is a Go learning project built in phases.

Current phase: a basic resilient job system with an HTTP API, an in-memory store, a shared queue, and a single background worker.

Small Go service for submitting background jobs and polling their status over HTTP.

The project exposes a simple API that stores jobs in memory, queues them for asynchronous processing, and updates each job as it moves through its lifecycle.

## What It Does

- Accepts job creation requests over HTTP.
- Queues jobs for background processing.
- Tracks job state in memory.
- Simulates work with a single worker.
- Randomly fails some jobs to demonstrate error handling.

## Project Layout

- `cmd/api`: starts the HTTP API server on `localhost:8080`.
- `cmd/flood`: sends many job creation requests to the API for quick load testing.
- `internal/httpapi`: HTTP handlers and route registration.
- `internal/worker`: background worker loop.
- `internal/store`: in-memory job storage.
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
  "created_at": "2026-05-18T12:00:00Z",
  "updated_at": "2026-05-18T12:00:00Z"
}
```

### `GET /jobs/{id}`

Returns the current state of a job.

Possible job statuses:

- `pending`
- `processing`
- `done`
- `failed`

Completed jobs include a `result` field. Failed jobs also return an error/result message from the worker.

## Notes

- Storage is in memory only, so all jobs are lost when the API process stops.
- Processing is simulated with a fixed delay of about 3 seconds.
- The worker intentionally fails some jobs to exercise retry/error paths during development.