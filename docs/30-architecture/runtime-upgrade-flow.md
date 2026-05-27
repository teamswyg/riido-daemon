# Runtime Upgrade Flow SSOT

> Riido task: RIID-4711 `[Daemon] Architecture SSOT docs migration`
>
> This document owns what the daemon does when a provider binary, capability
> fingerprint, policy bundle, or native config version changes.

## Invariant

An active provider run must not silently move to a different runtime identity
or policy/config bundle. Changes are observed, classified, and applied only at
safe boundaries.

## Versioned Inputs

| Input | Owner |
| --- | --- |
| Daemon binary version | `cmd/riido` release |
| Provider capability fingerprint | `riido-contracts/provider/capability` and adapter detect |
| Policy bundle version | C7 `internal/policy` |
| Native config version | C6 `internal/workdir` |
| Runtime ID / provider protocol | C4 runtimeactor/provider adapter |

## Flow

1. RuntimeActor periodically detects provider capability and compares the
   detected fingerprint with the active slot.
2. If no task is running, the runtime snapshot may be refreshed and future
   claims use the new fingerprint.
3. If a task is running and the fingerprint/policy/native config changes,
   RuntimeActor records the violation, cancels the provider session, and emits
   a terminal blocked/failed result through the normal reporter path.
4. Supervisor reports the event through local task DB or SaaS reporter. It does
   not mutate task state directly outside the owning adapter.
5. The next claim reevaluates compatibility from fresh inputs.

## Policy

- Silent upgrade during `Preparing`/`Running` is forbidden.
- Retry/resume decisions belong to the task lifecycle/control-plane source, not
  to provider adapters.
- Provider CLIs are external resources; detecting a newer CLI version is not a
  reason to self-update or install anything.

## Change Procedure

Any new runtime-pinned input must be added here, to the relevant domain SSOT,
and to the RuntimeActor/scheduling tests that enforce no silent migration.
