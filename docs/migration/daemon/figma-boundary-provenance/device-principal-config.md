# RIID-4881: DevicePrincipal Config

[Back to figma-boundary-provenance.md](../figma-boundary-provenance.md)

This slice mirrors the upstream contracts/control-plane decision that daemon
SaaS polling is bound by DevicePrincipal credentials and assignment snapshots,
not team/OpenAPI-key configuration.

This slice documents:

- `RIIDO_DEVICE_ID` and `RIIDO_DEVICE_SECRET` are the only daemon credential inputs for SaaS polling in the Desktop-launched flow
- `team_id`, `teamId`, OpenAPI task-context paths, Open API keys, and `X-Workspace-Api-Key` are not identity inputs
- those values are not binding, polling, or smoke-test inputs for daemon assignment execution
- daemon remains a downstream consumer of already-authorized assignment snapshots

It does not change daemon runtime behavior, add SaaS endpoints, edit provider
credential handling, alter workdir isolation, add deployment config, or remove
legacy local-only task sources.
