.PHONY: generate
generate:
	@mockery

.PHONY: build-frontend
build-frontend:
	@cd frontend && npm ci && npm run build

.PHONY: build
build: generate
	@go build -ldflags="-s -w" -o bin/tron-validator-watcher cmd/watcher/main.go

.PHONY: build-all
build-all: generate build-frontend build

.PHONY: run
run: generate
	@go run cmd/watcher/main.go --config-file config.yaml

.PHONY: tests
tests:
	@go test -v ./... -count=1

.PHONY: coverage
coverage:
	@echo "Generating coverage report..."
	@go test -race -coverprofile=coverage/coverage.out.tmp ./... -count=1
	@cat coverage/coverage.out.tmp | grep -v "mocks" > coverage/coverage.out
	@go tool cover -html=coverage/coverage.out
	@rm -rf coverage/*

.PHONY: lint
lint:
	@golangci-lint run ./...
