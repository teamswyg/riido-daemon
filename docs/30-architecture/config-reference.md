# Daemon Config Reference

Riido task: RIID-4711 `[Daemon] Architecture SSOT docs migration`.

This entrypoint owns the public `riido-daemon` env/flag catalog. SaaS server
`RIIDO_AI_SERVER_*` variables belong to `riido-control-plane`; private
deploy-time evidence paths belong to `riido-infra`.

Focused sections:

- [Naming rules](config-reference/naming-rules.md)
- [Provider executable overrides](config-reference/provider-executable-overrides.md)
- [Executable search path](config-reference/executable-search-path.md)
- [Test integration gate](config-reference/test-integration-gate.md)
- [Daemon identity and runtime config](config-reference/daemon-identity-runtime-config.md)
- [Task source selection](config-reference/task-source-selection.md)
- [SaaS device principal](config-reference/saas-device-principal.md)
- [SaaS polling and binding behavior](config-reference/saas-polling-runtime-bindings.md)
- [Agent instruction placement](config-reference/agent-instruction-placement.md)
- [Local daemon flags](config-reference/local-daemon-flags.md)
- [Change procedure](config-reference/change-procedure.md)

CI coverage anchors:

- provider paths: `RIIDO_CLAUDE_PATH`, `RIIDO_CODEX_PATH`,
  `RIIDO_OPENCLAW_PATH`, `RIIDO_CURSOR_PATH`
- integration gate: `AGENTBRIDGE_INTEGRATION`
- task sources: `RIIDO_TASK_QUEUE_DIR`, `RIIDO_TASK_DB_SOURCE_PATH`,
  `RIIDO_SAAS_URL`
- policy bundle: `RIIDO_POLICY_BUNDLE_PATH`
- workdir cleanup: no default size/task-count cleanup
