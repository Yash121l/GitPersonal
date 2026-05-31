# Quickstart

Get Forge running locally, verify the service, create a repository, and complete one meaningful Git action end to end.

## Before you begin

Forge is easiest to start with Docker Compose, but it can also run directly with Go if you want a lighter local loop.

You should have:

- Git installed
- Docker and Docker Compose for the recommended path
- Go if you want to run the service outside Docker
- Free local ports for `3000` (HTTP) and `2222` (SSH)

!!! tip "Recommended path"
    Use the Docker flow for the first run. It gives you the closest shape to the intended local deployment and validates the containerized runtime at the same time.

## Run Forge

=== "Docker Compose"

    1. Copy the environment file:

        ```bash
        cp .env.example .env
        ```

    2. Start the stack:

        ```bash
        docker-compose up --build -d
        docker-compose ps
        ```

    3. Verify readiness:

        ```bash
        curl http://127.0.0.1:3000/readyz
        ```

=== "Local Go process"

    1. Install module dependencies:

        ```bash
        go mod tidy
        ```

    2. Run the service:

        ```bash
        go run ./cmd/forge
        ```

    3. Verify readiness:

        ```bash
        curl http://127.0.0.1:3000/readyz
        ```

## Open the app

Forge serves a browser UI at:

```text
http://127.0.0.1:3000/app
```

Use it to:

1. Register a user account
2. Create a repository owned by your user
3. Optionally register an SSH public key if you want SSH transport

## Push code

Pick one transport and complete a full push. HTTP is the simplest place to start.

<div class="path-grid">
  <div class="surface-card" markdown>
  <strong>HTTP path</strong>

  ```bash
  mkdir demo && cd demo
  git init
  echo "# Forge demo" > README.md
  git add README.md
  git commit -m "Initial commit"
  git branch -M main
  git remote add origin http://127.0.0.1:3000/git/<owner>/demo.git
  git push origin main
  ```
  </div>
  <div class="surface-card" markdown>
  <strong>SSH path</strong>

  ```bash
  git remote add origin ssh://git@127.0.0.1:2222/<owner>/demo.git
  git push origin main
  ```

  Use this only after the matching public key is registered in Forge.
  </div>
</div>

## Optional API-first flow

If you do not want to use the browser app, you can perform the same onboarding through the JSON API:

- Register with `POST /api/v1/auth/register`
- Create a repository with `POST /api/v1/repos`
- Register an SSH key with `POST /api/v1/keys`

See the full [API reference](reference/api.md) for request and response details.

## Where to go next

<div class="grid cards" markdown>

-   **Deployment**

    ---

    Review services, environment variables, ports, volumes, and troubleshooting.

    [Open deployment guide](guides/deployment.md)

-   **Architecture**

    ---

    Understand how repository storage, auth, smart HTTP, SSH, and maintenance fit together.

    [Open architecture reference](reference/architecture.md)

-   **Testing**

    ---

    See which claims are currently backed by transport-level and runtime verification.

    [Open testing guide](guides/testing.md)

</div>
