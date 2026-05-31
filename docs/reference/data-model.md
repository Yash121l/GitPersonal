# Data Model

Forge splits state deliberately between PostgreSQL metadata and Git-native filesystem storage.

## Storage boundaries

| Layer | What lives there |
| --- | --- |
| PostgreSQL | Users, sessions, repositories, SSH keys, organizations, collaborators, webhooks, and maintenance metadata |
| `/data/repos` | Sharded bare Git repositories, staging directories, and trash directories |
| `/data/ssh` | SSH host key material for the Forge SSH listener |
| In-memory store | Test and no-database fallback only |

## Current schema shape

| Table | Purpose |
| --- | --- |
| `users` | Core identity rows with username, password hash, and role |
| `sessions` | Revocable session rows keyed by token identifier |
| `repositories` | Repository metadata, ownership, visibility, branch, path, size, and maintenance timestamps |
| `ssh_keys` | Public keys with fingerprint uniqueness and `last_used_at` tracking |
| `organizations` | Organization ownership model |
| `org_members` | Organization membership and role mapping |

## Repository invariants

- Every repository row belongs to exactly one owner type
- Repository names are unique per owner case-insensitively
- `repo_path` stores the actual filesystem path used by Git commands
- The current create flow provisions user-owned repositories while the schema already supports organizations

Filesystem fanout is derived from a SHA-256 of `owner/name`, producing prefix directories before the owner and repository slug.

## Store interface

The `internal/store` package defines a persistence contract for:

- User creation and lookup
- Repository create, lookup, list, delete, and stat updates
- Session creation, lookup, and revocation
- SSH key creation, lookup, and last-used updates
- Repository mutation serialization through `WithRepositoryLease`
- Readiness checks through `Check`

## Store implementations

| Implementation | What it does today |
| --- | --- |
| `internal/store/memory` | Uses in-memory maps and mutexes, including per-repo locking, and is ideal for tests |
| `internal/store/postgres` | Persists the real deployment data in PostgreSQL, though repository leasing is still incomplete |

## Current gaps

- Organization support is present in the schema but still growing at the product layer
- Authorization rules remain simple compared with a mature Git host
- Cross-process repository mutation serialization is not complete in PostgreSQL
- Audit, issue, pull-request, and CI-oriented tables are still future work
