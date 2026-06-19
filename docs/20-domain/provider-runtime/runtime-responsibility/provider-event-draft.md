# ProviderEventDraft Output

[Back to runtime-responsibility.md](../runtime-responsibility.md)

`ProviderEventDraft` is the only domain output an adapter may produce.

`EventIngestor` receives the draft and finalizes append-only record concerns:

- identity
- ordering
- runtime identity
- attribution
- schema
- timestamp policy

After finalization, the draft is stored as `CanonicalEvent`. Authorized callers
such as FSM Orchestrator or server transition layer may participate only by
calling the EventIngestor API. They do not own a direct writer.
