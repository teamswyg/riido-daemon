# Provider Adapter Port

[Back to runtime-responsibility.md](../runtime-responsibility.md)

This is the domain expression of the provider adapter port. Current public Go
boundaries live in `internal/agentbridge`, `internal/agentbridge/session`,
`internal/agentbridge/bridge`, `internal/agentbridge/detectutil`,
`internal/agentbridge/runtimeactor`, `internal/agentbridge/controlplane`,
`internal/agentbridge/supervisor`, concrete provider packages, and
`cmd/riido daemon ...`.

```text
Provider {
    Capability() ProviderCapability

    StartRun(ctx, RunRequest) -> RunHandle
    Cancel(ctx, RunHandle) -> error
    Interrupt(ctx, RunHandle) -> error

    ProvideInput(ctx, RunHandle, response) -> error
    ResolveApproval(ctx, approvalID, decision) -> error

    Drafts() <-chan ProviderEventDraft

    PinSession(ctx, RunHandle, providerSessionID) -> error
    ResumeSession(ctx, providerSessionID) -> RunHandle
}
```

Rules:

1. One `Provider` instance is bound to exactly one RuntimeID + CapabilityFingerprint pair.
2. If that pair changes, a new `Provider` instance is created; no same-instance reuse.
3. `Drafts()` is the only adapter ACL output path.
4. Raw provider data must not be exposed through another path.
5. `Provider` does not import an IR log writer.
