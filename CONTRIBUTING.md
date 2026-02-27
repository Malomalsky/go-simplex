# Contributing

Thanks for contributing to `go-simplex`.

## Scope

This repository aims to provide a production-ready, idiomatic Go SDK with high parity to the official SimpleX bot API.

Before implementing features, prefer compatibility and safety over API surface growth.

## Development setup

```bash
go test ./...
```

Optional full checks:

```bash
./scripts/check-compat.sh
go test -race ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## Upstream snapshot workflow

Refresh upstream files and regenerate SDK artifacts:

```bash
./scripts/update-upstream.sh
```

Compatibility gate:

```bash
./scripts/check-compat.sh
```

Do not hand-edit generated files:

- `sdk/command/generated_catalog.go`
- `sdk/command/generated_requests.go`
- `sdk/client/generated_senders.go`
- `sdk/types/generated_tags.go`
- `sdk/types/generated_records.go`
- `sdk/types/generated_types.go`

## Integration tests

Live contract tests use build tag `integration`.

```bash
SIMPLEX_WS_URL=ws://localhost:5225 go test -tags=integration ./integration/... -v
```

For local reproducible runs, use:

```bash
./scripts/integration-local.sh
```

Example with custom options:

```bash
./scripts/integration-local.sh --no-start --port 5225 --timeout 90 -- -run TestLiveContracts
```

Additional fixture variables are documented in `integration/README.md`.

## Release process

Releases are done from merged PR state:

1. Add/update changelog section in `CHANGELOG.md` (`## [vX.Y.Z] - YYYY-MM-DD`).
2. Merge PR to `main`.
3. Run GitHub Actions workflow `Release` with input `version=vX.Y.Z`.

The workflow validates the tag, extracts notes from `CHANGELOG.md`, creates/pushes the tag, and publishes GitHub Release.

## Pull requests

- keep PRs focused (one concern per PR)
- include tests for behavior changes
- include docs updates for user-facing changes
- avoid breaking exported API unless explicitly discussed

## Commit style

Repository currently uses conventional style where practical, for example:

- `feat(client): ...`
- `fix(ci): ...`
- `docs(...): ...`

## Questions

Use GitHub issues for design and roadmap discussions before large changes.
