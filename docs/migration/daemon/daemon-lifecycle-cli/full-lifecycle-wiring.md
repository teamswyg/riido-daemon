# RIID-4690: Full Daemon Lifecycle CLI Wiring

[Back to daemon-lifecycle-cli](../daemon-lifecycle-cli.md)

This slice restores the public-safe `riido daemon ...` process lifecycle adapter.

Surfaces:

- `cmd/riido/daemon.go`
- `cmd/riido/daemon_config.go`
- `riido daemon start|status|health|ready|metrics|stop|logs`
- docs updates in provider-runtime, runtime-scheduling, CLI migration, and daemon
  migration SSOT files
- focused public CI for foreground/background daemon start, Unix socket
  status/health/ready/metrics, cooperative stop, PID fallback, log tailing,
  Factor 12 env config loading, control-plane source selection, public boundary
  import checks, and local-only listener checks

The CLI adapter wires public runtimeactor/supervisor/provider adapters to public
control-plane sources: in-memory offline mode, file queue, `riido-task-db.v1`
through `taskdbplane`, and SaaS assignment HTTP through `saasplane`.

Source selection is Factor 12 env based:

- `RIIDO_TASK_QUEUE_DIR`
- `RIIDO_TASK_DB_SOURCE_PATH`
- `RIIDO_SAAS_URL`

The Desktop-launched SaaS path uses `RIIDO_DEVICE_ID` / `RIIDO_DEVICE_SECRET`
and dynamic `/v1/daemon/agent-bindings`. Legacy `RIIDO_SAAS_AGENTS` and
`RIIDO_SAAS_TOKEN` are no longer read by daemon settings.

Boundary:

- The daemon command imports public daemon packages only.
- It must not import private `riido_daemon` paths or `internal/riidoaiserver`.
- It does not bundle, install, or auto-download Claude/Codex/OpenClaw/Cursor CLIs.
- Figma runtime-settings empty states and web onboarding CTA surfaces remain
  client/product/distribution routing unless a future daemon SSOT promotes local
  helper behavior.
- This slice does not move server HTTP, SSE, Terraform/AWS/deploy evidence,
  packaging artifacts, private infra, secrets, provider CLI bundling, App
  Store/MSIX helper packaging, or local machine state.
