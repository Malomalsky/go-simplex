# SimpleX Bot API Research Notes

Date: 2026-02-26
Upstream branch inspected: `simplex-chat/stable`

## Sources

- https://github.com/simplex-chat/simplex-chat/tree/stable/bots
- https://github.com/simplex-chat/simplex-chat/tree/stable/bots/api
- https://github.com/simplex-chat/simplex-chat/blob/stable/bots/README.md
- https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/COMMANDS.md
- https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/EVENTS.md
- https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/TYPES.md
- https://github.com/simplex-chat/simplex-chat/tree/stable/bots/src/API/Docs

## What is officially supported for bots

The official bot interface is WebSocket JSON over the `simplex-chat` CLI process running as a local server (default examples use port `5225`).

Wire format:

- Command request:
  - `{"corrId":"<id>","cmd":"<cli command string>"}`
- Command response:
  - `{"corrId":"<same id>","resp":{"type":"<response-tag>", ...}}`
- Async event:
  - `{"resp":{"type":"<event-tag>", ...}}`

Important contract notes from upstream docs:

- Events and responses share `resp` envelope for compatibility.
- Some tags may appear as both response and event (`newChatItems` is the key example).
- Bot parser must tolerate:
  - undocumented event/response tags,
  - unknown enum values,
  - extra JSON properties.

## Documented API surface (stable as of 2026-02-26)

- Command categories: 8
- Documented commands: 42
- Documented response tags: 45
- Documented event tags: 45
- Documented types in `TYPES.md`: 172

Command categories:

- Address commands
- Message commands
- File commands
- Group commands
- Group link commands
- Connection commands
- Chat commands
- User profile commands

Event categories:

- Contact connection events
- Message events
- Group events
- File events
- Connection progress events
- Error events

## Behavior details that affect SDK design

1. `corrId` is client-managed.
- SDK must generate unique correlation IDs and match responses to pending requests.

2. Unknown data must not crash client.
- Typed decoding should allow fallback payload for unknown tags.

3. Network usage semantics exist in command docs.
- `no`, `interactive`, `background` do not change protocol framing, but matter for user expectations, retries, and helper APIs.

4. Bot bootstrap sequence is explicit in docs.
- Usual startup path:
  - get active user,
  - show/create bot address,
  - set address settings (auto-accept),
  - process `newChatItems` and connection events.

5. Files are controlled via message API + file commands.
- No dedicated "send file command" in the documented bot subset.
- SDK should expose clear helper methods for send/receive lifecycle.

6. Security model of CLI WebSocket endpoint.
- No built-in auth.
- Intended for localhost.
- Remote use requires external reverse proxy + TLS + auth.

## Source-of-truth generation in upstream

Upstream API docs and TS types are generated from Haskell metadata and tests enforce generated outputs.

Key generator files:

- `bots/src/API/Docs/Generate.hs` (markdown)
- `bots/src/API/Docs/Generate/TypeScript.hs` (TS codegen)
- `bots/src/API/Docs/Commands.hs`
- `bots/src/API/Docs/Responses.hs`
- `bots/src/API/Docs/Events.hs`
- `bots/src/API/Docs/Types.hs`
- `tests/APIDocs.hs` (generation consistency tests)

Implication for our Go SDK:

- We should avoid hand-maintaining command/response/event schemas.
- We should generate Go contracts from upstream artifacts, and run parity checks in CI.

## Scope boundaries to reflect in plan

- The docs explicitly mention CLI supports additional commands/events beyond reference.
- A "full implementation" for this SDK should mean:
  - full parity with documented bot API first,
  - graceful handling of undocumented/unknown records,
  - extension points for additional commands.

