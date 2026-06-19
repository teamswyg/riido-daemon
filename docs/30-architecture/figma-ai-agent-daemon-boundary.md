# Figma AI Agent Daemon Boundary

Riido task: RIID-4813 `[Daemon] Figma AI Agent 화면 경계 projection SSOT 게이트`.

This is the daemon-side projection of Figma `v.1.22 AI Agent`. The canonical UI
coverage owner remains `riido-contracts`; daemon records only executable
assignment, runtime, liveness, lifecycle, and provider-input boundaries.

Executable manifest:
[`figma-ai-agent-daemon-boundary.riido.json`](figma-ai-agent-daemon-boundary.riido.json),
schema `riido-figma-ai-agent-daemon-boundary.v1`.

Decision anchors:

- hardening tasks: RIID-4843, RIID-4847, RIID-4851
- upstream provenance: `teamswyg/riido-contracts#38`,
  `teamswyg/riido-contracts#52`, `teamswyg/riido-contracts#53`,
  `teamswyg/riido-contracts#54`
- mirrored field: `stabilized_by`
- limitation: `figma-metadata-page-list-underreports-pages.v1`
- representative nodes: `432:37336`, `432:46849`
- draft warning: workspace-less create is not a daemon command
- vocabulary warning: fixture is not a daemon template entity
- loop direction: Top-down and Bottom-up changes stay explicit

Focused sections:

- [Boundary criteria](figma-ai-agent-daemon-boundary/boundary-criteria.md)
- [Upstream provenance](figma-ai-agent-daemon-boundary/upstream-provenance.md)
- [Screen entries](figma-ai-agent-daemon-boundary/screen-entries.md)
- [Fixture vocabulary](figma-ai-agent-daemon-boundary/fixture-vocabulary.md)
- [Change loop](figma-ai-agent-daemon-boundary/change-loop.md)
- [Verification](figma-ai-agent-daemon-boundary/verification.md)
