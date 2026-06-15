# Riido Daemon Migration Plan: Figma Boundary Provenance

[Back to daemon.md](../daemon.md)

### RIID-4847 — Figma coverage upstream provenance full mirror guard

This slice tightens the daemon-side Figma boundary provenance guard.
`riido-control-plane` already mirrors the full contracts Figma coverage
stabilization history, but the daemon projection only recorded
`teamswyg/riido-contracts#52`. That was enough for the metadata page-list
limitation itself, but too weak for the broader upstream coverage manifest
dependency.

This slice does:

- expand `source_coverage_manifest_provenance.stabilized_by` to mirror the full
  contracts coverage stabilization history:
  `teamswyg/riido-contracts#38`, `#39`, `#45`, `#46`, `#51`, and `#52`
- keep `mirrored_supporting_tool_limitations[].source_stabilized_by` narrowed to
  `teamswyg/riido-contracts#52`, because that field describes the limitation
  slice, not the whole coverage manifest
- update `tools/figmaboundary` so full upstream coverage provenance and
  limitation-local provenance are verified separately

This slice does not change daemon runtime behavior, add Figma integration,
add SaaS endpoints, or make daemon the owner of Figma page discovery. It only
improves the SSOT dependency mirror between contracts coverage and daemon
projection.

### RIID-4851 — Figma coverage provenance source-field mirror marker

This slice records that daemon's upstream coverage provenance now mirrors a
contracts-owned source field, not local memory.

After `teamswyg/riido-contracts#53`, the canonical Figma coverage manifest owns
top-level `stabilized_by`. Daemon already preserved the full contracts coverage
history, but it did not identify that list as a mirror of the contracts source
field. That made the downstream boundary weaker than the contracts/control-plane
SSOT chain.

This slice does:

- add `mirrors_source_field = "stabilized_by"` to
  `source_coverage_manifest_provenance`
- record `source_field_introduced_by = "teamswyg/riido-contracts#53"`
- keep the mirrored full history as `teamswyg/riido-contracts#38`, `#39`, `#45`,
  `#46`, `#51`, and `#52`
- update `tools/figmaboundary` so the source-field marker and human doc mention
  are verified with the rest of the Figma boundary manifest

This slice does not change daemon runtime behavior, add Figma integration,
add SaaS endpoints, or make daemon the owner of Figma page discovery. It only
improves the SSOT dependency mirror between contracts coverage and daemon
projection.

### RIID-4859 — Figma onboarding draft-create downstream boundary mirror

This slice absorbs `teamswyg/riido-contracts#54` into the daemon projection.
The upstream Figma planning node `432:46849` changes onboarding explanation
order to agent draft/configuration, runtime selection, then workspace
selection. Daemon keeps that as downstream boundary evidence only: local draft
state, final create submit timing, and workspace/runtime selection are
client/control-plane facts, while daemon consumes only the final assignment
snapshot after SaaS authorization.

This slice does:

- add `teamswyg/riido-contracts#54` to the mirrored upstream coverage
  provenance
- preserve `432:46849` as non-UI daemon boundary evidence so supporting Figma
  metadata limitations do not erase the planning fact
- state in context/provider-runtime docs that client-local draft does not
  create daemon execution or a workspace-less provider start path
- update `tools/figmaboundary` so the revised onboarding order remains
  downstream-only unless contracts/control-plane promote a new executable
  daemon input

This slice does not change daemon runtime behavior, add Figma integration,
add SaaS endpoints, create a persisted draft API, create a workspace-less
agent create route, or make daemon the owner of onboarding sequence.

### RIID-4881 — DevicePrincipal config excludes team/OpenAPI inputs

This slice mirrors the upstream contracts/control-plane decision that daemon
SaaS polling is bound by DevicePrincipal credentials and assignment snapshots,
not team/OpenAPI-key configuration.

This slice does:

- document that `RIIDO_DEVICE_ID` / `RIIDO_DEVICE_SECRET` are the only daemon
  credential inputs for SaaS polling in the Desktop-launched flow
- state that `team_id`, `teamId`, OpenAPI task-context paths, Open API keys, and
  `X-Workspace-Api-Key` are not identity, binding, polling, or smoke-test inputs
  for daemon assignment execution
- keep the daemon as a downstream consumer of already-authorized assignment
  snapshots

This slice does not change daemon runtime behavior, add SaaS endpoints, edit
provider credential handling, alter workdir isolation, add deployment config, or
remove any legacy local-only task source.

### RIID-4890 — Detect-selected executable start parity

This slice closes the provider runtime gap where capability detection could
select one executable path but process start could re-resolve a different
same-name binary from `PATH`.

This slice does:

- add a provider-neutral `StartRequest.Executable` field for the executable path
  selected during adapter `Detect`
- pass the selected executable from `bridge.Run` and `runtimeactor.Submit` into
  provider `BuildStart`
- make Claude, Codex, OpenClaw, and Cursor command builders use
  `StartOptions.Executable`, then `StartRequest.Executable`, then the provider
  default executable name
- update OpenClaw integration coverage so the real prompt roundtrip starts the
  exact executable that passed OpenClaw's calendar-version detect gate
- preserve the existing env override rule: explicit `RIIDO_<PROVIDER>_PATH`
  remains a pin, not a hint

This slice does not install provider CLIs, change provider auth, add SaaS
endpoints, change assignment polling, or make daemon responsible for provider
binary distribution.

### RIID-4901 — Provider validation matrix evidence closure

This slice closes the public daemon provider verification SSOT gap after the
OpenClaw, Claude Code, Cursor Agent, and Codex worktree/real-provider evidence
slices.

This slice does:

- add `docs/30-architecture/provider-validation-matrix.riido.json` as the
  executable current evidence matrix for provider validation status
- keep `docs/30-architecture/integration-matrix.md` focused on verification
  policy and link it to the executable matrix
- record that Claude, Codex, and Cursor can prove worktree side effects only
  when their opt-in real provider integration gates actually pass
- record that OpenClaw text completion, SaaS completion, and optional local
  artifact attempts must not be promoted into daemon-selected worktree support
  while runtime capability remains `supports_worktree=false`
- require the C5 scheduling invariant: OpenClaw worktree-required tasks with
  `required_surfaces=[worktree]` must fail with
  `MISSING_REQUIRED_SURFACE:worktree`

This slice does not install provider CLIs, run provider auth setup, change
provider native commands, change SaaS endpoints, edit Desktop, edit generated
client code, or make provider CLIs bundled artifacts.

### RIID-4917 — runtime progress numeric args and telemetry prompt correction

This slice closes a live development finding where Codex could emit progress
telemetry with an integer `count` argument for code `1102`. The upstream
`riido-contracts/progressmessage` catalog already defines that argument as an
int, but the daemon projection accepted only string-valued args, so malformed or
incomplete progress telemetry could fall back to raw JSON-shaped copy later in
the public thread stream.

This slice does:

- accept primitive JSON progress args and normalize them into the string
  metadata shape used by the existing SaaS assignment event contract
- preserve rendered Korean progress text for `1102` when `count` is numeric
- update the injected telemetry instruction so providers know that `1102`
  requires `label`, `count`, and `representative_title`
- keep the public SSE/client response shape unchanged: frontend still receives
  rendered `message` strings and optional message metadata

This slice does not add new progress codes, change the append-only progress
catalog, change frontend rendering, or alter provider final-answer content.
