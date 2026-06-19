# RIID-4813: Figma AI Agent Daemon Boundary Projection

[Back to daemon-lifecycle-cli](../daemon-lifecycle-cli.md)

This slice adds a daemon-local projection of the Figma v1.22 AI Agent screen
coverage. It records which screen facts are only consumed by the daemon and which
remain upstream contracts/control-plane/client/desktop ownership.

This slice does:

- add `docs/30-architecture/figma-ai-agent-daemon-boundary.md`
- add `docs/30-architecture/figma-ai-agent-daemon-boundary.riido.json`
- refresh stale agent-settings evidence to current `node-id=432-37336`
- replace onboarding "template entity" wording with fixture wording for `리도`,
  `영실`, `홍도`, and `지원`
- add a focused public Go test for daemon-relevant Figma nodes, manifest fields,
  cross-doc links, stale node ids, and fixture terminology
- wire architecture-docs workflow to run the boundary test when the manifest,
  provider-runtime/context-map docs, CLI docs, or tests change

This slice does not add daemon UI, Figma integration, SaaS endpoint, generated
client, provider install flow, provider CLI bundling, or a new local daemon
command. Figma remains product evidence and the contracts/control-plane coverage
manifest remains upstream.
