# Current Daemon Slice Status

[Back to Assignment Lifecycle FSM](assignment-lifecycle-fsm.md)

Verified on 2026-06-17:

- `saasplane` runtime key prefers `assignment_id`; logical `task_id` stays metadata.
- cancellation watcher, runtime mapping, and partial body buffer are execution-id keyed.
- supervisor shares per-task lifetime context across prepare, submit, and cancellation watcher.
- workspace/IR path computation can still use logical task metadata.
- SaaS JSON transport retries only safe/idempotent transient failures.
- launch env injects detectutil frozen PATH into build and spawn paths.
- slow clone/hash runs outside the actor loop so heartbeat and poll continue.
- public GitHub worktree materialization shallow-clones before provider start.
- private/token-bearing/unsupported repo fails closed before provider start.
- active assignment recovery resumes only when durable `provider_session_id` exists.
- control-plane action response can expose active thread stream handoff.

Remaining boundaries:

- private repository auth token broker remains contracts/control-plane/infra policy.
- client/desktop optimistic cache, stream subscription, and update/quit handoff are outside daemon ownership.
- broader `WorkspacePlan`, stream envelope, and approval DTO codegen promotion remains future contracts work.
