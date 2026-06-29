package main

import (
	"context"
	"strings"
	"testing"
)

func TestLocalQAScheduleClaimDSLOnlyChangeFails(t *testing.T) {
	repo := fixtureRepo(t)
	err := run(context.Background(), options{
		Repo: repo, Manifest: manifestPath(),
		ChangedFiles: []string{
			"docs/30-architecture/local-qa-daily-trigger.dsl.json",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "local-qa-daily-trigger-freshness") {
		t.Fatalf("expected daily trigger claim failure, got %v", err)
	}
}

func TestLocalQAScheduleClaimFullPeerChangePasses(t *testing.T) {
	repo := fixtureRepo(t)
	err := run(context.Background(), options{
		Repo: repo, Manifest: manifestPath(),
		ChangedFiles: localQAScheduleClaimPeerFiles(),
	})
	if err != nil {
		t.Fatalf("expected local QA schedule claim peers to pass, got %v", err)
	}
}

func localQAScheduleClaimPeerFiles() []string {
	return []string{
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
		".github/workflows/local-qa-runner.yml",
		"docs/30-architecture/loop-engineering/local-qa-daily-trigger.riido.json",
	}
}
