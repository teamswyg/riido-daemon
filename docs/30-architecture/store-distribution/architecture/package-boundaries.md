# Store Distribution Architecture: Package Boundaries

[Back to architecture](../architecture.md)

RIID-4570 decision: `riido-daemon` owns the C11 Store App contracts and local
helper runtime shape. A future desktop/app repository may own the concrete Store
App GUI adapter and OS entitlement calls, but must consume the C11/local API
contracts instead of redefining domain facts.

```txt
Store App
  -> concrete GUI / OS entitlement adapter outside daemon domain
  -> C11 Host Integration contracts
  -> Local IPC client
  -> ConsentLedger view
  -> ExternalToolRegistry view

Local Helper / Broker
  -> C11 local IPC adapter
  -> C3 ProviderCapability
  -> C4 ProviderRuntime
  -> C5 RuntimeScheduling
  -> C6 Workspace
  -> C7 SecurityPolicy
  -> C10 SaaS polling/sync adapter

SaaS Control Plane
  -> C10 assignment / SSE / routing
  -> receives distribution metadata, never provider executable paths
```

`cmd/riido` remains the local helper binary in this repository. A future GUI
wrapper may live in another repo, but it must call the C11 contracts rather than
bypass them.

| Surface | Owner | Non-owner |
| --- | --- | --- |
| C11 domain facts and pure models | `riido-daemon` | Store App GUI repo must not copy/redefine them |
| Local helper / broker executable | `riido-daemon` (`cmd/riido`) | Store App GUI must not run provider CLIs directly |
| Local IPC handler and request envelope | `riido-daemon` | Store App GUI may only be a client |
| Store distribution executable contract | `riido-daemon` | Private infra must not weaken public review invariants |
| Store App native UI, entitlement calls, picker/bookmark adapter | future desktop/app repository | `riido-daemon` domain packages do not import GUI frameworks |
| Signing, provisioning, submission credentials, live evidence | private operator/infra environment | public repositories never store secrets |
| Shared DTO/schema needed by multiple repos | `riido-contracts` after promotion | no repo may fork the same fact |
