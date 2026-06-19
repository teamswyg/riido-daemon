# Allowed Draft Fields

[Back to adapter-draft-fields.md](../adapter-draft-fields.md)

```text
ProviderEventDraft {
    Type              ir.EventType
    Payload           map[string]any
    Unknown           map[string]any

    ProviderSessionID string
    ProviderTurnID    string

    RawType           string
    Raw               map[string]any

    ObservedAt        time.Time
}
```

Allowed meanings:

- `Type`: normalized event type, mostly Cat C; ingest validates transition candidates.
- `Payload`: normalized payload shaped by `riido-contracts` IR event log catalog.
- `Unknown`: preserved unknown raw fields.
- `ProviderSessionID`: provider-native session/thread id.
- `ProviderTurnID`: provider-native turn id when available.
- `RawType` / `Raw`: replay and reinterpretation asset.
- `ObservedAt`: time when the adapter read the provider line/event.
