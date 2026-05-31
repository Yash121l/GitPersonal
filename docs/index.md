---
hide:
  - toc
---

<div class="hero" markdown>

# Run Forge locally, push your first repository, then read the system from the inside out.

Forge is a self-hosted Git learning project with real smart HTTP and SSH transport, PostgreSQL-backed metadata, revocable sessions, repository maintenance, and a browser UI. This documentation is now organized as a proper docs site: quickstart first, then guides and deeper reference material.

<div class="hero__actions" markdown>

[Open the quickstart](quickstart.md){ .md-button .md-button--primary }
[Browse the guides](guides/index.md){ .md-button }

</div>

<div class="hero__meta">
  <div class="hero__meta-card">
    <strong>10 minutes</strong>
    <span>Recommended path with Docker Compose</span>
  </div>
  <div class="hero__meta-card">
    <strong>2 transport modes</strong>
    <span>Git over smart HTTP and SSH</span>
  </div>
  <div class="hero__meta-card">
    <strong>Reference included</strong>
    <span>Architecture, API, data model, deployment, and testing</span>
  </div>
</div>

</div>

## Start with the path you need

<div class="surface-grid">
  <div class="surface-card">
    <strong>Quickstart</strong>
    <span>Bring the stack up, verify readiness, create a user, create a repo, and push code.</span>
  </div>
  <div class="surface-card">
    <strong>Guides</strong>
    <span>Read the setup, deployment, and testing flows without digging through implementation detail first.</span>
  </div>
  <div class="surface-card">
    <strong>Reference</strong>
    <span>Use the architecture, API, and data model pages when you need exact runtime behavior.</span>
  </div>
</div>

<div class="grid cards" markdown>

-   **Quickstart**

    ---

    The fastest route from clone to a working Forge instance and a successful first push.

    [Open quickstart](quickstart.md)

-   **Guides**

    ---

    Deployment notes, testing posture, and the reading map for the project.

    [Open guides](guides/index.md)

-   **Reference**

    ---

    Architecture, API surface, storage boundaries, and current invariants.

    [Open reference](reference/architecture.md)

</div>

## Current scope

Implemented now:

- HTTP server with graceful shutdown, readiness checks, request IDs, and baseline security headers
- JWT cookie auth with persisted sessions and logout revocation
- PostgreSQL-backed metadata with in-memory fallback for tests and no-DB runs
- Repository CRUD, sharded bare-repository provisioning, and background maintenance
- Smart HTTP Git transport through `git-http-backend`
- SSH Git transport with registered public keys
- Browser UI at `/app` for sign-in, repositories, organizations, SSH keys, collaborators, and webhooks

Still missing:

- Pull requests, issues, code review flows, releases, and CI runners
- Richer admin workflows, audit logging, and broader product depth
- Persistent worker queues for retries and future automation
- Git LFS, search, and wider hardening work

!!! tip "Recommended sequence"
    Run the [quickstart](quickstart.md) first. Once the system is live, use the [deployment guide](guides/deployment.md) for operations details and the [architecture reference](reference/architecture.md) when you want to understand how the pieces fit together.
