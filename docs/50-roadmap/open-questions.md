# Open Questions

> Riido task: RIID-4711 `[Daemon] Architecture SSOT docs migration`
>
> This file owns public daemon unresolved questions that are referenced by
> domain SSOT docs. Closed implementation decisions should move into the owning
> domain or architecture document instead of staying here.

| ID | Area | Question | Current handling |
| --- | --- | --- | --- |
| Q-CTX-005 | C11/UI ownership | Does the Store App/helper UI live in this repo or a future desktop repo? | Keep public daemon contracts local; concrete GUI can live elsewhere if it calls C11/local API contracts. |
| Q-DIST-003 | Consent storage | Should `ConsentLedger` persist only as local JSON append log or also use OS secure storage? | Pure model is public; durable OS adapter remains future work. |
| Q-WS-002 | Workdir retention | What default retention size/time should the daemon use? | Default cleanup remains disabled; operators opt in with env. |
| Q-RT-003 | Approval timeout ownership | Which layer owns provider-native approval timeout UI/telemetry? | C4 emits provider-neutral pending/blocked state; concrete UX remains adapter/UI work. |
| Q-GATE-001 | Real CLI CI | Should provider real-CLI integration run on a scheduled public workflow? | Default PR CI stays deterministic; scheduled/manual integration can be added without bundled CLIs. |

Questions that require shared DTO/schema changes must first be evaluated under
the `riido-contracts` promotion rule.
