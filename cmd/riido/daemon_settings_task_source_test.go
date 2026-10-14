package main

import (
	"path/filepath"
	"testing"
)

func TestLoadDaemonSettingsDefaultsTaskReportDirFromQueue(t *testing.T) {
	env := map[string]string{envTaskQueueDir: "/tmp/riido-queue"}
	settings := loadDaemonSettingsForTest(t, env)
	want := filepath.Join("/tmp/riido-queue", "reports")
	if settings.TaskReportDir != want {
		t.Fatalf("task report dir = %q, want %q", settings.TaskReportDir, want)
	}
}

func TestLoadDaemonSettingsRejectsTaskDBSourceWithFileQueue(t *testing.T) {
	env := map[string]string{
		envTaskQueueDir:     "/tmp/riido-queue",
		envTaskDBSourcePath: "/tmp/riido-task-db.json",
	}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected task DB source and file queue conflict")
	}
}
