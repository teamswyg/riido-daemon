package main

import (
	"os"
	"path/filepath"
	"testing"
)

func fixturePaths() []string {
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
		"tools/localqarunner/closed_loop_candidate_graph.go",
		"tools/localqarunner/closed_loop_candidate_types.go",
		"tools/localqarunner/closed_loop_candidate_test.go",
		"tools/localqadashboard/run_candidate.go",
		"tools/localqadashboard/run_evidence_test.go",
		"tools/localqadashboard/render_expectation_test.go",
		"docs/30-architecture/loop-engineering/local-qa-closed-loop-candidates.riido.json",
		"docs/30-architecture/local-qa-daily-trigger.dsl.json",
		"tools/localqaschedule/evidence_types.go",
		"tools/localqaschedule/evidence_test.go",
		"tools/localqaschedule/trigger.go",
		"tools/localqadashboard/schedule_evidence.go",
		"tools/localqadashboard/schedule_evidence_checks.go",
		"tools/localqadashboard/schedule_evidence_types.go",
		"tools/localqadashboard/schedule_evidence_freshness_test.go",
		"docs/30-architecture/local-acceptance-coverage.riido.json",
		"docs/30-architecture/executable-knowledge.riido.json",
		"docs/30-architecture/executable-knowledge.md",
		"docs/30-architecture/loop-engineering/local-qa-daily-trigger.riido.json",
		"tools/localproductacceptance/local_qa_daily_trigger.generated.json",
		"tools/localproductacceptance/local_acceptance_coverage.generated.json",
	}
}

func writeFixtureFile(t *testing.T, repo, path string) {
	t.Helper()
	writeFixtureFileWithData(t, repo, path, []byte(path+"\n"))
}

func writeFixtureFileWithData(t *testing.T, repo, path string, data []byte) {
	t.Helper()
	full := filepath.Join(repo, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
