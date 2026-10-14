package main

func fixturePaths() []string {
	return append(fixtureBasePaths(), fixtureLoopPaths()...)
}

func fixtureBasePaths() []string {
	return []string{
		"docs/30-architecture/closed-loop-maturity.dsl.json",
		"tools/localproductacceptance/closed_loop_maturity.go",
		"tools/localproductacceptance/closed_loop_maturity_counts.go",
		"tools/localproductacceptance/closed_loop_maturity_product.go",
		"tools/localproductacceptance/closed_loop_maturity_types.go",
		"tools/localproductacceptance/closed_loop_maturity_walk.go",
		"tools/localproductacceptance/closed_loop_maturity_test.go",
		"tools/localproductacceptance/closed_loop_maturity.generated.json",
		".github/workflows/local-qa-runner.yml",
		"docs/30-architecture/loop-engineering/closed-loop-maturity.riido.json",
		"docs/30-architecture/loop-engineering.md",
		"docs/executable-knowledge.md",
		"docs/30-architecture/executable-knowledge.md",
		"docs/20-domain/runtime-scheduling/invariants/local-daemon-contract.riido.json",
		"docs/20-domain/runtime-scheduling/invariants/local-daemon-contract.md",
		"internal/agentbridge/controlplane/taskdbplane/task_request_from_record.go",
		"internal/agentbridge/controlplane/taskdbplane/runtime_lease_require.go",
		"internal/agentbridge/controlplane/taskdbplane/task_claim_lease_metadata_test.go",
		"internal/agentbridge/controlplane/taskdbplane/runtime_lease_start_reject_test.go",
		".github/workflows/local-daemon-contract-evidence.yml",
	}
}

func fixtureLoopPaths() []string {
	return append(fixtureCandidatePaths(), fixtureDailyPaths()...)
}
