# Compatibility and Coverage

This document tracks compatibility with the official SimpleX bot API snapshot stored in this repository.

## Upstream sources used for generation

Local snapshots:

- `spec/upstream/commands.ts`
- `spec/upstream/events.ts`
- `spec/upstream/responses.ts`
- `spec/upstream/COMMANDS.md`

Official upstream references:

- https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/COMMANDS.md
- https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/EVENTS.md
- https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/TYPES.md
- https://github.com/simplex-chat/simplex-chat/blob/stable/packages/simplex-chat-client/typescript/README.md

## Current coverage snapshot

Command layer:

- upstream API command interfaces (`commands.ts`): `42`
- generated Go request structs (`sdk/command/generated_requests.go`): `42`
- generated typed sender methods (`sdk/client/generated_senders.go`): `42`

Event layer:

- upstream event tags (`events.ts` / `CEvt.Tag`): `45`
- generated Go event tags (`sdk/types/generated_tags.go` / `EventType`): `45`

Response layer:

- upstream response tags (`responses.ts` / `CR.Tag`): `45`
- generated Go response tags (`sdk/types/generated_tags.go` / `ResponseType`): `45`

Interpretation:

- generated low-level command API currently has full parity with the tracked upstream bot command snapshot
- generated event/response type tags currently have full parity with tracked upstream tags

## What "full parity" means here

- every command interface in `spec/upstream/commands.ts` has a generated Go request struct
- every command interface has a generated typed `Send...` helper on `*client.Client`
- every upstream event/response tag in tracked snapshots has generated Go tag constants

## What is intentionally separate from parity

- high-level convenience helpers in `sdk/client/api.go` are opinionated ergonomics on top of generated senders
- runtime/router (`sdk/bot`) is a Go-native development layer, not a mirror of TypeScript SDK abstractions

## Re-verify compatibility after upstream changes

Refresh snapshots and regenerate:

```bash
./scripts/update-upstream.sh
go run ./cmd/simplexgen
```

Run checks:

```bash
go test ./...
go test -race ./...
go vet ./...
```

Quick count checks:

```bash
rg -c '^export interface ' spec/upstream/commands.ts
rg -c '^type [A-Za-z0-9_]+ struct' sdk/command/generated_requests.go
rg -c '^func \(c \*Client\) Send[A-Za-z0-9_]+' sdk/client/generated_senders.go

rg -c '^  export interface ' spec/upstream/events.ts
rg -c '^\s*EventType' sdk/types/generated_tags.go

rg -c '^  export interface ' spec/upstream/responses.ts
rg -c '^\s*ResponseType' sdk/types/generated_tags.go
```

If counts diverge, regenerate first. If still diverged, parser/generator update is required.
