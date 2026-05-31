# API Reference

Forge exposes JSON endpoints plus real Git transport over HTTP and SSH.

## Surface area

| Interface | Entry point | Current role |
| --- | --- | --- |
| Health | `/healthz`, `/readyz` | Liveness and readiness checks |
| JSON API | `/api/v1/*` | Auth lifecycle, current user, orgs, SSH keys, repos, collaborators, and webhooks |
| Git smart HTTP | `/git/{owner}/{repo}.git` | Clone, fetch, push, and ref advertisement through `git-http-backend` |
| Git over SSH | `ssh://git@host:2222/{owner}/{repo}.git` | Clone, fetch, and push through `git-upload-pack` and `git-receive-pack` |

## Common conventions

- JSON endpoints return `application/json`
- JSON decoding is strict and rejects unknown fields
- Request IDs are returned in `X-Request-Id`
- Request contexts are bounded by `FORGE_REQUEST_TIMEOUT`
- Request bodies are bounded by `FORGE_MAX_REQUEST_BODY_BYTES`

## Authentication

HTTP auth modes:

- Session cookie for browser-oriented auth
- HTTP Basic auth for Git clients and scripts

SSH auth mode:

- Public-key authentication using stored key fingerprints

Session behavior:

- Passwords are hashed with bcrypt
- Successful auth creates a JWT plus a persisted session row
- Logout revokes the stored session and expires the cookie
- The cookie is `HttpOnly`, `SameSite=Lax`, and `Secure` when HTTPS is configured

## Validation rules

| Field | Rule |
| --- | --- |
| `username` | Registration accepts 3 to 39 characters matching `^[a-zA-Z0-9._-]+$` |
| `password` | Registration requires at least 12 characters |
| `repository.name` | Required, 1 to 100 characters, matching `^[a-zA-Z0-9._-]+$` |
| `repository.visibility` | `public` or `private`, defaulting to `private` |
| `repository.default_branch` | Defaults to `main` when empty |
| `ssh key public_key` | Must parse as an OpenSSH authorized key line |

## Route reference

### Health and auth

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/healthz` | Liveness check |
| `GET` | `/readyz` | Dependency-aware readiness check |
| `POST` | `/api/v1/auth/register` | Register a user and start a session |
| `POST` | `/api/v1/auth/login` | Authenticate an existing user |
| `POST` | `/api/v1/auth/logout` | Revoke the current session |
| `GET` | `/api/v1/me` | Return the authenticated user |

Example register request:

```json
{
  "username": "yash",
  "password": "correct horse battery staple"
}
```

### SSH keys

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/keys` | List registered public keys |
| `POST` | `/api/v1/keys` | Add a public key |

### Organizations

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/orgs` | List organizations visible to the user |
| `POST` | `/api/v1/orgs` | Create an organization |
| `POST` | `/api/v1/orgs/{org}/members` | Add or update an organization member |

### Repositories

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/repos` | List repositories |
| `GET` | `/api/v1/repos/{owner}/{repo}` | Get repository details |
| `POST` | `/api/v1/repos` | Create a repository |
| `DELETE` | `/api/v1/repos/{owner}/{repo}` | Delete a repository |
| `POST` | `/api/v1/repos/{owner}/{repo}/collaborators` | Add or update a collaborator |
| `GET` | `/api/v1/repos/{owner}/{repo}/webhooks` | List webhooks |
| `POST` | `/api/v1/repos/{owner}/{repo}/webhooks` | Create a webhook |
| `DELETE` | `/api/v1/repos/{owner}/{repo}/webhooks/{webhookID}` | Delete a webhook |

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

## Git transport

### Smart HTTP

Use:

```text
http://host:3000/git/{owner}/{repo}.git
```

### SSH

Use:

```text
ssh://git@host:2222/{owner}/{repo}.git
```

## Error model

Error responses are JSON. Common classes include:

- Validation failures
- Authentication failures
- Authorization failures
- Resource conflicts
- Dependency-readiness failures
- Internal server errors
