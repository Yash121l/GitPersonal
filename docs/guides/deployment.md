# Deployment

Forge is written to be run by the person learning from it. The default posture is local-first self-hosting.

## Quick start

```bash
cp .env.example .env
docker-compose up --build -d
docker-compose ps
curl http://127.0.0.1:3000/readyz
```

## Services

| Service | Purpose |
| --- | --- |
| `forge` | Main application process serving HTTP API, browser UI, smart HTTP Git, and SSH Git transport |
| `db` | PostgreSQL metadata store for users, sessions, repositories, SSH keys, orgs, collaborators, and webhooks |
| `redis` | Reserved for future queue or cache work |
| `mailpit` | Local SMTP sink for future notification flows during development |

## Environment variables

| Variable | Current role |
| --- | --- |
| `FORGE_ENV` | Runtime mode: `development`, `test`, or `production` |
| `FORGE_ADDR` | HTTP listen address |
| `FORGE_BASE_URL` | External base URL used for secure-cookie and validation logic |
| `FORGE_SECRET` | JWT signing secret |
| `FORGE_COOKIE_NAME` | Session cookie name |
| `FORGE_SESSION_TTL` | JWT and session lifetime |
| `FORGE_REQUEST_TIMEOUT` | Per-request timeout middleware |
| `FORGE_MAX_REQUEST_BODY_BYTES` | Maximum accepted HTTP request body |
| `FORGE_REPOS_ROOT` | Root directory for Git repository storage |
| `FORGE_SSH_ENABLED` | Enables SSH listener |
| `FORGE_SSH_ADDR` | SSH listen address |
| `FORGE_SSH_HOST_KEY_PATH` | SSH host key path, created automatically if missing |
| `FORGE_SSH_USER` | Expected SSH username, usually `git` |
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_URL` | Reserved for future Redis-backed integrations |

## Ports and persistence

- `3000` for the HTTP API, browser UI, and smart HTTP Git transport
- `2222` for SSH Git transport
- `5432` for PostgreSQL
- `6379` for Redis
- `1025` and `8025` for Mailpit

Data is persisted in named Docker volumes. The container entrypoint repairs ownership on mounted data paths before dropping privileges to the non-root runtime user.

## Operational notes

- Use `/readyz`, not only `/healthz`, when you want a meaningful readiness signal
- Embedded migrations run automatically on startup
- Repository maintenance is scheduled after create and write operations
- Production mode enforces stricter config validation, including HTTPS base URL requirements and a non-trivial secret

## Troubleshooting

| Symptom | Likely cause |
| --- | --- |
| Container restarts on boot | Inspect `docker-compose logs forge` for config or permission failures |
| Smart HTTP returns errors | Verify `git-http-backend` is available and the repo path exists under the configured storage root |
| SSH auth fails | Confirm the public key was registered and the SSH username is correct |
| `/readyz` fails | PostgreSQL may be unavailable or repository storage may not be writable |
