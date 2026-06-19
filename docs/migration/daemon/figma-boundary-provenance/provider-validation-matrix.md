# RIID-4901: Provider Validation Matrix Evidence

[Back to figma-boundary-provenance.md](../figma-boundary-provenance.md)

This slice closes the public daemon provider verification SSOT gap after the
OpenClaw, Claude Code, Cursor Agent, and Codex worktree/real-provider evidence
slices.

This slice:

- adds `docs/30-architecture/provider-validation-matrix.riido.json` as executable current evidence
- keeps `docs/30-architecture/integration-matrix.md` focused on verification policy and links it to the executable matrix
- records that Claude, Codex, and Cursor prove worktree side effects only when their opt-in real provider integration gates pass
- records that OpenClaw text completion, SaaS completion, and optional local artifact attempts must not become daemon-selected worktree support while runtime capability remains `supports_worktree=false`
- requires the C5 scheduling invariant: OpenClaw worktree-required tasks with `required_surfaces=[worktree]` must fail with `MISSING_REQUIRED_SURFACE:worktree`

It does not install provider CLIs, run provider auth setup, change provider
native commands, change SaaS endpoints, edit Desktop, edit generated client
code, or make provider CLIs bundled artifacts.
