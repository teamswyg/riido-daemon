package validation

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidationRequestHelpersCoverErrorPaths(t *testing.T) {
	if _, err := normalizeCommandRequest(CommandRequest{CommandID: "id"}); err == nil {
		t.Fatal("empty command should fail")
	}
	if _, err := normalizeCommandRequest(CommandRequest{Command: "echo ok"}); err == nil {
		t.Fatal("empty command id should fail")
	}
	missing := filepath.Join(t.TempDir(), "missing")
	req := CommandRequest{Command: "echo ok", CommandID: "id", Workdir: missing}
	if _, err := normalizeCommandRequest(req); err == nil || !strings.Contains(err.Error(), "stat") {
		t.Fatalf("missing workdir error=%v", err)
	}
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveValidationWorkdir(file); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("file workdir error=%v", err)
	}
}

func TestValidationFormattingAndTimeoutHelpers(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	got, err := resolveValidationWorkdir(" ")
	if err != nil || got != cwd {
		t.Fatalf("default workdir=%q err=%v", got, err)
	}
	if validationTimeout(time.Second) != time.Second {
		t.Fatal("positive timeout should be preserved")
	}
	if !validationStartTime(time.Time{}).After(time.Time{}) {
		t.Fatal("zero start time should use now")
	}
	if sanitizeID(" A/B @# ") != "A-B---" {
		t.Fatalf("sanitize=%q", sanitizeID(" A/B @# "))
	}
	if summarize("go test", 124, nil, context.DeadlineExceeded) != "validation command timed out: go test" {
		t.Fatal("timeout summary missing")
	}
	long := summarize("cmd", 1, []byte(strings.Repeat("x", 450)), nil)
	if len(long) > len("validation command exited 1: cmd: ")+400 {
		t.Fatalf("summary too long: %d", len(long))
	}
}
