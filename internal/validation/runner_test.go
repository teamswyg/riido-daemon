package validation

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestRunCommandRecordsSuccessfulValidation(t *testing.T) {
	dir := t.TempDir()
	result, err := RunCommand(context.Background(), CommandRequest{
		Command:   "printf ok",
		Workdir:   dir,
		CommandID: "command:test:validation",
		Provider:  "codex",
	}, time.Date(2026, 5, 20, 15, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("RunCommand returned error: %v", err)
	}
	if result.ExitCode != 0 || result.Result != "passed" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.ValidationGate != DefaultGate {
		t.Fatalf("unexpected validation gate: %s", result.ValidationGate)
	}
	if result.ProviderRunID != "provider-run:codex:command:test:validation" {
		t.Fatalf("unexpected provider run id: %s", result.ProviderRunID)
	}
	if result.ProviderRunResult != "passed" {
		t.Fatalf("unexpected provider run result: %s", result.ProviderRunResult)
	}
	if result.Workdir != filepath.Clean(dir) {
		t.Fatalf("unexpected workdir: %s", result.Workdir)
	}
}

func TestRunCommandRecordsFailedValidation(t *testing.T) {
	result, err := RunCommand(context.Background(), CommandRequest{
		Command:   "exit 7",
		Workdir:   t.TempDir(),
		CommandID: "command:test:validation",
		Provider:  "codex",
	}, time.Date(2026, 5, 20, 15, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("RunCommand returned error: %v", err)
	}
	if result.ExitCode != 7 || result.Result != "failed" || result.ProviderRunResult != "failed" {
		t.Fatalf("unexpected failed result: %#v", result)
	}
}
