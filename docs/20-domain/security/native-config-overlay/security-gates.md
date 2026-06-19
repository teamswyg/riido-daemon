# Native Config Overlay: Six Security Gates

[Back to native-config-overlay](../native-config-overlay.md)

Security gates are ordered by decision point. Each gate has one responsibility.

| # | Gate | Location | Caller | Input | On failure |
| --- | --- | --- | --- | --- | --- |
| G-S1 | `PreClaimSecurityGate` | before task claim | C5 scheduler | task surfaces, runtime capability/trust tier, policy bundle | task `Blocked(category=POLICY_*)` |
| G-S2 | `PreExecuteSecurityGate` | before provider start | C4 adapter/orchestrator | capability, workdir state, policy bundle | task `Blocked`, provider not started |
| G-S3 | `ToolUseSecurityGate` | before `ToolCallStarted` | server transition layer | tool, args, runtime tier, policy bundle | provider `Interrupt` or `ApprovalRequested` |
| G-S4 | `FileEffectSecurityGate` | after `FileChanged` / `CommandStarted` | server transition layer | path, kind, diff, protected paths, sandbox | rollback request and `BlockerRaised(SECURITY_VIOLATION)` |
| G-S5 | `NetworkEgressGate` | on network attempt | adapter ACL through server | host, port, protocol, allowlist | provider deny result and event |
| G-S6 | `PreCompleteAuditGate` | before `PatchReady -> Completed` | C8 validation result handler | IR log, diff, affected paths | `Blocked(SECURITY_AUDIT_FAILED)` or forced `HumanReview` |

Rules:

1. Gates decide only. The calling context executes the consequence.
2. Every gate decision is persisted as an IR event.
3. Adding a gate is `change:additive`.
4. Changing an existing gate responsibility is `change:breaking-policy`.
