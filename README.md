# go-simplex

Go SDK for building bots on top of the official SimpleX Bot API.

Current stage: research and architecture planning.

## Documents

- API research: `docs/research/upstream-api.md`
- TypeScript SDK research: `docs/research/upstream-sdk.md`
- Implementation roadmap: `docs/plan/go-sdk-roadmap.md`

## Principles

- full parity with documented SimpleX bot API
- idiomatic Go API
- generated contracts to prevent upstream drift
- practical bot-developer ergonomics

## Development

Generate command catalog from upstream snapshot:

```bash
go run ./cmd/simplexgen
```

Run tests:

```bash
go test ./...
```

## Quickstart (current API)

Run SimpleX CLI with websocket API:

```bash
simplex-chat -p 5225
```

Create client and bot runtime:

```go
ctx := context.Background()
cli, err := client.NewWebSocket(ctx, "ws://localhost:5225")
if err != nil {
    panic(err)
}
defer cli.Close(ctx)

rt, err := bot.NewRuntime(cli)
if err != nil {
    panic(err)
}

rt.On("newChatItems", func(ctx context.Context, cli *client.Client, msg protocol.Message) error {
    // handle event
    return nil
})

if err := rt.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
    panic(err)
}
```
