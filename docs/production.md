# Production Deployment Guide

This guide describes a practical baseline for running `go-simplex` bots in production with strong data-safety defaults.

Official references:

- SimpleX bot overview: https://github.com/simplex-chat/simplex-chat/tree/stable/bots
- Bot API commands: https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/COMMANDS.md
- Bot API events: https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/EVENTS.md

## 1) Deployment model

Recommended topology:

- `simplex-chat` process exposing websocket API
- one or more Go bot processes using `go-simplex` client/runtime
- process supervisor (`systemd`) for restart and controlled rollout

Keep bot runtime and SimpleX state on the same trusted host when possible.

## 2) Service account and file permissions

- create dedicated non-login user for services (for example `simplexbot`)
- ensure SimpleX state directory and bot config directories are owned by this user
- set strict permissions:
  - directories: `0700`
  - secrets/env files: `0600`
- avoid running either process as `root`

## 3) Network and TLS

- default to loopback binding for websocket (`ws://127.0.0.1:<port>`) when bot runs on same host
- if remote access is required, terminate TLS at reverse proxy and expose only `wss://`
- restrict ingress by source IP/network; do not expose raw local websocket publicly

In bot code, enforce transport hardening:

- `ws.WithRequireWSS(true)` for remote links
- `ws.WithTLSMinVersion(...)`
- `ws.WithReadLimit(...)`

## 4) Secrets and sensitive data

- pass credentials via environment or secret manager, not command-line flags
- do not commit `.env`, private keys, websocket credentials, or fixture IDs
- avoid logging message payloads and raw command bodies in production
- rotate credentials and tokens on schedule and after incidents

## 5) Runtime hardening in SDK

Use these controls by default for production bots:

- strict outbound raw-command policy:
  - `client.WithRawCommandAllowPrefixes(...)`
  - `client.WithRawCommandValidator(...)`
  - `client.WithRawCommandMaxBytes(...)`
- bounded channels and backpressure:
  - `client.WithEventOverflowPolicy(...)`
  - `client.WithErrorOverflowPolicy(...)`
  - `client.WithDropHandler(...)`
- forward compatibility:
  - `client.WithStrictResponses(false)` during upstream migrations
- abuse mitigation:
  - `router.EnablePerContactRateLimit(...)`

## 6) Observability and incident response

- collect process logs via `journald`/central log pipeline
- add restart alerts (unexpected exits, restart loops)
- track bot-level metrics:
  - reconnect count
  - command latency
  - dropped events/errors
- keep `SECURITY.md` process visible for coordinated disclosure

## 7) Backup and restore

- backup SimpleX state directory and bot configuration daily (encrypted at rest)
- verify restore in a non-production environment on a schedule
- document RPO/RTO and operational owner
- during restore drills, validate:
  - websocket bootstrap
  - command send path
  - event handling path

## 8) Example systemd units

`simplex-chat.service`:

```ini
[Unit]
Description=SimpleX Chat API
After=network.target

[Service]
User=simplexbot
Group=simplexbot
WorkingDirectory=/opt/go-simplex
ExecStart=/usr/local/bin/simplex-chat -p 5225
Restart=always
RestartSec=2
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true

[Install]
WantedBy=multi-user.target
```

`go-simplex-bot.service`:

```ini
[Unit]
Description=Go SimpleX Bot
After=network.target simplex-chat.service
Requires=simplex-chat.service

[Service]
User=simplexbot
Group=simplexbot
WorkingDirectory=/opt/go-simplex-bot
Environment="SIMPLEX_WS_URL=ws://127.0.0.1:5225"
ExecStart=/opt/go-simplex-bot/my-bot
Restart=always
RestartSec=2
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true

[Install]
WantedBy=multi-user.target
```

Adjust hardening options if your runtime needs additional filesystem access.
