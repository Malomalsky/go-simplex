# Security and Data Safety

This guide covers practical controls for running bots safely in production.

## Threat model focus

- untrusted inbound payloads from network peers
- protocol drift from upstream SimpleX versions
- resource exhaustion (large payloads, event bursts)
- accidental disclosure of personal chat data in logs/storage

## Transport security

For remote deployments, require encrypted websocket transport.

```go
cli, err := client.NewWebSocketWithOptions(
    ctx,
    "wss://bot.example/ws",
    []ws.Option{
        ws.WithRequireWSS(true),
        ws.WithTLSMinVersion(tls.VersionTLS12),
        ws.WithReadLimit(16 << 20),
    },
)
```

Recommendations:

- use `wss://` for all non-local traffic
- terminate TLS with modern ciphers and cert rotation
- avoid exposing websocket endpoint directly to the public internet

## Input hardening

### Raw command controls

If you use `SendRaw`, always constrain command surface.

```go
cli, err := client.New(
    transport,
    client.WithRawCommandAllowPrefixes("/_", "/api"),
    client.WithRawCommandMaxBytes(1<<20),
)
```

For stronger policy, use `client.WithRawCommandValidator(...)`.

### Reference format validation

High-level methods validate `sendRef`/`chatRef` format before sending commands.

Allowed forms:

- `@<contactId>`
- `#<groupId>`
- `*<contactConnectionId>`

This blocks malformed refs early and avoids unsafe command construction.

## Resilience under load

### Event/error channel overflow

Set explicit overflow policy and monitor drops.

```go
cli, err := client.New(
    transport,
    client.WithEventOverflowPolicy(client.OverflowPolicyDropNewest),
    client.WithErrorOverflowPolicy(client.OverflowPolicyDropNewest),
    client.WithDropHandler(func(kind string, dropped uint64) {
        log.Printf("dropped %s: %d", kind, dropped)
    }),
)
```

### Per-contact command throttling

```go
_ = router.EnablePerContactRateLimit(20, time.Minute)
router.OnRateLimited(func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
    return cmd.Reply(ctx, cli, "rate limit exceeded")
})
```

## Forward compatibility vs strictness

`client.WithStrictResponses(false)` lets bots continue working if upstream adds new response tags.

- use `false` in production to reduce breakage risk during upstream upgrades
- keep regression tests that assert required response types for critical flows

## Data safety practices

- do not log full inbound/outbound payloads in production
- redact contact links, user addresses, and message text from structured logs
- encrypt persistent bot data at rest
- keep `.env` and key material outside repository
- rotate credentials and access tokens regularly

## File handling guidance

When receiving files:

- validate and sanitize destination paths
- avoid writing to privileged directories
- enforce per-file size limits at infrastructure level
- scan files before downstream processing in sensitive environments

## Update hygiene

- refresh upstream snapshots regularly with `./scripts/update-upstream.sh`
- rerun checks after each refresh:

```bash
go test ./...
go test -race ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## Official references

- API commands: https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/COMMANDS.md
- API events: https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/EVENTS.md
- API types: https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/TYPES.md
