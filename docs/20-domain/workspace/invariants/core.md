# Workspace Invariants: Core

[Back to invariants](../invariants.md)

Primary invariant:

> Workspace does not decide policy. Workspace deterministically materializes the
> policy/native config bundle allowed by C7 Security into the task workdir.

Additional invariants:

1. Every task run has an isolated workdir tree. Two runs of the same task use
   separate `run_id` directories.
2. `WorkspacePrepared` is required before `Running`, but is not a `Claim`
   precondition. C5 may inspect feasibility while claiming; `Preparing -> Running`
   owns the actual prepared-state precondition.
3. A run cannot enter `Running` without `NativeConfigVersion`.
4. `PolicyBundleVersion` and `NativeConfigVersion` are fixed before `Running`.
   `NativeConfigVersion` is required on execution-bound RunScope events only.
5. `protected path` is decided by Security (C7); Workspace implements the
   resulting filesystem visibility and permissions.
6. Shared repo cache locks are short. Wrapping the whole agent run in a repo lock
   violates this invariant.
