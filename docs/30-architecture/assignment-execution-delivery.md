# Assignment execution delivery

## Delivery contract

The existing daemon executes providers; the control-plane owns assignment state
and authenticated HTTP/SSE projections; the desktop owns local paths and process
launch; the client renders responses. No new public daemon HTTP listener or
permission bypass is introduced.

Observable acceptance criteria and ordered slices:

1. Provider protocol: Claude emits partial text before its final message without
   duplicating it. Codex modern tool requests retain RPC identity; failed and
   interrupted turns never become successful after partial output. Verify with
   parser/driver fixtures and provider package tests.
2. Delivery: preserve assignment identity, partial response identity, error
   detail and tool/approval activity through daemon reports and control-plane
   SSE. A short final delta must not wait for another delta to flush. Verify
   reporter HTTP tests and control-plane event projection tests.
3. Execution: distinguish a request to perform work from advice. Supply an
   authorized workspace and report concrete blockers and completion evidence;
   never infer completed work from a successful model reply alone. Verify an
   implementation task against a disposable workspace.
4. Local desktop: build the current daemon and control-plane, connect the chosen
   client UI to an isolated local profile, and document start/stop and manual
   streaming, failure, cancellation and execution checks.

Provider compatibility belongs to adapters. Keep old frame aliases where safe;
unknown frames remain observable. Existing authentication and workspace grants
remain authoritative. A missing repository is a prerequisite blocker, not an
invitation to guess a checkout or silently broaden filesystem permissions.

Rollout: deploy compatible server/client projections before the daemon emits new
metadata. Rollback uses the previous binary and preserves persisted assignment
data. Durable retry/recovery changes require explicit ordering and idempotency
tests; do not claim exactly-once delivery from an in-memory queue.

Stopping point: all four slices are verified locally and the user has a runnable
desktop profile. Production deployment and CLI auto-updates are outside this
work unit. Any acceptance criterion still unverified must remain listed as open.

## Current verification and open work

Verified locally on 2026-10-12 with a separate authenticated test principal and
Codex CLI 0.161.0:

- A disposable repository was cloned per assignment, calculator.py was actually
  changed, and its Python tests passed. The original checkout stayed unchanged.
- HTTP SSE emitted evolving assistant snapshots before completion. This verifies
  live transport; provider first-response latency is not eliminated.
- A deliberately failing host check produced a failed assignment with the exact
  check failure in the client-facing message.
- Stopping an active provider reached the client `stopped` state in 4 seconds;
  daemon running tasks and Codex running sessions both returned to zero.
- Regression tests reject successful model replies lacking required file changes.
  Provider, SaaS bridge, supervisor and execution race tests pass. The full daemon
  suite passes with installed CLI probes disabled through explicit test paths;
  actual Codex execution was tested separately above.

The dedicated PR branch passes the full daemon suite with external CLI probes
pinned to nonexistent test executables; focused provider/supervisor/SaaS/execution
race tests cover the integration boundaries. A fresh source build passed the
repeatable desktop validation harness with 15 SSE partial snapshots. The contracts
PR's optional workspace_id is pinned by immutable pseudo-version in go.mod.
Client-v2's query/card tests and typecheck pass in its PR worktree; its full check
is blocked by 16 baseline formatting issues. UI route wiring belongs to the
separate assignment work and is not included in this PR set.

Still open: normal client-v2 login trust with the local authorizer and manual
Electron UI verification (user selected separate local-account server/daemon
validation first). Bound repository checks are opt-in host configuration; a UI
repository picker, bounded continuation, broader CLI compatibility probes and
CLI auto-update policy are not implemented. Event uploads remain synchronous with
no durable retry, and restart recovery for an already-created repository clone
is incomplete. Heartbeats and provider cancellation are independent of held
uploads, but this does not establish durable terminal delivery. All four original
acceptance slices must not be reported complete.

## Explicit local execution bindings

`RIIDO_EXECUTION_BINDINGS_FILE` optionally points to host-owned JSON bindings.
Each binding selects an exact workspace/task pair, an absolute local Git repository,
required changed file paths, and argv-based checks with a bounded timeout. Remote
prompts cannot choose host paths or verification commands. Only committed HEAD is
cloned into the run's `workdir/repository`; the source checkout and its instructions
are preserved. Runtime-generated instructions remain in the parent workdir.
A bound run cannot report completed unless required files changed from its initial
snapshot and every configured check passed. Missing evidence reports a concrete
failure, retaining the provider response and local verification artifact. This is
opt-in; unbound advice tasks retain the existing response completion contract.
Verification does not prove arbitrary task semantics. Automatic retry/continuation
and a desktop repository-picker remain future work, not implicit authorization.

Desktop-provided `RIIDO_*_PATH` pins must reach Detect as well as process launch.
The CLI adapter boundary merges these four host environment overrides into the
Detect input without overwriting explicit caller entries. An invalid pin remains
unavailable; it must not silently select a different CLI installed on PATH.

Cancellation is delivered from refreshed heartbeat assignments in `cancelling`
or `cancelled`, even while the runtime capacity gate suspends claim polling.
Keep the local assignment identity until its terminal report succeeds. Provider
cancellation runs independently of the supervisor report mailbox, so an event
upload cannot postpone the kill. Tests cover held uploads and cancelling leases.

## PR integration boundary

This Draft preserves the source baseline used for local validation. Newer main
contains substantial module decomposition; reconcile the implementation there
before merging. Do not overwrite the newer main modules with these baseline files.
The original shared checkout remains untouched by PR preparation.
