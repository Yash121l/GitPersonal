# Architecture

Forge is a Git transport and metadata system built from explicit subsystems. Git repositories live on disk. PostgreSQL stores metadata and revocable sessions. HTTP and SSH transports map authenticated users to real Git server commands.

## System shape

```text
Browser / Git Client / SSH Client
          |
          v
   +-------------------+
   |   Forge Process   |
   |-------------------|
   | HTTP API          |
   | Smart HTTP Git    |
   | SSH Git Server    |
   | Repo Service      |
   | Maintenance Queue |
   +-------------------+
      |            |
      v            v
 PostgreSQL      /data/repos
 users           sharded bare repositories
 sessions        staging + trash paths
 ssh keys
```

## Storage model

Forge uses bare repositories on disk. The main scaling and correctness work happens around that primitive:

- Sharded path layout
- Atomic provisioning in a staging area
- Safe deletion through a trash path
- Explicit path-safety checks
- Background maintenance instead of inline Git housekeeping during user traffic

### Repository layout

```text
/data/repos/
  bf/35/yash/forge.git
  9a/10/team/demo.git
```

### Provisioning strategy

1. Create the repository in a staging area
2. Run `git init --bare --shared=group`
3. Write the `HEAD` reference for the configured default branch
4. Apply repository configuration for transport and maintenance behavior
5. Move the repository into its final path atomically

## Authentication and sessions

Forge uses bcrypt for passwords and JWT cookies for browser sessions, but each token also has a persisted session record keyed by token ID. Logout revokes the stored session, so revocation is real rather than client-side only.

Persisted session-aware auth matters because it lets the server invalidate a token before expiry.

## Smart HTTP Git transport

Git HTTP traffic enters through `/git/{owner}/{repo}.git`. Forge:

1. Parses the repository path
2. Loads repository metadata
3. Determines whether the request is read or write oriented
4. Authenticates using a session cookie or HTTP Basic credentials
5. Applies repository authorization
6. Invokes `git-http-backend`
7. Relays the response back to the Git client
8. Enqueues repository maintenance after successful writes

## SSH Git transport

Forge also exposes Git over SSH on a dedicated listener. The SSH server:

- Authenticates public keys against stored fingerprints
- Maps the key back to a user
- Accepts only `git-upload-pack` and `git-receive-pack`
- Applies the same access rules used by smart HTTP
- Enqueues maintenance work after successful writes

## Maintenance

Maintenance is intentionally backgrounded so user-facing requests do not block on Git housekeeping.

Current tasks include:

- `git gc --auto`
- `git commit-graph write --reachable`
- Repository size accounting and maintenance timestamps

## Runtime boot model

Startup is fail-fast:

1. Load and validate config
2. Open the selected store backend
3. Apply embedded PostgreSQL migrations when relevant
4. Initialize repository storage and the maintenance worker
5. Start HTTP and SSH listeners
6. Wait for shutdown signals and stop gracefully

## Current limits

- Authorization is still relatively simple compared with a full Git hosting product
- Multi-node coordination is not implemented
- The PostgreSQL store does not yet implement a true cross-process repository lease
- Rate limiting, CSRF protection, and audit logging remain future hardening work
