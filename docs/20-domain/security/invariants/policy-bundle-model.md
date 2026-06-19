# Security Invariants: Policy Bundle Model

[Back to invariants](../invariants.md)

Policy bundle is the single decision source for this domain. Every gate receives
the active policy bundle and returns allow, reject, or hold.

Domain shape:

```txt
PolicyBundle {
    SchemaVersion    string
    Version          string
    EffectiveSince   time.Time
    SupersededAt     time.Time | null

    TrustTierPolicies map[TrustTier]TrustTierPolicy
    DefaultDeny       []PolicyTarget
}

TrustTierPolicy {
    AllowedSurfaces   AllowedSurfaceSet
    Sandbox           SandboxPolicy
    NetworkEgress     EgressPolicy
    Secrets           SecretsPolicy
    ProtectedPaths    []PathPattern
    DestructiveOps    DestructiveOpPolicy
    MCPAllowlist      []MCPServerID
    NativeConfigRules NativeConfigPolicy
}
```

Evolution:

- Every new bundle has a new `Version`; versions are never overwritten.
- Even when a diff allows more, each task is evaluated with its start-time bundle.
- Moving an in-flight task to a new bundle follows T-POLICY in
  [`runtime-upgrade-flow.md`](../../../30-architecture/runtime-upgrade-flow.md).

Actor:

- Policy bundles are deployed by human operator PR.
- Agents and daemon cannot change policy.
