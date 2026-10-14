package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildDaemonControlPlaneUsesFileQueue(t *testing.T) {
	queueDir := t.TempDir()
	reportDir := filepath.Join(t.TempDir(), "reports")
	source, reporter, kind, err := buildDaemonControlPlane(daemonSettings{
		TaskQueueDir:  queueDir,
		TaskReportDir: reportDir,
	}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if kind != "file" {
		t.Fatalf("kind = %q", kind)
	}
	writeFileQueueTask(t, queueDir)
	claimed, err := source.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatal(err)
	}
	if claimed == nil || claimed.ID != "task-1" {
		t.Fatalf("claimed = %+v", claimed)
	}
	if err := reporter.StartTask(context.Background(), "task-1"); err != nil {
		t.Fatal(err)
	}
	err = reporter.CompleteTask(context.Background(), "task-1", agentbridge.Result{
		Status: agentbridge.ResultCompleted,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertOneReportFile(t, reportDir)
}
