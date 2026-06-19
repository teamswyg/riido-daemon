# Security Invariants: Scope

[Back to invariants](../invariants.md)

This SSOT fills the C7 Security / Policy context. C7 is cross-cutting and
supplies decisions to C3, C4, C5, C6, and C8.

Responsibilities:

- Define what "allowed" means.
- Define which surfaces can be active in each trust tier.
- Define policy bundle structure and evolution.
- Define six security gate locations, checks, and failure behavior.

Non-responsibilities:

- Policy execution belongs to adjacent contexts.
- Workdir file injection belongs to [`workspace.md`](../../workspace.md) (C6).
- Validation result interpretation belongs to [`validation.md`](../../validation.md)
  (C8).
- Security-compatible runtime assignment belongs to
  [`runtime-scheduling.md`](../../runtime-scheduling.md) (C5).
- Provider process flag/env mapping belongs to C4 Provider Runtime.

The context map SSOT is [`context-map.md`](../../context-map.md).
