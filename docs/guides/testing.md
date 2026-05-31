# Testing

Forge is moving toward test-driven development with transport-level realism. The valuable tests are the ones that prove the system behaves like a Git host, not just that handlers return the expected JSON shape.

## Current automated coverage

| Area | What is covered now |
| --- | --- |
| Config | Validation behavior, especially unsafe production defaults |
| Repository service | Provisioning, deletion, sharded paths, and maintenance stat updates |
| HTTP server | Auth lifecycle, repository CRUD, headers, readiness, and unauthorized access |
| Transport integration | Logout revocation, smart HTTP push and read, and SSH access with a registered key |

## Why the transport tests matter

The highest-value tests currently verify that:

- A user can authenticate and lose access after logout because the session is actually revoked
- A repository can be created, pushed to over smart HTTP, and read back through Git commands
- An SSH public key can be registered and then used to access the same repository over SSH
- Repository maintenance updates metadata after work is queued

## Useful commands

```bash
go test ./...

docker-compose up --build -d
docker-compose ps
docker-compose logs --no-color forge

curl http://127.0.0.1:3000/readyz

git ls-remote http://user:password@127.0.0.1:3000/git/user/repo.git
GIT_SSH_COMMAND='ssh -i ~/.ssh/id_rsa -o IdentitiesOnly=yes -o StrictHostKeyChecking=no' \
  git ls-remote ssh://git@127.0.0.1:2222/user/repo.git
```

## Next test targets

- Authorization matrix coverage once collaborator and team permissions deepen
- Rate limiting and CSRF tests when those layers are added
- Webhook-delivery worker tests once persistent jobs exist
- PostgreSQL-backed transport tests beyond the in-memory store
- Migration-compatibility tests as the schema grows
