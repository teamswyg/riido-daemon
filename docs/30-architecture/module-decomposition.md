# Module Decomposition SSOT

> Riido task: RIID-4711 `[Daemon] Architecture SSOT docs migration`
>
> This document owns the public `riido-daemon` package layout, hexagonal import
> rules, and 12-factor daemon adapter boundary.

## Decisions

1. The Go module is `github.com/teamswyg/riido-daemon`.
2. The only deployable binary in this repository is `cmd/riido`.
3. SaaS server code stays in `riido-control-plane`; deploy/apply/evidence code stays in `riido-infra`; shared DTO/schema facts stay in `riido-contracts`.
4. Claude, Codex, OpenClaw, and Cursor CLIs are external attached resources. This repository never bundles, installs, or silently downloads them.
5. Domain packages depend inward on contracts and ports. Provider, local IPC, filesystem, process, SaaS HTTP, and host/store behavior enter through adapters.
6. Store App GUI code is not a daemon domain package. A future desktop/app repository may own UI and OS entitlement adapters, while `riido-daemon` owns the C11 contracts, helper runtime, local IPC server, and store distribution gates.
7. Future Store App GUI adapter ownership must remain an outer adapter concern, never a daemon domain dependency.

## Detail Surfaces

- [Package map](module-decomposition/package-map.md)
- [Import rules](module-decomposition/import-rules.md)
- [Hexagonal ports](module-decomposition/hexagonal-ports.md)
- [12-factor boundary](module-decomposition/12-factor-boundary.md)
- [Change procedure](module-decomposition/change-procedure.md)

The user/operator command boundary is owned by
[`cli-surface.md`](cli-surface.md). `cmd/riido` remains a thin adapter over
these packages; command semantics stay in the backing domain packages and SSOT
docs.
