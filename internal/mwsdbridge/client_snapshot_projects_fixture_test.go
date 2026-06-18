package mwsdbridge

func fakeProjectsResponse() string {
	return `{
		"schema_version": "mws-project-registry.v1",
		"root": "/workspace",
		"domain_path": "/workspace/domains/macmini-workspace.lisp",
		"repository_count": 1,
		"repositories": [{
			"name": "riido-daemon",
			"owner": "kimjooyoon",
			"visibility": "private",
			"ssot_scope": "project-daemon",
			"local_path": "/Users/teddy/github/kimjooyoon/riido-daemon",
			"remote": "https://github.com/teamswyg/riido-daemon",
			"role": "project-ssot",
			"consumes": ["mws-doc-graph", "mws-cl-domain"],
			"local_present": true,
			"git_present": true,
			"remote_matches": true
		}],
		"diagnostics": []
	}`
}
