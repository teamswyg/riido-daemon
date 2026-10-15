# Agent task execution and bounded correction

[Policy and evidence manifest](agent-task-execution.riido.json) is the structured source for this reader.

RIID-5817 owns action-oriented instruction delivery, host-verified completion,
and bounded correction. The host remains the authority for repository paths,
verification commands, and whether a correction may run.

## Observed gaps and changes

Concrete implementation requests must lead to changes and validation, with an
explicit blocker when work cannot proceed. A planning-only response is not
completion. Pure questions and genuinely ambiguous requests keep their existing
clarification behavior. Final answers state outcome, changed files, verification,
and any remaining blocker in the user's language.

Codex receives Riido agent, execution, and telemetry instructions through native
developer instructions. Its user input starts with the requested task. The canonical
assignment title is display metadata for native `thread/name/set`, sent after
`turn/start` acknowledges the persisted thread. Unsupported naming emits a warning
and never fails task execution. Cursor
inlines a supplied system instruction because its CLI lacks a separate surface;
it must never silently discard that instruction. Native error flags, error
details, and error-limit result subtypes remain failures despite partial output.

Execution bindings may specify `continuation_limit` from 0 to 2 (default 0).
Only an otherwise completed run with missing/unchanged required files or a
failed host check may continue, and only with a native resumable session.
The continuation reuses the exact assignment, prepared repository, provider,
model, and permission envelope. It asks the provider to address the bounded
verification feedback and then verifies again. It never clones over changes,
reclaims the assignment, or repeats an acknowledged terminal report. Cancellation,
authentication, policy, provider failure, invalid output paths, and verification
timeouts do not trigger continuation. Exhaustion stays failed with its reason.
Verification runs outside the supervisor mailbox so checks cannot hold unrelated
assignments or cancellation. Heartbeats retain supervisor-owned execution IDs
during preparation, verification, correction and pending terminal acknowledgement,
even when the native provider slot is free. Only an acknowledged completion
releases that ownership; a failed start report never acquires it. Daemon IPC
activity counts include those retained executions so desktop updates cannot treat
verification or an unacknowledged result as idle. Evidence is
persisted before the terminal report.

OpenClaw discovery uses bounded native `models list --json`, including CLI harness
routes absent from legacy config. Its selected model is passed to `agent exec
--model` so model selection is honored. Listing a model does not prove all routes
can infer; the configured native default has its own actual file-write gate.

## Open-source comparison

Reviewed upstream source on 2026-10-13; these are design references, not copied
implementations or evidence that upstream behavior is installed locally.

- [OpenHands stuck detector](https://github.com/OpenHands/software-agent-sdk/blob/7bd668c9ab74f62c85745cd95adb503f53e7228b/openhands-sdk/openhands/sdk/conversation/stuck_detector.py)
  bounds its observation window and detects repeated action/error and monologue
  patterns. Riido applies a bounded host-evidence correction limit, avoiding a
  text-only heuristic that could interrupt legitimate long-running tools.
- [SWE-agent action execution](https://github.com/SWE-agent/SWE-agent/blob/3ea751c087f32b16e039a2233dd6eefecef325d5/sweagent/agent/agents.py)
  distinguishes proposed actions, execution observations, and exit outcomes.
  Riido likewise separates provider completion from verified file/check evidence.
- [Codex app-server](https://github.com/openai/codex/blob/be5296d4672f59205c1b8a76f46e74eae4f10d51/codex-rs/app-server/README.md)
  supplies native thread resume and instruction surfaces. Riido preserves a native
  session instead of inventing a fresh conversation to repair verification.
- [OpenClaw agent exec](https://github.com/openclaw/openclaw/blob/b16b840223aed400b89dba0a13dd325eab0d74f7/docs/cli/agent.md)
  describes an isolated execution whose cwd is the tool workspace. The installed
  CLI's disposable-file gate must pass independently; this does not override
  Riido's currently conservative worktree capability declaration.

## Provider recovery open loop

The loop registry tracks `provider-recovery-evidence-intake` as open intake.
RIID-5822 owns recurring compatibility and failure-injection observations;
RIID-5823 owns recovery policy and provider-change handling. Promote automatic
recovery only after proving: transient failure classification; native session and
workspace continuity; cancellation and approval fencing; bounded attempts and
latency; no duplicate external effects; restart/outbox convergence. Login and
quota failures require operator action, not repeated calls or provider switching.
Never auto-update CLIs or restart active work merely because a health probe fails.

## Verification scope

On the initial main baseline, real Claude, Codex, Cursor and OpenClaw disposable
file-write gates passed. This proves the selected local CLI/account environment,
not all versions, machines, accounts, approvals, or restart paths. Deterministic
regressions, cancellation/limit/concurrency race tests and fresh real instruction
probes passed after the changes. A real Codex two-turn native resume preserved
the same clone, wrote the expected file and passed the host check.
Record failed or skipped gates explicitly; unbound tasks still lack host semantic
verification and provider outage self-recovery is not claimed.

## Execution environment and status · RIID-5820

The contracts [Provider recovery design](https://github.com/teamswyg/riido-contracts/blob/main/docs/20-domain/provider-recovery.md) defines the target recovery hierarchy; this slice does not implement provider failover or blind retries. Host-only execution binding configuration owns exact workspace/task repository paths and verification argv. Atomic local updates affect subsequent executions; each preparation captures an immutable binding and resumed sessions retain their scope checks. Local IPC observations expose the actual execution directory to the host. SaaS receives only bounded workspace kind, verification status, diagnostic code and next action. Absolute paths, commands and credentials are not shared diagnostics. A provider-reported completion without a host check stays `not_configured`, not `passed`. Stop/cancel, timeout, preparation, native-session and verification failures carry distinct codes rather than inferred authentication failures.
