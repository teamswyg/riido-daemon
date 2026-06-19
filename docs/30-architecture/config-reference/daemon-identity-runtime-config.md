# Daemon Identity And Runtime Config

[Back to Daemon Config Reference](../config-reference.md)

| Variable | Consumer | Default | Meaning |
| --- | --- | --- | --- |
| `RIIDO_DAEMON_ID` | `cmd/riido/daemon_config.go` | `RIIDO_DEVICE_ID` in SaaS mode, else `agentd-local` | stable daemon/runtime slot id |
| `RIIDO_DAEMON_VERSION` | daemon config + supervisor events | `riido-agentd v0.0.0` | daemon binary version label |
| `RIIDO_DAEMON_PROFILE` | daemon status | `local` | user-facing daemon profile |
| `RIIDO_SERVER_URL` | daemon status | empty | display-only server URL unless SaaS source is configured |
| `RIIDO_DEVICE_ID` | `saasplane` | empty | device principal id |
| `RIIDO_DEVICE_SECRET` | `saasplane` | empty | sent only as `X-Riido-Device-Secret` |
| `RIIDO_DEVICE_NAME` | daemon status | hostname or `localhost` | device display name |
| `RIIDO_RUNTIME_OWNER` | daemon status | `$USER` or `local` | runtime owner display name |
| `RIIDO_RUNTIME_AGENTS` | daemon status | empty list | comma-separated attached agent names |
| `RIIDO_WORKSPACE_COUNT` | daemon status | `0` | non-negative display count |
| `RIIDO_RUNTIME_MAX_CONCURRENT` | runtime actor slot pool | `4` | max concurrent sessions per provider |
| `RIIDO_WORKDIR_ROOT` | supervisor + workdir FS adapter | C11 app data root | isolated run workdir root |
| `RIIDO_WORKDIR_RETENTION_SECONDS` | workdir cleanup | `0` | archived workdir TTL cleanup; zero disables |
| `RIIDO_WORKDIR_CLEANUP_INTERVAL_SECONDS` | workdir cleanup | `3600` with retention | cleanup loop interval |
| `RIIDO_DAEMON_PPROF_ADDR` | local diagnostics | empty, or dev default `127.0.0.1:6061` | loopback-only pprof listener |
| `RIIDO_POLICY_BUNDLE_PATH` | policy loader | empty | optional `riido-policy-bundle.v1` file |
| `RIIDO_POLICY_BUNDLE_VERSION` | config/runtime/workdir | `policy-bundle.local.v0` | active policy version |
