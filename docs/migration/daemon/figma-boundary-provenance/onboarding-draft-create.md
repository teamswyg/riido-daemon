# RIID-4859: Onboarding Draft-Create Boundary

[Back to figma-boundary-provenance.md](../figma-boundary-provenance.md)

This slice absorbs `teamswyg/riido-contracts#54` into the daemon projection. The
upstream Figma planning node `432:46849` changes onboarding explanation order to
agent draft/configuration, runtime selection, then workspace selection.

Daemon keeps that as downstream boundary evidence only:

- local draft state is a client/control-plane fact
- final create submit timing is a client/control-plane fact
- workspace/runtime selection is a client/control-plane fact
- daemon consumes only the final assignment snapshot after SaaS authorization

This slice:

- adds `teamswyg/riido-contracts#54` to mirrored upstream coverage provenance
- preserves `432:46849` as non-UI daemon boundary evidence
- states in context/provider-runtime docs that client-local draft creates no daemon execution or workspace-less provider start path
- updates `tools/figmaboundary` so the revised onboarding order remains downstream-only

It does not change daemon runtime behavior, add Figma integration, add SaaS
endpoints, create a persisted draft API, create a workspace-less agent create
route, or make daemon the owner of onboarding sequence.
