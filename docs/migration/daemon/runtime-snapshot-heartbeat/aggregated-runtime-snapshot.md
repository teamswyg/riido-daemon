# RIID-4917: Aggregated Device / Runtime Snapshot Heartbeat

[Back to Runtime Snapshot Heartbeat](../runtime-snapshot-heartbeat.md)

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
