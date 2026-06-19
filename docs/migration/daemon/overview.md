# Riido Daemon Migration Plan: Overview

[Back to daemon.md](../daemon.md)

Riido task: RIID-4636 `[Daemon] 기존 riido_daemon daemon 마이그레이션 계획/문서화`.

This entrypoint defines how the daemon/runtime part of the former private
`riido_daemon` repository moves into public `riido-daemon`.

`riido-daemon` owns customer-PC daemon runtime, local host integration, and
provider execution boundary. It must stay public, store-reviewable, and free of
non-Riido dependencies unless a later ADR explicitly changes that rule.

Focused sections:

- [Retired historical source boundary](overview/retired-source-boundary.md)
- [Target boundary](overview/target-boundary.md)
- [Migration order](overview/migration-order.md)
- [Current migration slices](overview/current-slices.md)
