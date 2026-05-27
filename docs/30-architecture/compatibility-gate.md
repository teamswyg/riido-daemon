# Compatibility Gate SSOT

> Riido task: RIID-4711 `[Daemon] Architecture SSOT docs migration`
>
> This document owns the daemon pre-execute compatibility gate between C3
> provider capability, C7 policy, C11 host provenance, and C4 runtime startup.

## Purpose

The daemon must not start a provider process merely because a binary exists.
It starts only when provider detection, policy, host integration, and runtime
request requirements are compatible.

## Gate Inputs

| Input | Owner |
| --- | --- |
| Provider kind/protocol/detected version/fingerprint | provider adapter detect path + `riido-contracts/provider/capability` |
| User-selected executable path provenance | C11 `internal/hostintegration` |
| Store/distribution channel restrictions | C11 + C7 store channel policy |
| Unsafe bypass/tool/native-config permissions | C7 `internal/policy` |
| Task provider/runtime requirements | C5 scheduling and control-plane source adapters |

## Gate Order

1. Resolve executable using `RIIDO_<PROVIDER>_PATH` or PATH fallback.
2. If an explicit override is invalid, fail closed without PATH fallback.
3. Run adapter detect/probe and normalize to provider capability contract.
4. Apply C11 provenance/store-channel constraints.
5. Apply C7 policy decisions for unsafe bypass, native config, and tool-use
   surfaces.
6. C5 scheduling selects only runtimes whose capability snapshot satisfies the
   task requirement.
7. C4 runtimeactor starts the session only after the selected runtime remains
   compatible.

## Outputs

The gate emits provider-neutral compatibility status, blocked/degraded reasons,
detected fingerprint metadata, and runtime status snapshots. It does not expose
raw provider JSON or executable path details to SaaS telemetry.

## Failure Semantics

- Missing optional provider binary: unavailable, not fatal to daemon startup.
- Misconfigured explicit path: unavailable with fail-closed reason.
- Policy/store violation: blocked and not runnable.
- Runtime fingerprint drift while running: handled by runtime upgrade flow.

## Change Procedure

Changes to compatibility meaning must update provider capability contracts or
C7/C11 docs before adapter code. Public CI should verify deterministic detect
and policy cases without requiring real provider CLIs.
