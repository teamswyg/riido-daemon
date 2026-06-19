# Verification Evidence

[Back to Assignment Lifecycle FSM](assignment-lifecycle-fsm.md)

The executable manifest is verified by `go test ./tools/agentexecutionevidence`.

| Risk | Evidence |
| --- | --- |
| same task multi-assignment | `TestRuntimeActorUsesAssignmentIDAsExecutionKey` |
| public worktree materialization | `TestMaterializeAssignmentWorktreeRunsShallowBranchClone` |
| private repo fail-closed | `TestSupervisorBlocksPrivateAssignmentWorktreeBeforeProviderStart`, `TestStoreActorDropsSensitiveAssignmentWorktreeURL` |
| async workspace preparation | `TestSupervisorHeartbeatContinuesDuringWorkspaceMaterialization` |
| restart recovery fresh-start refusal | `TestPlaneFailsActiveAssignmentWithoutSessionAfterLocalStateLoss` |
| restart recovery resume | `TestPlaneClaimsActiveAssignmentAfterLocalStateLoss` |
| watcher release | `TestSupervisorCancelsCancellationWatchAfterTaskCompletion` |
| stale PID kill refusal | `TestDaemonStopRejectsNonDaemonPidFile` |
| launch PATH freeze | `TestRuntimeActorPassesDetectedExecutableToBuildStartAndSpawn` |
| capability TTL re-detect | `TestRuntimeActorRefreshesUnavailableCapabilityAfterTTL` |
| transient poll retry | `TestPlaneRetriesTransientPoll` |
| transient bindings retry | `TestPlaneRetriesTransientAgentBindings` |
| transient heartbeat retry | `TestPlaneRetriesTransientHeartbeat` |
| transient runtime snapshot retry | `TestPlaneRetriesTransientRuntimeSnapshot` |
| permanent poll no-retry | `TestPlaneDoesNotRetryPermanentPollFailure` |
| idempotent event POST retry | `TestPlaneRetriesTransientEventPostWithIdempotencyKey` |
| headless risky tool fail-closed | `TestPolicyToolStartGateBlocksClassifiedRiskWithoutApprovalPath` |
| Windows stale claim recovery | `TestReclaimStaleLockClaimRemovesOldClaim` |
| Windows fresh claim retained | `TestReclaimStaleLockClaimKeepsFreshClaim` |
| workspace prepare cancel fence | `TestSupervisorCancellationDuringWorkspacePrepareStopsBeforeRuntimeStart` |
| generated FSM daemon consumption | `TestContractsBaseline` |
| active stream handoff | `TestHTTPAIAgentClientDevelopmentV2WorkspaceScopedCreateAndThreadStream` |
| terminal late progress fence | `TestC2LateRuntimeProgressDoesNotReactivateStoppedThread` |
| generated FSM conformance | `TestVerifyGeneratedFSMFiles` |
| web approval contract | `TestAssignmentContractToolApprovalWireShape` |
| web approval session resolver | `TestSessionResolverApprovalWritesProviderInput` |
| web approval round trip | `TestResolveToolApprovalCreatesAndWaitsForDecision` |
