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
