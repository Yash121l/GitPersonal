FROM golang:1.23-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/forge ./cmd/forge

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates git \
	&& addgroup -S forge \
	&& adduser -S -G forge -h /app forge \
	&& mkdir -p /data/repos /app \
	&& chown -R forge:forge /app /data

COPY --from=build /out/forge /usr/local/bin/forge

USER forge

EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/forge"]
