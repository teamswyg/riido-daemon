package agentexecutionevidence

var requiredRisks = []string{
	"same-task-multiple-assignments",
	"public-repo-worktree-materialization",
	"private-repo-fail-closed",
	"async-workspace-preparation",
	"restart-recovery-skips-unresumable-active",
	"restart-recovery-provider-session-resume",
	"cancellation-watcher-release",
	"stale-pid-kill-refusal",
	"launch-path-freeze",
	"capability-ttl-redetect",
	"transient-poll-retry",
	"transient-agent-bindings-retry",
	"transient-heartbeat-retry",
	"transient-runtime-snapshot-retry",
	"permanent-poll-no-retry",
	"idempotent-event-post-retry",
	"headless-tool-risk-fail-closed",
	"windows-stale-claim-recovery",
	"windows-fresh-claim-retained",
	"workspace-prepare-cancel-fence",
	"generated-fsm-daemon-consumption",
	"web-approval-contract-consumption",
	"private-repo-url-redaction",
	"active-stream-handoff",
	"terminal-late-progress-fence",
	"generated-fsm-conformance",
	"web-approval-contract",
	"web-approval-session-resolver",
	"web-approval-round-trip",
}

var requiredRemainingBoundaries = []string{
	"private-repo-auth",
	"client-desktop-consumption",
}
