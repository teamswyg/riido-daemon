package mwsdbridge

func fakeOrchestrationResponse() string {
	return `{
		"schema_version": "mws-orchestration-snapshot.v1",
		"root": "/workspace",
		"domain_path": "/workspace/domains/macmini-workspace.lisp",
		"harness_run_path": "/workspace/harness/runs.jsonl",
		"domain_schema_version": "mws-cl-domain.v1",
		"harness_schema_version": "mws-harness-run.v1",
		"mode": "orchestration-over-choreography",
		"decision_gate": "human-approval-required",
		"decision_by": ["human"],
		"decision_llms": ["codex"],
		"provider_candidates": [
			{"id": "codex", "source_workflow": "provider-selection", "available": true, "approval_required": true},
			{"id": "claude-code", "source_workflow": "provider-selection", "available": true, "approval_required": true},
			{"id": "cursor", "source_workflow": "provider-selection", "available": true, "approval_required": true}
		],
		"recommended_provider": "codex",
		"recommended_decision_llm": "codex",
		"next_action": {
			"direction": "top-down",
			"command_surface": "mwsd harness + riido task queue + mws-viewer cockpit",
			"reason": "lift the latest bottom-up evidence into the next SSOT plan",
			"requires_human_approval": true
		},
		"top_down_count": 1,
		"bottom_up_count": 1,
		"last_direction": "bottom-up",
		"balanced": true,
		"direction_bias": false,
		"workflows": [{
			"name": "provider-selection",
			"top_down": ["goal", "constraints"],
			"bottom_up": ["capability", "history"],
			"decision_by": ["human"],
			"decision_llm": ["codex"],
			"providers": ["codex", "claude-code", "cursor"],
			"loop_steps": ["propose", "choose", "assign", "verify", "record"]
		}],
		"recent_runs": [{
			"id": "run-1",
			"direction": "bottom-up",
			"source": "mwsd",
			"provider": "rust-binary",
			"command": "verify",
			"result": "passed"
		}],
		"diagnostics": []
	}`
}
