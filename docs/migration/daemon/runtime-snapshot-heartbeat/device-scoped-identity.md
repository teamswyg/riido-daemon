# RIID-4917: Device-Scoped SaaS Daemon Identity

[Back to Runtime Snapshot Heartbeat](../runtime-snapshot-heartbeat.md)

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
