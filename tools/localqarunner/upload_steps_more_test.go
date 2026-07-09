package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunUploadStepsRecordsS3CommandsWithFakeAWS(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	bin := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(root, "aws.log")
	body := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '" + logPath + "'\n"
	if err := os.WriteFile(filepath.Join(bin, "aws"), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin+":"+os.Getenv("PATH"))

	cfg := uploadTestConfig("", "", "", "", "", "")
	screenshots := filepath.Join(root, "screenshots")
	if err := os.MkdirAll(screenshots, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg.productScreenshots = &screenshots
	evidence := runEvidence{ObservedAt: "2026-06-22T07:32:49Z", Status: statusPassed}
	runUploadSteps(root, cfg, &evidence)

	if len(evidence.Steps) != 8 {
		t.Fatalf("steps=%+v", evidence.Steps)
	}
	recursive := false
	for _, step := range evidence.Steps {
		if step.Status != statusPassed {
			t.Fatalf("upload step failed: %+v", step)
		}
		if !strings.Contains(step.Command, "aws s3 cp") ||
			!strings.Contains(step.Command, "--cache-control no-store") {
			t.Fatalf("unexpected upload command: %+v", step)
		}
		recursive = recursive || strings.Contains(step.Command, "--recursive")
	}
	if !recursive {
		t.Fatal("expected screenshot upload to use --recursive")
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "s3 cp .riido-local/index.html") {
		t.Fatalf("fake aws log missing dashboard upload: %s", string(data))
	}
	if evidence.Status != statusPassed {
		t.Fatalf("evidence status = %q", evidence.Status)
	}
}
