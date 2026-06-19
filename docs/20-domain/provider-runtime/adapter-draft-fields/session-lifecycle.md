# Session Lifecycle

[Back to adapter-draft-fields.md](../adapter-draft-fields.md)

| Step | Adapter responsibility |
| --- | --- |
| pin | When the provider reports a session id, immediately emit `ProviderEventDraft(Type=SessionPinned, ProviderSessionID=...)`. |
| resume | `ResumeSession(ctx, providerSessionID)` starts a new run from an existing provider session; ingest assigns the new `RunID`. |
| fork | Optional experimental surface, such as Codex `thread/fork`; same flow as resume with `Payload.fork=true`. |
| close | After process stop, session is considered closed; no separate draft is emitted. |

`SessionPinned` must be emitted as soon as the provider first exposes the
session id. Late pinning breaks crash recovery and resume.
