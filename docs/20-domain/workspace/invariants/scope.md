# Workspace Invariants: Scope

[Back to invariants](../invariants.md)

This SSOT fills the C6 Workspace / Native Config context.

Responsibilities:

- Define workdir directory structure and lifecycle.
- Materialize native config files deterministically into task workdirs.
- Define `NativeConfigVersion` assignment rules.
- Separate repo cache from task workdir.
- Express lock-use policy at the domain level.

Non-responsibilities:

- Policy decisions belong to C7 Security / Policy migration slices.
- Lock acquisition primitives belong to [`locking.md`](../../locking.md) (C9).
- Provider flag/env mapping belongs to C4 Provider Runtime.
- Lease and scheduling belong to
  [`runtime-scheduling/invariants.md`](../../runtime-scheduling/invariants.md)
  (C5).
