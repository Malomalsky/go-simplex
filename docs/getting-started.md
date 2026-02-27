# Getting Started

This guide shows the fastest way to launch a production-style SimpleX bot with this SDK.

## Prerequisites

- Go 1.23+
- SimpleX CLI with websocket API enabled
- network path from your bot host to SimpleX websocket endpoint

Official references:

- https://github.com/simplex-chat/simplex-chat/tree/stable/bots
- https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/COMMANDS.md

## 1. Run SimpleX CLI websocket API

```bash
simplex-chat -p 5225
```

For remote deployments, expose only `wss://` behind TLS and firewall rules.

## 2. Scaffold a new bot project

From this repository:

```bash
go run ./cmd/simplexbot-init \
  -module github.com/you/my-simplex-bot \
  -out ./my-simplex-bot \
  -ws ws://localhost:5225
```

Flags:

- `-module` (required): Go module path for generated bot project
- `-out`: output directory (`./simplex-bot` by default)
- `-ws`: websocket endpoint injected into generated `main.go`
- `-name`: optional display name for generated README
- `-sdk-module`: import path for this SDK module (`github.com/Malomalsky/go-simplex` by default)
- `-force`: overwrite existing generated files

## 3. Run generated bot

```bash
cd ./my-simplex-bot
go mod tidy
go run .
```

The scaffold includes:

- `/help` command with auto-generated command list
- `/ping` command
- `/echo <text>` command with quoted args support
- reconnect supervisor
- per-contact rate limiting

## 4. Manual setup (without scaffold)

```go
ctx := context.Background()
cli, err := client.NewWebSocketWithOptions(
    ctx,
    "ws://localhost:5225",
    []ws.Option{ws.WithReadLimit(16 << 20)},
    client.WithStrictResponses(false),
)
if err != nil {
    return err
}
defer cli.Close(context.Background())

boot, err := cli.BootstrapBot(ctx)
if err != nil {
    return err
}
_ = boot.Address
```

## 5. Recommended production toggles

Client:

- `client.WithStrictResponses(false)` for forward compatibility with new upstream `resp.type`
- `client.WithRawCommandMaxBytes(...)`
- `client.WithEventOverflowPolicy(...)`
- `client.WithDropHandler(...)`

Transport:

- `ws.WithRequireWSS(true)`
- `ws.WithTLSMinVersion(tls.VersionTLS12)` or higher
- `ws.WithReadLimit(...)` according to your message/file profile

Router/runtime:

- `router.EnablePerContactRateLimit(...)`
- `rt.OnError(...)` to centralize error reporting
- `bot.WithReconnectBackoff(...)` and `bot.WithReconnectMaxConsecutiveFailures(...)`

## 6. Validate your setup

```bash
go run ./cmd/simplex-smoke --ws ws://localhost:5225
go test ./...
```

If smoke fails, verify websocket URL, CLI process, and TLS mode (`ws://` vs `wss://`).
