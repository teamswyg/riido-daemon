# Riido Daemon Migration Plan: riidoapi Local API

[Back to daemon.md](../daemon.md)

This migration area is split by RIID so each file exposes one public-surface
decision and its evidence boundary.

- [RIID-4684 — riidoapi local API adapter](riidoapi-local-api/4684-local-api-adapter.md)
- [RIID-4685 — task/api/bridge CLI adapter](riidoapi-local-api/4685-task-api-bridge-cli.md)
- [RIID-4686 — mwsdbridge/project projection sync](riidoapi-local-api/4686-mwsdbridge-project-sync.md)
- [RIID-4689 — saasplane assignment polling adapter](riidoapi-local-api/4689-saasplane-assignment-polling.md)

The entrypoint intentionally stays small. Runtime ownership, local IPC
boundaries, SaaS assignment polling, and client presentation boundaries are
documented in the RIID-specific files above.
