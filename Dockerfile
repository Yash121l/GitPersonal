FROM node:22-alpine AS frontend-build

WORKDIR /src/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN mkdir -p /src/internal/server/ui && npm run build

FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend-build /src/internal/server/ui/dist ./internal/server/ui/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/forge ./cmd/forge

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates curl git gosu \
	&& rm -rf /var/lib/apt/lists/* \
	&& groupadd --system forge \
	&& useradd --system --gid forge --home-dir /app --create-home forge \
	&& mkdir -p /data/repos /app \
	&& chown -R forge:forge /app /data

COPY --from=build /out/forge /usr/local/bin/forge
COPY deploy/entrypoint.sh /usr/local/bin/forge-entrypoint
RUN chmod +x /usr/local/bin/forge-entrypoint

EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/forge-entrypoint"]
