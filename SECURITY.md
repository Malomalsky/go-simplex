# Security Policy

## Reporting a vulnerability

If you discover a security issue, please report it privately.

Preferred channel:

- GitHub Security Advisories for this repository (private report)

Please include:

- affected package and version/commit
- reproduction steps or proof-of-concept
- impact assessment
- suggested remediation, if known

## Response expectations

- initial triage acknowledgement target: within 72 hours
- remediation timeline depends on impact and complexity

## Scope notes

Security-sensitive areas in this repository include:

- websocket transport and TLS handling (`sdk/transport/ws`)
- raw command validation and execution boundaries (`sdk/client`)
- generated command/type compatibility with upstream snapshots

## Disclosure

Please do not open public issues for unpatched vulnerabilities.
