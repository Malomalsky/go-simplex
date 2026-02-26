# Go SDK Roadmap: SimpleX Bot API Full Implementation

Date: 2026-02-26
Status: planning

## Goal

Build an idiomatic Go SDK that fully covers the documented SimpleX bot API and makes it easy to build production bots.

## What "full implementation" means for this project

1. Protocol parity with official documented bot API.
- All 42 documented commands represented in Go.
- All documented response/event tags represented in Go.
- Generated contracts synchronized with upstream docs/types.

2. Runtime parity.
- WebSocket envelope handling compatible with upstream CLI.
- Correlated command/response flow via `corrId`.
- Async event stream handling with unknown-tag tolerance.

3. Developer ergonomics.
- Minimal code to build bots.
- Typed helpers for common flows (bootstrap, messaging, groups, files).
- Predictable error model and cancellation semantics.

## Non-goals for initial release

- Implementing undocumented/internal CLI commands by default.
- Replacing SimpleX CLI process itself.
- Cross-language code generation in v1 (Go only).

## Constraints and risks

1. Upstream drift.
- Docs and generated types can change between stable releases.

2. Event/response overlap and unknown tags.
- Decoder design must avoid brittle assumptions.

3. Integration environment complexity.
- Integration tests need a controllable `simplex-chat` CLI setup.

4. Licensing.
- Upstream repo is AGPL-3.0. We need to confirm distribution implications for generated artifacts and copied logic before public release policy is finalized.

## Target package layout

```text
/cmd/simplexgen/              # code generator from upstream artifacts
/internal/upstreamfetch/      # optional helpers for pulling stable artifacts
/sdk/types/                   # generated API types (commands, responses, events, shared)
/sdk/command/                 # typed command builders + raw command path
/sdk/protocol/                # envelopes, corrId, wire decode/encode
/sdk/transport/ws/            # websocket transport
/sdk/client/                  # low-level client (send cmd, receive events)
/sdk/bot/                     # high-level bot runtime and router
/sdk/errors/                  # typed errors
/examples/echo/
/examples/menu-bot/
/examples/moderation/
```

## Public API sketch (Go idioms)

```go
type Client interface {
    Send(ctx context.Context, cmd command.Command) (types.Response, error)
    SendRaw(ctx context.Context, cmd string) (types.Response, error)
    Events() <-chan protocol.EventEnvelope
    Close(ctx context.Context) error
}
```

```go
type Bot interface {
    Run(ctx context.Context) error
}
```

Design rules:

- `context.Context` everywhere for networked operations.
- Functional options for config (`WithTimeout`, `WithLogger`, `WithBackoff`).
- `errors.Is/As` friendly typed errors.
- Stable zero-allocation hot paths only where it matters (after correctness).

## Code generation strategy

Preferred source inputs for generation:

1. `bots/api/COMMANDS.md`
2. `bots/api/EVENTS.md`
3. `bots/api/TYPES.md`

Secondary reference (validation target):

- `packages/simplex-chat-client/types/typescript/src/*.ts`

Generator outputs:

- `sdk/types/generated_types.go`
- `sdk/types/generated_responses.go`
- `sdk/types/generated_events.go`
- `sdk/command/generated_commands.go`
- `sdk/command/generated_cmdstring_test.go` (parity fixtures)

Why this approach:

- Avoid manual drift.
- Keep parity testable in CI.
- Make updates to new upstream stable versions mechanical.

## Milestones

### M0 - Repository bootstrap

Deliverables:

- repo initialized, baseline docs, `.gitignore`.
- contribution + versioning policy draft.

Acceptance:

- clean `git status`.

### M1 - Generator foundation

Deliverables:

- `cmd/simplexgen` parses upstream markdown and emits Go AST/templates.
- deterministic outputs.
- golden tests for parser and emitted command strings.

Acceptance:

- generator roundtrip reproducible.
- command count parity check (`42`).

### M2 - Core protocol and transport

Deliverables:

- wire envelope structs (`corrId/cmd`, response/event envelope).
- websocket transport with read/write loops and graceful shutdown.
- pending request registry with correlation.

Acceptance:

- unit tests for:
  - response routing,
  - unknown corrId behavior,
  - timeout/cancel behavior.

### M3 - Typed client API

Deliverables:

- low-level `Client` with typed `Send`.
- raw command escape hatch.
- typed errors.

Acceptance:

- all generated command wrappers compile and run against mocked transport.

### M4 - Bot framework layer

Deliverables:

- event router/dispatcher:
  - by tag,
  - fallback unknown handler.
- middleware hooks (logging, panic recovery, metrics).

Acceptance:

- example echo bot under 40 lines of business logic.

### M5 - Integration tests with SimpleX CLI

Deliverables:

- reproducible integration harness.
- smoke scenarios:
  - active user,
  - address create/show,
  - contact connect,
  - send/receive message,
  - group basics.

Acceptance:

- CI or documented local command for full integration suite.

### M6 - Documentation and release

Deliverables:

- README with quickstart and architecture.
- cookbook examples for common bot patterns.
- semantic versioning policy and changelog template.

Acceptance:

- user can build first bot from docs without reading internals.

## Testing strategy (TDD-oriented)

1. Parser/generator tests first.
- golden files for parsed commands/events/types.
- parity checks against known upstream fixtures.

2. Protocol tests second.
- envelope decode/encode.
- concurrent send/receive.
- shutdown edge cases.

3. Client behavior tests third.
- command timeout, context cancellation.
- unknown tag handling.
- unexpected frame resilience.

4. Integration tests last.
- non-deterministic event order handled via predicate waits.

## Commit strategy

Small, reviewable commits:

1. `chore: bootstrap repository and planning docs`
2. `feat(gen): add upstream markdown parser`
3. `feat(types): generate API contracts`
4. `feat(client): add websocket protocol client`
5. `feat(bot): add event router and bot runtime`
6. `test(integration): add simplex cli smoke tests`
7. `docs: quickstart and bot cookbook`

## Definition of done for v0.1.0

- Generated typed coverage for all documented commands/responses/events.
- Stable low-level client API with context-aware request lifecycle.
- One production-grade bot example and one moderation-oriented example.
- Passing unit tests and documented integration test flow.

## Immediate next implementation step

Start M1:

1. Initialize Go module.
2. Implement markdown parser for `COMMANDS.md` first.
3. Generate command builders + cmd string parity tests.

