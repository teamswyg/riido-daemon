package mwsdbridge

func fakeDomainResponse() string {
	return `{
		"schema_version": "mws-cl-domain.v1",
		"path": "/workspace/domains/macmini-workspace.lisp",
		"domain": "macmini-workspace",
		"repositories": [{
			"name": "riido-daemon",
			"owner": "kimjooyoon",
			"visibility": "private",
			"ssot_scope": "project-daemon",
			"local_path": "/Users/teddy/github/kimjooyoon/riido-daemon",
			"remote": "https://github.com/teamswyg/riido-daemon",
			"role": "project-ssot",
			"consumes": ["mws-doc-graph", "mws-cl-domain"]
		}],
		"diagnostics": []
	}`
}

func fakeHarnessResponse() string {
	return `{
		"schema_version": "mws-harness-run.v1",
		"path": "/workspace/harness/runs.jsonl",
		"run_count": 2,
		"top_down_count": 1,
		"bottom_up_count": 1,
		"last_direction": "bottom-up",
		"next_direction": "top-down",
		"consecutive_direction_count": 1,
		"recent_directions": ["top-down", "bottom-up"],
		"diagnostics": []
	}`
}
