package mwsdbridge

func fakeStatusResponse() string {
	return `{
		"root": "/workspace",
		"socket_path": "/tmp/mwsd.sock",
		"graph_schema_version": "mws-doc-graph.v1",
		"domain_schema_version": "mws-cl-domain.v1",
		"harness_schema_version": "mws-harness-run.v1",
		"document_count": 23,
		"repository_count": 3,
		"domain_name": "macmini-workspace",
		"harness_run_count": 2,
		"harness_next_direction": "top-down",
		"harness_recent_directions": ["top-down", "bottom-up"],
		"ssot_conflict_count": 0,
		"domain_diagnostic_count": 0,
		"harness_diagnostic_count": 0,
		"orchestration_schema_version": "mws-orchestration-snapshot.v1"
	}`
}

func fakeGraphResponse() string {
	return `{
		"schema_version": "mws-doc-graph.v1",
		"root": "/workspace",
		"stats": {
			"document_count": 23,
			"node_count": 23,
			"edge_count": 100,
			"diagnostic_count": 0,
			"error_count": 0,
			"warning_count": 0,
			"unresolved_link_count": 0
		}
	}`
}
