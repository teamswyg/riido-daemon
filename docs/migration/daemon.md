# Riido Daemon Migration Plan

This SSOT is split into focused parts so handwritten files stay below the repository line-count threshold.
The original entrypoint remains here to preserve existing links.

## Compatibility Markers

- `figma-ai-agent-daemon-boundary`
- `RIID-4901`
- `provider-validation-matrix.riido.json`
- `supports_worktree=false`
- `required_surfaces=[worktree]`
- `MISSING_REQUIRED_SURFACE:worktree`
- Store App GUI must remain a client of C11/local API contracts

## Parts

- [Part 01: Goal](daemon/part-01.md)
- [Part 02: RIID-4648 — distribution host integration domain](daemon/part-02.md)
- [Part 03: RIID-4571 — macOS external Provider CLI entitlement/review closure](daemon/part-03.md)
- [Part 04: RIID-4881 / RIID-4917 — Codex app-server auth and full-access harness correction](daemon/part-04.md)
- [Part 05: RIID-4684 — riidoapi local API adapter migration](daemon/part-05.md)
- [Part 06: RIID-4690 — full daemon lifecycle CLI wiring migration](daemon/part-06.md)
- [Part 07: RIID-4847 — Figma coverage upstream provenance full mirror guard](daemon/part-07.md)
- [Part 08: RIID-4917 — aggregated device/runtime snapshot heartbeat](daemon/part-08.md)
