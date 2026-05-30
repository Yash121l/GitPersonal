APP_NAME := forge

.PHONY: run build test fmt tidy

run:
	go run ./cmd/forge

build:
	go build -o ./bin/$(APP_NAME) ./cmd/forge

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

