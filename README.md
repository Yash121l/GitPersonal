# Forge

Forge is a self-hosted Git platform for small trusted groups. This repository currently contains the Phase 0 bootstrap described in the PRD: a Go service with auth and repository-management API foundations, a production-shaped project layout, and the initial PostgreSQL schema/migrations for the core entities.

## Current Scope

Implemented in this bootstrap:

- HTTP server with graceful shutdown and structured logging
- `GET /healthz` liveness endpoint and `GET /readyz` dependency readiness checks
- JWT cookie auth with register, login, logout, and current-user endpoints
- Repository CRUD API scaffold with ownership checks
- PostgreSQL-backed store when `DATABASE_URL` is set, with in-memory fallback for tests and no-DB runs
- Sharded bare repository provisioning under `FORGE_REPOS_ROOT` with atomic staging, safe deletion, and repo-level mutation locking
- Embedded PostgreSQL migrations applied automatically on startup
- Production-oriented config validation, database pool tuning, request IDs, body limits, and baseline security headers
- Non-root container runtime, health checks, and safer compose defaults for internal deployment
- `docker-compose.yml`, `Dockerfile`, and `sqlc` configuration to anchor local development

Not implemented yet:

- Git smart HTTP / SSH transport
- Pull requests, issues, organizations UI, CI runners, notifications, and web UI
- PostgreSQL-backed repository implementations and Redis-backed session invalidation

## Quick Start

1. Copy `.env.example` to `.env` and adjust the values you care about.
2. Run `docker-compose up --build`.
3. Visit `http://localhost:3000/healthz`.

For local-only development without Docker:

```bash
go mod tidy
go run ./cmd/forge
```

## API Surface

The current API is JSON-only.

- `GET /healthz`
- `GET /readyz`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/me`
- `GET /api/v1/repos`
- `POST /api/v1/repos`
- `DELETE /api/v1/repos/{owner}/{repo}`

Example register request:

```json
{
  "username": "yash",
  "password": "correct horse battery staple"
}
```

Example repository creation request:

```json
{
  "name": "forge",
  "description": "Self-hosted git platform",
  "visibility": "private",
  "default_branch": "main"
}
```

## Layout

- `cmd/forge`: process entrypoint
- `internal/config`: environment-driven application config
- `internal/auth`: password hashing and JWT session helpers
- `internal/database`: database connection bootstrap
- `internal/repository`: repository metadata/filesystem orchestration
- `internal/server`: HTTP router, middleware, and handlers
- `internal/store`: storage interface plus memory and PostgreSQL implementations
- `db/migrations`: PostgreSQL schema evolution
- `db/queries`: starter `sqlc` query definitions
- `deploy`: container and reverse proxy assets

## Architecture Note

Bare repositories on disk are still the correct Git storage primitive here. The scaling work is in the operational layer around them: sharded layout, atomic provisioning, coordinated mutations, background maintenance, transport isolation, and indexing.

## Next Build Step

The natural next step is to layer in git smart HTTP / SSH transport and post-receive indexing against the provisioned repositories, followed by background maintenance tasks such as scheduled `git gc` / commit-graph refresh instead of running expensive maintenance inline with user requests.
