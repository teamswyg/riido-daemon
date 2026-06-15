# Riido Daemon Migration Plan: Daemon Lifecycle CLI

[Back to daemon.md](../daemon.md)

### RIID-4690 — full daemon lifecycle CLI wiring migration

This slice restores the public-safe `riido daemon ...` process lifecycle
adapter:

- `cmd/riido/daemon.go`
- `cmd/riido/daemon_config.go`
- `riido daemon start|status|health|ready|metrics|stop|logs`
- docs updates in provider-runtime, runtime-scheduling, CLI migration, and
  daemon migration SSOT files
- focused public CI for foreground/background daemon start, local Unix socket
  status/health/ready/metrics, cooperative stop, PID fallback, log tailing,
  12-factor env config loading, control-plane source selection, public
  boundary import checks, and local-only listener checks

The CLI adapter wires the already public runtimeactor/supervisor/provider
adapters to the already public control-plane sources: in-memory offline mode,
file queue, `riido-task-db.v1` via `taskdbplane`, and SaaS assignment HTTP via
`saasplane`. Source selection is 12-factor env based:
`RIIDO_TASK_QUEUE_DIR`, `RIIDO_TASK_DB_SOURCE_PATH`, or `RIIDO_SAAS_URL`. The
Desktop-launched SaaS path uses `RIIDO_DEVICE_ID` / `RIIDO_DEVICE_SECRET` and
dynamic `/v1/daemon/agent-bindings`. Legacy `RIIDO_SAAS_AGENTS` /
`RIIDO_SAAS_TOKEN` inputs are no longer read by the daemon settings model.

The daemon command imports public daemon packages only and must not import
private `riido_daemon` paths or `internal/riidoaiserver`. It does not bundle,
install, or auto-download Claude/Codex/OpenClaw/Cursor CLIs.

Figma runtime-settings empty states (`node-id=275-22731`) do not change that
boundary. Provider install cards and hover states are client/product
presentation over external provider links, and Windows app waitlist /
marketing-consent mutations are not daemon commands.

Figma web onboarding (`node-id=236-29749`) does not change that boundary either.
The macOS app download CTA is distribution/product routing to a Riido desktop
artifact, not a daemon command to install provider CLIs. Sign-up, terms consent,
member invite, Windows waitlist/marketing consent, chat animation, and
progress-bar references remain client/auth/team/product surfaces unless a future
daemon SSOT explicitly promotes a local helper behavior.

This slice does not move server HTTP implementation, SSE transport,
Terraform/AWS/deploy evidence, packaging artifacts, private infra, secrets,
provider CLI bundling, App Store/MSIX helper packaging, or local machine state.

### RIID-4703 — store distribution contract migration

This slice moves the executable store distribution contract into the public
daemon repository:

- `packaging/store/riido_daemon_store_distribution.riido.json`
- `tools/storecontract`
- `docs/30-architecture/store-distribution.md`
- `NOTICE.md`
- `.github/workflows/store-distribution-contract.yml`

The gate fixes the public daemon boundary for Developer ID, Mac App Store,
MSIX sideload, and Microsoft Store review surfaces. It makes provider CLI
non-bundling, store-managed update rules, local-only IPC, App Sandbox/login
item expectations, Windows named pipe/package local data expectations, demo
review account surface, and privacy metadata allowlist requirements executable.

This slice does not build/sign/notarize app bundles, produce MSIX packages,
submit to App Store Connect or Partner Center, bundle provider CLIs, move
private infra/account artifacts, or change the control-plane review account
runtime seed. The SaaS review account artifact remains owned by
`riido-control-plane`.

### RIID-4711 — architecture SSOT docs migration

This slice moves the public daemon architecture SSOT into `riido-daemon` after
the split-repo package migration.

This slice does:

- add `docs/20-domain/context-map.md` for public daemon bounded-context
  ownership and split-repo dependency direction
- add `docs/30-architecture/module-decomposition.md` for hexagonal package and
  import rules
- add `docs/30-architecture/config-reference.md` for daemon-only Factor 12
  env/flag ownership
- add `docs/30-architecture/integration-matrix.md` for optional real provider
  CLI integration gates
- add `docs/30-architecture/compatibility-gate.md` and
  `docs/30-architecture/runtime-upgrade-flow.md` for pre-execute and
  no-silent-upgrade boundaries
- add `docs/50-roadmap/open-questions.md` for public daemon unresolved
  questions referenced by domain SSOT docs
- add focused public CI for architecture doc presence, stale split-repo
  wording, config coverage, dependency boundary, and Go tests

This slice does not move `cmd/riido_ai_server`, `internal/riidoaiserver`,
Terraform/AWS/deploy evidence, private state, `.riido-local`, provider CLI
binaries, or provider installation automation.

### RIID-4630 — ApprovalRequested timeout owner SSOT cleanup

This slice closes the public daemon `Q-RT-003` open question by moving the
approval wait timeout decision into the provider-runtime SSOT.

This slice does:

- state that C4 session actor run clocks own approval wait timeout policy
- remove `Q-RT-003` from daemon open questions
- keep `EventIngestor` as an append authority only, not a timeout owner
- keep UI/review surfaces as display/response senders, not terminal timeout
  sources
- add a focused public workflow that fails if `Q-RT-003` drifts back into open
  questions
- add a reducer test that `EventToolApprovalNeeded` resets the semantic idle
  watchdog

This slice does not change provider-native approval RPC frames, add UI, change
CLI flags, introduce dependencies, or alter hard/semantic timeout defaults.

### RIID-4813 — Figma AI Agent daemon boundary projection gate

This slice adds a daemon-local projection of the Figma v1.22 AI Agent screen
coverage. It records which screen facts are only consumed by the daemon and
which remain upstream contracts/control-plane/client/desktop ownership.

This slice does:

- add `docs/30-architecture/figma-ai-agent-daemon-boundary.md`
- add `docs/30-architecture/figma-ai-agent-daemon-boundary.riido.json`
- refresh stale agent-settings evidence to the current `node-id=432-37336`
- replace onboarding "template entity" wording with fixture wording where the
  daemon boundary discusses `리도`, `영실`, `홍도`, and `지원`
- add a focused public Go test that checks daemon-relevant Figma nodes,
  manifest fields, cross-doc links, stale node ids, and fixture terminology
- wire the architecture-docs workflow to run that boundary test when the
  manifest, provider-runtime/context-map docs, CLI docs, or test changes

This slice does not add a daemon UI, Figma integration, SaaS endpoint,
generated client, provider install flow, provider CLI bundling, or new local
daemon command. It also does not make Figma a daemon SSOT; Figma remains product
evidence and the contracts/control-plane coverage manifest remains upstream.

### RIID-4843 — Figma metadata page-list limitation downstream guard

This slice mirrors the upstream `riido-contracts` Figma metadata tooling
limitation into the daemon boundary manifest.

This slice does:

- record `teamswyg/riido-contracts#52` as the upstream provenance for the
  `figma-metadata-page-list-underreports-pages.v1` limitation
- require the daemon projection to preserve authoritative pages `129:5215`,
  `42:3014`, and `0:1`, even when supporting metadata lists only the UI page
- require non-UI/onboarding daemon evidence nodes to stay in
  `figma-ai-agent-daemon-boundary.riido.json`
- add a focused public Go test for the mirrored limitation and downstream
  non-UI node preservation

This slice does not change daemon runtime behavior, add Figma integration,
add SaaS endpoints, or make daemon the owner of Figma page discovery. The
canonical Figma coverage and inspection method remain in `riido-contracts`;
daemon only guards its downstream projection.
