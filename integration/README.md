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

Notes:

- Tests are excluded from default `go test ./...` via build tag `integration`.
- If `SIMPLEX_WS_URL` is not set, live tests are skipped.
- These tests cover active user, bootstrap flow, typed sender parity, and users list contracts.
