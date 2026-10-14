package main

import (
	"context"
	"strings"
	"testing"
)

func TestLocalQAGraphClaimCodeOnlyChangeFails(t *testing.T) {
	repo := fixtureRepo(t)
	err := run(context.Background(), options{
		Repo: repo, Manifest: manifestPath(),
		ChangedFiles: []string{
			"tools/localqarunner/closed_loop_candidate_graph.go",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "local-qa-candidate-evidence-graph") {
		t.Fatalf("expected local QA graph claim failure, got %v", err)
	}
}

func TestLocalQAGraphClaimFullPeerChangePasses(t *testing.T) {
	repo := fixtureRepo(t)
	err := run(context.Background(), options{
		Repo: repo, Manifest: manifestPath(),
		ChangedFiles: localQAGraphClaimPeerFiles(),
	})
	if err != nil {
		t.Fatalf("expected local QA graph claim peers to pass, got %v", err)
	}
}

func localQAGraphClaimPeerFiles() []string {
	return []string{
		"tools/localqarunner/closed_loop_candidate_graph.go",
		"tools/localqarunner/closed_loop_candidate_types.go",
		"tools/localqarunner/closed_loop_candidate_test.go",
		"tools/localqarunner/closed_loop_candidate_promotion_test.go",
		"tools/localqarunner/closed_loop_promotion_apply.go",
		"tools/localqarunner/closed_loop_promotion_load.go",
		"tools/localqarunner/closed_loop_promotion_test.go",
		"tools/localqarunner/closed_loop_promotion_types.go",
		"tools/localqadashboard/run_candidate.go",
		"tools/localqadashboard/run_evidence_test.go",
		"tools/localqadashboard/render_candidate_fixture_test.go",
		"tools/localqadashboard/render_expectation_test.go",
		"tools/localqadashboard/render_run_fixture_test.go",
		".github/workflows/local-qa-runner.yml",
		"docs/30-architecture/loop-engineering/local-qa-closed-loop-candidates.riido.json",
		"docs/30-architecture/local-qa-closed-loop-promotions.dsl.json",
	}
}
