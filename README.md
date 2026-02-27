# go-simplex

Go SDK for building SimpleX bots with an idiomatic, typed API and a practical bot runtime.

## Scope

- full generated coverage of the official SimpleX bot command API snapshot (`spec/upstream/commands.ts`)
- typed event/response tags and records from upstream snapshots
- high-level client helpers for common bot workflows
- bot runtime with middleware, direct-message extraction, command router, and reconnect supervisor
- security-focused defaults and explicit hardening options

## Official SimpleX docs

- Bot overview: https://github.com/simplex-chat/simplex-chat/tree/stable/bots
- Bot API commands: https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/COMMANDS.md
- Bot API events: https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/EVENTS.md
- Bot API types: https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/TYPES.md
- Official TypeScript SDK: https://github.com/simplex-chat/simplex-chat/blob/stable/packages/simplex-chat-client/typescript/README.md

## Quick start

### 1. Start SimpleX CLI websocket API

```bash
simplex-chat -p 5225
```

### 2. Generate a new bot project

```bash
go run ./cmd/simplexbot-init \
  -module github.com/you/my-simplex-bot \
  -template basic \
  -out ./my-simplex-bot
cd ./my-simplex-bot
go mod tidy
go run .
```

### 3. Or run the included example

```bash
go run ./examples/echo
go run ./examples/moderation
```

Available scaffold templates:

- `basic`: `help`, `ping`, `echo`
- `moderation`: deny-list moderation (`addword`, `delword`, `words`) for direct messages

## Minimal bot example

```go
ctx := context.Background()
router := bot.NewTextRouter()
_ = router.EnablePerContactRateLimit(20, time.Minute)
_ = router.OnWithDescription("ping", "health check", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
    return cmd.Reply(ctx, cli, "pong")
})

err := bot.RunWebSocketWithReconnect(
    ctx,
    "ws://localhost:5225",
    nil,
    []client.Option{client.WithStrictResponses(false)},
    func(cli *client.Client) (bot.Runner, error) {
        rt, err := bot.NewRuntime(cli)
        if err != nil {
            return nil, err
        }
        bot.OnDirectCommands(rt, router)
        return rt, nil
    },
)
if err != nil && !errors.Is(err, context.Canceled) {
    panic(err)
}
```

## API layers

- generated command layer: `sdk/command/generated_requests.go`
- generated typed sender layer: `sdk/client/generated_senders.go`
- high-level convenience client: `sdk/client/api.go`
- runtime/router layer: `sdk/bot/*`

## Security and resilience features

- websocket hardening:
  - `ws.WithRequireWSS(true)`
  - `ws.WithTLSMinVersion(...)`
  - `ws.WithReadLimit(...)`
- raw command controls:
  - `client.WithRawCommandAllowPrefixes(...)`
  - `client.WithRawCommandValidator(...)`
  - `client.WithRawCommandMaxBytes(...)`
- forward-compatible unknown response handling:
  - `client.WithStrictResponses(false)`
- bounded channel overflow policies:
  - `client.WithEventOverflowPolicy(...)`
  - `client.WithErrorOverflowPolicy(...)`
  - `client.WithDropHandler(...)`
- ref validation in high-level helpers (`sendRef`/`chatRef` must be `@id`, `#id`, `*id`)
- per-contact command rate limiting in router:
  - `router.EnablePerContactRateLimit(...)`
  - `router.OnRateLimited(...)`

More details: `docs/security.md`.

## Documentation

- Getting started: `docs/getting-started.md`
- Bot development guide: `docs/bot-development.md`
- Compatibility and coverage: `docs/compatibility.md`
- Security guide: `docs/security.md`
- Upstream API research notes: `docs/research/upstream-api.md`
- Upstream TS SDK research notes: `docs/research/upstream-sdk.md`
- Implementation roadmap: `docs/plan/go-sdk-roadmap.md`

## Compatibility snapshot

Current generated coverage against tracked upstream snapshots:

- bot API command interfaces in `spec/upstream/commands.ts`: `42`
- generated request structs in `sdk/command/generated_requests.go`: `42`
- generated typed sender methods in `sdk/client/generated_senders.go`: `42`
- upstream event tags in `spec/upstream/events.ts`: `45`
- generated `types.EventType` constants: `45`
- upstream response tags in `spec/upstream/responses.ts`: `45`
- generated `types.ResponseType` constants: `45`

Details and verification commands are in `docs/compatibility.md`.

## Development

Regenerate from local upstream snapshots:

```bash
go run ./cmd/simplexgen
```

Refresh upstream snapshots and regenerate:

```bash
./scripts/update-upstream.sh
```

Run checks:

```bash
go test ./...
go test -race ./...
go vet ./...
```

Optional vulnerability scan:

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```
