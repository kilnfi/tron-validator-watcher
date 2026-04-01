# Contributing

Contributions are welcome! Here's how to get started.

## Local setup

**Prerequisites:** [mise](https://mise.jdx.dev)

```sh
git clone https://github.com/kilnfi/tron-validator-watcher.git
cd tron-validator-watcher
mise install                 # installs Go, Node, golangci-lint, mockery, task
cp config.example.yaml config.yaml
# Edit config.yaml with your RPC endpoint and validator addresses
task setup                   # downloads deps, generates mocks, builds frontend
```

Run directly:

```sh
task run
```

## Running tests

```sh
task test
```

## Linting

```sh
task lint
```

## Regenerating mocks

```sh
task generate
```

## Project structure

```
cmd/watcher/         # Entrypoint
internal/
  tron/              # Tron API client (blocks, witnesses, accounts)
  watcher/block/     # Block-by-block monitoring logic
  watcher/network/   # Network-level metrics
  status/            # In-memory state store
  ui/                # Embedded frontend (React/Vite)
  server/http/       # HTTP server (/metrics, /api/status, dashboard)
frontend/            # React + TypeScript source
```

## Submitting a PR

1. Fork the repo and create a branch from `main`
2. Make your changes with tests
3. Ensure `make tests` and `make lint` pass
4. Open a PR against `main`

## Reporting a vulnerability

Please do not open a public issue for security vulnerabilities. See [SECURITY.md](SECURITY.md).
