# RIID-4649: Security Policy Domain

[Back to distribution-host-integration](../distribution-host-integration.md)

This slice moves the pure C7 security / policy decision domain.

Moved surfaces:

- `internal/policy`
- `docs/20-domain/security.md`
- `docs/20-domain/security-redaction.md`

The package imports C11 host integration types from public
`internal/hostintegration`.

This slice does not move provider adapters, runtime/session/supervisor actors,
ToolRef.Args / EventIngestor wiring, concrete sandbox/network/OS adapters, task
DB/project/mwsd local API packages, packaging artifacts, private infra, secrets,
or local machine state.
