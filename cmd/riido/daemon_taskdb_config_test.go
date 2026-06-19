package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/taskdbplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestBuildDaemonControlPlaneUsesTaskDBSource(t *testing.T) {
	taskDBPath := filepath.Join(t.TempDir(), "task-db.json")
	db := taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                  "task-1",
			ProjectID:           "workspace-1",
			State:               task.StateQueued,
			Title:               "run from task DB",
			RecommendedProvider: "codex",
		}},
	}
	if err := taskdb.SaveTaskDB(taskDBPath, db); err != nil {
		t.Fatal(err)
	}

	source, reporter, kind, err := buildDaemonControlPlane(daemonSettings{TaskDBSourcePath: taskDBPath}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if kind != "taskdb" {
		t.Fatalf("kind = %q", kind)
	}
	if _, ok := source.(*taskdbplane.Plane); !ok {
		t.Fatalf("source type = %T", source)
	}
	if _, ok := reporter.(*taskdbplane.Plane); !ok {
		t.Fatalf("reporter type = %T", reporter)
	}
	claimed, err := source.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatal(err)
	}
	if claimed == nil || claimed.ID != "task-1" || claimed.Provider != "codex" {
		t.Fatalf("claimed = %+v", claimed)
	}
}

func TestBuildDaemonControlPlaneRejectsTaskDBSourceWithReportDir(t *testing.T) {
	_, _, _, err := buildDaemonControlPlane(daemonSettings{
		TaskDBSourcePath: filepath.Join(t.TempDir(), "task-db.json"),
		TaskReportDir:    t.TempDir(),
	}, time.Time{})
	if err == nil {
		t.Fatal("expected task DB source and report dir conflict")
	}
}

func TestBuildDaemonControlPlaneRejectsReportDirWithoutQueueDir(t *testing.T) {
	_, _, _, err := buildDaemonControlPlane(daemonSettings{TaskReportDir: t.TempDir()}, time.Time{})
	if err == nil {
		t.Fatal("expected error for report dir without queue dir")
	}
}
