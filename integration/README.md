# Live Contract Tests

These tests validate SDK behavior against a live SimpleX websocket API.

## Run locally

1. Start SimpleX CLI websocket API:

```bash
simplex-chat -p 5225
```

2. Run tests:

```bash
SIMPLEX_WS_URL=ws://localhost:5225 go test -tags=integration ./integration/... -v
```

## Optional fixture env vars for extended flows

Set these to enable additional contract tests:

- `SIMPLEX_TEST_CONTACT_ID` - contact ID for send/update/delete/reaction tests
- `SIMPLEX_TEST_CONTACT_CHAT_ITEM_ID` - existing contact chat item ID for update/delete/reaction
- `SIMPLEX_TEST_GROUP_ID` - group ID for send/update/delete/reaction tests
- `SIMPLEX_TEST_GROUP_CHAT_ITEM_ID` - existing group chat item ID for update/delete/reaction
- `SIMPLEX_TEST_FILE_ID` - file ID for `ReceiveFile` / `CancelFile` tests

If a required fixture variable is not set, the corresponding test is skipped.

## Coverage

Current live contracts cover:

- active user and bootstrap
- typed sender parity (`SendShowActiveUser` vs helper API)
- list users, contacts, groups
- contact message flow: send, update, add/remove reaction, delete
- group message flow: send, update, add/remove reaction, delete
- file command flow: receive, cancel
