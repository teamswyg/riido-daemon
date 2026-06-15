# Riido Daemon Migration Plan: Part 08

[Back to daemon.md](../daemon.md)

### RIID-4917 — aggregated device/runtime snapshot heartbeat

This slice closes a live development finding where an idle-but-running daemon
did not refresh the SaaS device/runtime read model after initial runtime
registration. Once control-plane applies the 20 second stale projection, the
daemon must refresh liveness even when no assignment is active.

This slice does:

- store registered runtime snapshot facts inside `saasplane` mailbox state
- send an aggregated daemon runtime snapshot during the 5 second heartbeat
  cadence for DevicePrincipal mode
- include daemon process facts in that snapshot: profile, PID, started-at
  timestamp, and uptime seconds; daemon app version may also be sent as
  server-side telemetry without changing the client read shape
- preserve provider model catalog and experimental opt-in facts in heartbeat
  snapshots
- rate-limit same-tick runtime heartbeat calls so the daemon does not fan out
  one SaaS snapshot request per runtime

This slice does not change frontend generated code, assignment SSE shape,
provider command construction, device credential enrollment, or static
test-only bearer-token paths.

### RIID-4917 — device-scoped SaaS daemon identity

This slice closes a development demo finding where multiple customer-PC daemons
used the default `agentd-local` daemon id. Because provider runtime ids are
derived as `{daemon_id}:{provider}`, unrelated devices could collide on
`agentd-local:codex`, causing generated client daemon details and assignment
polling to point at another user's latest runtime projection.

This slice does:

- keep explicit `RIIDO_DAEMON_ID` behavior unchanged
- default the daemon id to `RIIDO_DEVICE_ID` when SaaS DevicePrincipal
  credentials are present and no explicit daemon id is configured
- preserve the local non-SaaS default `agentd-local`
- document that Desktop-launched SaaS runs should use the device-principal
  default so runtime ids are device-scoped

This slice does not change device enrollment, generated client routes,
assignment SSE shape, provider command construction, or local task-db/file-queue
daemon identity.

## Validation Gates

Required before a daemon migration PR is mergeable:

```bash
go test ./...
go list -m all
go test ./tools/storecontract
go run ./tools/storecontract -contract packaging/store/riido_daemon_store_distribution.riido.json -repo .
test -f docs/20-domain/context-map.md
test -f docs/30-architecture/config-reference.md
```

When the migrated files include the old audit tooling, restore the stronger
private-repo gates in public CI:

```bash
make check
```

Provider real-CLI integration checks stay environment-gated:

```bash
AGENTBRIDGE_INTEGRATION=1 go test ./internal/provider/... -run TestIntegration -v
```

## Store Review Invariants

- Provider CLIs are external tools, not bundled app payloads.
- The daemon must expose local-only IPC, not public TCP listeners.
- Unsafe provider modes are opt-in policy decisions, not defaults.
- Host trust tier must reject unsafe bypass.
- App Store and MSIX helper/runtime contracts stay in C11 docs and tests.

## Open Follow-Ups

| Follow-up | Repository |
| --- | --- |
| Promote shared DTO/schema only after two repositories need the same fact. | `riido-contracts` / RIID-4637 |
| Move SaaS server code separately. | `riido-control-plane` / RIID-4638 |
| Move Terraform/deploy evidence privately. | `riido-infra` / RIID-4639 |
