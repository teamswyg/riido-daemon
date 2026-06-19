# Split-Repo Ownership

[Back to context-map.md](../context-map.md)

`riido-daemon` must not redefine shared task/IR/provider capability facts.
When both daemon and control-plane need the same DTO/schema, promote it to
`riido-contracts` first. When a fact is deployment-only, keep it in
`riido-infra`. When a fact is server runtime behavior, keep it in
`riido-control-plane`.

Agent settings follow the same direction. `riido-contracts` owns the shared
meaning of agent profile fields and instruction limits. `riido-control-plane`
owns create/save/update API behavior. `riido-daemon` owns only the customer-PC
runtime consumption of an assigned instruction value and must not redefine:

- thumbnail presentation
- one-line description presentation
- `created_at` / `updated_at` timestamp meaning
- RBAC/editability
- API shape
- model-default request semantics
- client required-control presentation
- server storage policy

The model catalog is a runtime-scoped contracts/control-plane read-model fact.
C4 consumes only the selected model value in provider execution requests.

The daemon-side projection of Figma boundaries is
[`../30-architecture/figma-ai-agent-daemon-boundary.md`](../../30-architecture/figma-ai-agent-daemon-boundary.md).
That manifest is a downstream boundary check, not a replacement for the
contracts/control-plane SSOT.
