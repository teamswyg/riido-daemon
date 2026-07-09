package main

import (
	"path/filepath"
	"testing"
)

func TestRunScheduleInspectStepAppendsInjectedRunnerResult(t *testing.T) {
	original := runLocalQAScheduleStep
	t.Cleanup(func() { runLocalQAScheduleStep = original })
	var capturedRoot, capturedID, capturedExe string
	var capturedArgs []string
	runLocalQAScheduleStep = func(root, id, exe string, args ...string) stepEvidence {
		capturedRoot, capturedID, capturedExe = root, id, exe
		capturedArgs = append([]string{}, args...)
		return stepEvidence{ID: id, Status: statusPassed, Command: exe}
	}
	root := t.TempDir()
	cfg := scheduleArgsTestConfig(filepath.Join(root, "schedule.json"))
	evidence := runEvidence{}
	runScheduleInspectStep(root, cfg, &evidence)
	if len(evidence.Steps) != 1 || evidence.Steps[0].ID != "schedule-evidence" {
		t.Fatalf("steps=%+v", evidence.Steps)
	}
	if capturedRoot != root || capturedID != "schedule-evidence" || capturedExe != "go" {
		t.Fatalf("captured root=%q id=%q exe=%q", capturedRoot, capturedID, capturedExe)
	}
	if len(capturedArgs) == 0 || capturedArgs[0] != "run" {
		t.Fatalf("unexpected args: %v", capturedArgs)
	}
}
