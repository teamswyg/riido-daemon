# Forbidden Draft Fields

[Back to adapter-draft-fields.md](../adapter-draft-fields.md)

These fields are decided by ingest / authorized orchestration layers. If an
adapter fills them directly, it violates the `riido-contracts` IR append
authority split.

| Field | Filled by | Why not adapter |
| --- | --- | --- |
| `EventID` | EventIngestor | central ULID/UUID7 issue keeps monotonic order |
| sequence / ordering metadata | EventIngestor | events for one task must be monotonic |
| `RuntimeID` | EventIngestor lease lookup | lease owns the real runtime |
| `CapabilityFingerprint` | EventIngestor lease lookup | must match paired lease fingerprint |
| `ActorKind` | authorized caller / EventIngestor config | adapter-owned attribution breaks invariant |
| `ActorID` | server transition layer | same attribution boundary |
| `EventSchemaVersion` | EventIngestor | active reducer version, not adapter-local |
| `FSMVersion` | server transition layer | only transition events get active FSM schema |
| `OccurredAt` vs `IngestedAt` policy | EventIngestor | ingest owns timestamp policy |

Invariant:

> Adapter observes. Adapter ACL creates a normalized draft. Only EventIngestor appends.

`Only EventIngestor / FSM Orchestrator / server transition layer may append
CanonicalEvent. Provider Adapter may only produce ProviderEventDraft.`
