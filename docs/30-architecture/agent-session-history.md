# Agent session and history management

RIID-5819 owns daemon-side native session continuity. The server remains the
source of the requested resume ID; the daemon never selects a most recent
session or starts another session as an automatic recovery policy.

## Durable workspace binding

For providers with native resume capability, the filesystem workdir adapter stores a metadata-only native session binding
before publishing the provider pin. It binds provider/session ID to the exact
workspace, logical task, agent, runtime capability fingerprint, repository grant
and host execution binding. The SaaS source copies its stable Agent ID into
the provider-neutral `controlplane.MetadataAgentID`; SaaS report-routing keys
remain separate. Prompts, replies, credentials and environment values
are not stored in this index. Existing run IR remains the execution history.

A known resumed session reuses its original working directory (and therefore its
native CLI project history/config home), while each new run keeps its own event,
output and verification artifacts. Host execution checks take a fresh baseline
of the retained clone; no clone/reset discards previous work. Repository or
runtime changes, missing retained directories and conflicting pins block resume
with a diagnostic. There is no fallback to a fresh session for a known broken
binding. Operator archive cleanup remains authoritative: an expired directory
may be removed and the next resume then reports the missing history.

A native session can have only one supervisor-owned turn at a time, including
verification and terminal delivery. Bindings survive daemon restart. Older sessions without an index retain the
legacy workspace selection, with an explicit warning that continuity could not
be verified. Existing nonempty assignment clones are reused only when resuming
and their repository grant is confirmed by a binding; an unindexed Git resume
is blocked rather than silently cloning a replacement. Adapters without the
optional persistence port retain their existing behavior.

## Native identity boundary

Before an event is published, all providers share the same guard: the requested
resume ID and the first nonempty provider pin are immutable. Blank/duplicate
pins are ignored. Claude/Cursor result-only IDs pass through the same guard,
so a missing initialization event does not drop identity or bypass validation. A conflicting pin terminates the run without exposing or
persisting the foreign ID; no later success overrides that failure. A resumed
run retains its expected ID even if the provider exits before identifying it.
A successful resume must confirm that identity; a success without a pin is
blocked as unverified continuity. This does not prove provider semantic understanding or implement session fork,
OpenClaw resume, UI navigation, provider login-account identity isolation, or
provider automatic recovery. Capability fingerprints describe the executable
and execution policy, not the provider login account.

## Evidence

Hermetic tests cover two follow-ups across adapter reconstruction, retained
files/native configuration with separate run artifacts, owner/runtime/grant
isolation, missing/corrupt bindings, fresh host verification baselines, and
native identity drift before publication. No model calls are required.

요청 식별자 메타데이터를 추가해도 기존 Agent 소유권, 작업 제목, 서버가 지정한 native resume ID를 유지한다.
