# Forge

Forge is a self-hosted Git platform for small trusted groups and for learning how GitHub-style systems actually work. The current repository contains a deployable backend foundation with smart HTTP and SSH Git transport, revocable sessions, PostgreSQL-backed metadata, organizations and collaborators, a browser UI, repository webhooks, and a MkDocs Material documentation site under `docs/`.

## Current Scope

Implemented now:

- HTTP server with graceful shutdown and structured logging
- `GET /healthz` liveness endpoint and `GET /readyz` dependency readiness checks
- JWT cookie auth with register, login, logout, and current-user endpoints
- Session persistence and logout revocation via stored session records
- Repository CRUD API with repository detail responses and owner-aware clone URLs
- PostgreSQL-backed store when `DATABASE_URL` is set, with in-memory fallback for tests and no-DB runs
- Sharded bare repository provisioning under `FORGE_REPOS_ROOT` with atomic staging, safe deletion, and advisory repo-level mutation locking
- Embedded PostgreSQL migrations applied automatically on startup
- Organization ownership, org membership roles, and repository collaborators
- Smart HTTP Git transport through `git-http-backend`
- SSH Git transport with registered public keys and `git-upload-pack` / `git-receive-pack`
- Background repository maintenance for `git gc --auto`, commit-graph refresh, and size accounting
- Repository webhooks for push and delete events with signed async delivery and delivery status tracking
- Browser UI at `/app` for sign-in, repo creation, org management, SSH key management, collaborator management, and webhook management
- Production-oriented config validation, database pool tuning, request IDs, body limits, and baseline security headers
- Non-root container runtime, health checks, and safer compose defaults for internal deployment
- MkDocs Material documentation site under `docs/`
- `docker-compose.yml`, `Dockerfile`, and `sqlc` configuration to anchor local development

Not implemented yet:

- Pull requests, issues, code review flows, releases, and CI runners
- Notifications, admin workflows, and broader instance management
- Rate limiting, CSRF protection, audit logging, and richer security hardening
- Webhook retries with persistent delivery queues, Git LFS, and search

## Quick Start

1. Copy `.env.example` to `.env` and adjust the values you care about.
2. Run `docker-compose up --build -d`.
3. Visit `http://localhost:3000/app`.

For local-only development without Docker:

```bash
go mod tidy
go run ./cmd/forge
```

## Documentation Website

The project ships a MkDocs Material documentation site in `docs/` with a quickstart, guides, and reference pages for architecture, API behavior, data model details, deployment, and testing.

Install the docs dependencies and preview locally:

```bash
python3 -m pip install -r requirements-docs.txt
mkdocs serve
```

Then open `http://127.0.0.1:8000/`.

To produce a production build locally:

```bash
mkdocs build --strict
```

GitHub Pages deployment is wired through `.github/workflows/deploy-pages.yml`. To use it, enable GitHub Pages in the repository settings and choose GitHub Actions as the source.

## API Surface

The current API is JSON plus Git transport, with a browser app mounted at `/app`.

- `GET /healthz`
- `GET /readyz`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/me`
- `GET /api/v1/keys`
- `POST /api/v1/keys`
- `GET /api/v1/orgs`
- `POST /api/v1/orgs`
- `POST /api/v1/orgs/{org}/members`
- `GET /api/v1/repos`
- `GET /api/v1/repos/{owner}/{repo}`
- `POST /api/v1/repos`
- `DELETE /api/v1/repos/{owner}/{repo}`
- `POST /api/v1/repos/{owner}/{repo}/collaborators`
- `GET /api/v1/repos/{owner}/{repo}/webhooks`
- `POST /api/v1/repos/{owner}/{repo}/webhooks`
- `DELETE /api/v1/repos/{owner}/{repo}/webhooks/{webhookID}`
- Smart HTTP Git at `/git/{owner}/{repo}.git`
- SSH Git at `ssh://git@host:2222/{owner}/{repo}.git`
- Browser UI at `/app`, `/app/repos`, `/app/orgs`, `/app/keys`, and `/app/repos/{owner}/{repo}`

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
  "owner": "team",
  "owner_type": "organization",
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
- `docs`: MkDocs documentation source files
- `mkdocs.yml`: MkDocs Material site configuration

## Architecture Note

Bare repositories on disk are still the correct Git storage primitive here. The scaling work is in the operational layer around them: sharded layout, atomic provisioning, PostgreSQL advisory leases, coordinated mutations, background maintenance, transport isolation, authorization, and webhook delivery around Git events.

## Next Build Step

The natural next step is product depth rather than core plumbing: pull requests, issues, release flows, stronger security controls, and a more durable background job model for webhook retries and future automation.
