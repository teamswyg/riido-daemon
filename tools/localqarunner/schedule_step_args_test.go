package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScheduleStepArgsCreatesRepoLocalScheduleEvidence(t *testing.T) {
	root := t.TempDir()
	cfg := scheduleArgsTestConfig(filepath.Join(root, ".riido-local/evidence/local-qa-schedule.json"))
	args, id := scheduleStepArgs(root, cfg)
	joined := joinArgs(args)
	if id != "schedule-evidence" {
		t.Fatalf("id = %q, want schedule-evidence", id)
	}
	if !strings.Contains(joined, "-plist "+filepath.Join(root, ".riido-local/local-qa.plist")) {
		t.Fatalf("args missing repo-local plist: %v", args)
	}
	if strings.Contains(joined, "-inspect") {
		t.Fatalf("missing evidence should not inspect launchd: %v", args)
	}
}

func TestScheduleStepArgsInspectsExistingScheduleEvidence(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".riido-local/evidence/local-qa-schedule.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	args, id := scheduleStepArgs(root, scheduleArgsTestConfig(path))
	joined := joinArgs(args)
	if id != "schedule-inspect" || !strings.Contains(joined, "-inspect") {
		t.Fatalf("existing evidence should inspect launchd: %s %v", id, args)
	}
	if strings.Contains(joined, "-plist") {
		t.Fatalf("inspect should use scheduler default plist: %v", args)
	}
}

func scheduleArgsTestConfig(schedule string) config {
	cfg := uploadTestConfig("product.json", "", "coverage.json", "", schedule, "")
	tool, client, base := "./tools/localqaschedule", "../riido-client", "http://localhost:3000"
	agent, riido, storage := "https://staging.ai-api.riido.io", "https://staging.api.riido.io", "storage.json"
	runProduct, startClient, mutations, fixture := false, false, true, true
	workspace, team, task, first, second, comment := "", "", "", "", "", ""
	cfg.scheduleTool, cfg.clientRoot, cfg.productBaseURL = &tool, &client, &base
	cfg.productAgentHost, cfg.productRiidoHost, cfg.productStorage = &agent, &riido, &storage
	cfg.runProduct, cfg.productStartClient = &runProduct, &startClient
	cfg.productMutations, cfg.productTaskFixture = &mutations, &fixture
	cfg.productWorkspace, cfg.productTeamID, cfg.productTaskID = &workspace, &team, &task
	cfg.productAgentID1, cfg.productAgentID2, cfg.productCommentBody = &first, &second, &comment
	return cfg
}
