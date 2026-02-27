# Changelog

All notable changes to this project are documented in this file.

## [Unreleased]

## [v0.2.2] - 2026-02-27

### Added

- release workflow (`.github/workflows/release.yml`) for tag+GitHub Release from merged PR state
- changelog note extractor script: `scripts/release-notes.sh`
- production deployment and data safety guide: `docs/production.md`

### Changed

- `scripts/integration-local.sh` now supports:
  - `--no-start`
  - `--port <port>`
  - `--timeout <seconds>`
  - `--verbose`
  - explicit `--` passthrough for extra `go test` args
- docs expanded for harness options and release process (`README.md`, `CONTRIBUTING.md`, `integration/README.md`)

## [v0.2.1] - 2026-02-27

### Added

- local integration harness script:
  - starts `simplex-chat`
  - waits for websocket readiness
  - runs integration tests
  - cleans up process on exit
- documentation updates for local harness usage in README and integration docs

## [v0.2.0] - 2026-02-27

### Added

- concrete runnable example bots:
  - `examples/echo-bot`
  - `examples/faq-bot`
  - `examples/welcome-bot`
  - scenario guide in `examples/README.md`
- CI workflows:
  - main CI (`test`, `race`, `vet`, `vulncheck`, compatibility snapshot)
  - live contract workflow (`contract-live`) with optional fixture inputs/secrets
  - upstream snapshot sync workflow with auto-PR
- compatibility guard script: `scripts/check-compat.sh`
- live integration contract test suite (`integration` build tag), including:
  - bootstrap and typed sender parity
  - list contacts/groups
  - optional contact/group message flows (send/update/delete/reaction)
  - optional file command flow (receive/cancel)
- repository governance and maintenance files:
  - `CONTRIBUTING.md`
  - `SECURITY.md`
  - `CODEOWNERS`
  - issue/PR templates
  - `dependabot.yml`

### Changed

- CI `vulncheck` now runs on stable Go toolchain to avoid false failures from outdated runner stdlib.
- README/docs expanded with compatibility, security, and live-contract execution instructions.

## [v0.1.0] - 2026-02-27

### Added

- initial production-oriented SDK baseline:
  - typed API helpers
  - reconnect runtime and command router
  - rate limiting and safety validations
  - scaffold command with `basic` and `moderation` templates

[v0.2.2]: https://github.com/Malomalsky/go-simplex/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/Malomalsky/go-simplex/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/Malomalsky/go-simplex/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/Malomalsky/go-simplex/releases/tag/v0.1.0
[Unreleased]: https://github.com/Malomalsky/go-simplex/compare/v0.2.2...HEAD
