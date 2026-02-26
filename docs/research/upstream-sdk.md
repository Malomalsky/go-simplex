# SimpleX TypeScript SDK Research Notes

Date: 2026-02-26
Upstream branch inspected: `simplex-chat/stable`

## Sources

- https://github.com/simplex-chat/simplex-chat/tree/stable/packages/simplex-chat-client/typescript
- https://github.com/simplex-chat/simplex-chat/tree/stable/packages/simplex-chat-client/types/typescript
- `packages/simplex-chat-client/typescript/src/*.ts`
- `packages/simplex-chat-client/types/typescript/src/*.ts`
- `packages/simplex-chat-client/typescript/tests/*.ts`

## Package split upstream

1. `@simplex-chat/types`
- Auto-generated types:
  - commands (`commands.ts`)
  - responses (`responses.ts`)
  - events (`events.ts`)
  - API types (`types.ts`)
- Includes command string formatters (`cmdString`) for commands and some types.

2. `simplex-chat`
- Runtime client with:
  - WebSocket transport,
  - async queue,
  - pending-command map by `corrId`,
  - high-level helper methods around common bot operations.

## Runtime architecture in TypeScript client

Main pieces:

- `WSTransport`: websocket wrapper, buffered writes with timeout.
- `ChatTransport`: JSON encoding/decoding for protocol envelopes.
- `ABQueue`: bounded async queue abstraction with async iterator support.
- `ChatClient`:
  - incrementing `corrId`,
  - `Map<corrId, pending request>`,
  - background receive loop:
    - if frame has `corrId`, resolve pending command,
    - else push event into `msgQ`.

## Existing high-level API coverage

The TS client ships helper methods for:

- active user create/get,
- address create/get/delete,
- address auto-accept settings,
- send/update/delete messages,
- contact connect/invite/accept/reject,
- group lifecycle and member management,
- file receive,
- profile update.

Observations:

- Helper coverage is practical but not exhaustive for every documented command variant.
- Library focuses on convenience over strict low-level completeness.

## Error-handling behavior and implications

Current behavior (important for Go design):

- Low-level `sendChatCmd` returns raw typed `ChatResponse`.
- High-level helpers usually switch on response tags and throw on mismatch.
- Parse/transport errors are represented as `ChatResponseError`.
- Incoming response with unknown `corrId` is logged and ignored.

Potential gaps we should improve in Go:

- Better typed error hierarchy (`errors.Is/As`).
- Explicit request timeouts and cancellation via `context.Context`.
- Configurable policy for unmatched `corrId` frames.
- Better reconnect hooks and lifecycle state machine.

## Concurrency model implications for Go SDK

- TS queue-based model maps naturally to Go channels + goroutines.
- Need clear ownership for:
  - write loop,
  - read loop,
  - pending requests map with locking,
  - graceful shutdown and context cancellation.

## API ergonomics implications for Go

To feel idiomatic in Go:

- `context.Context` on all network operations.
- Functional options for client config.
- Strongly typed `Command` wrappers plus raw-command escape hatch.
- Event stream via:
  - channel API (`<-chan EventEnvelope`),
  - optional typed router/dispatcher.

## Testing insights from upstream TS project

- Existing TS queue tests validate bounded async queue semantics.
- Integration client tests are mostly skipped and note event-order nondeterminism.

Implications for our plan:

- Keep transport unit tests deterministic.
- Use integration tests with predicate-based waits, not strict event ordering.
- Build contract/parity tests against generated command strings and tag sets.

