package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePathsBuildsDefaultLaunchAgentAndLogPaths(t *testing.T) {
	cfg := testConfig()
	repo := t.TempDir()
	cfg.repo = &repo
	paths, err := resolvePaths(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if paths.repo != repo {
		t.Fatalf("repo path=%q, want %q", paths.repo, repo)
	}
	for _, want := range []string{
		filepath.Join("Library", "LaunchAgents", "io.test.plist"),
		filepath.Join(".riido-local", "logs", "local-qa-launchd.out.log"),
		filepath.Join(".riido-local", "logs", "local-qa-launchd.err.log"),
	} {
		if !strings.Contains(paths.plist+paths.stdout+paths.stderr, want) {
			t.Fatalf("paths missing %q: %+v", want, paths)
		}
	}
	if paths.launchctl != "/bin/launchctl" {
		t.Fatalf("launchctl=%q", paths.launchctl)
	}
}

func TestRunReportsPlistWriteFailure(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()
	repo := filepath.Join(dir, "repo")
	blocker := filepath.Join(dir, "not-dir")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	plist := filepath.Join(blocker, "qa.plist")
	cfg.repo = &repo
	cfg.plistPath = &plist
	errPath := filepath.Join(dir, "evidence.json")
	cfg.evidenceOut = &errPath
	_, err := run(cfg)
	if err == nil || !strings.Contains(err.Error(), "write plist") {
		t.Fatalf("expected plist write error, got %v", err)
	}
}

func TestLaunchdEvidenceForRunReportsInspectFailure(t *testing.T) {
	restore := fakeLaunchd(t, "7", "boom", nil)
	defer restore()
	cfg := testConfig()
	inspect := true
	cfg.inspect = &inspect
	_, err := launchdEvidenceForRun(cfg, schedulePaths{launchctl: "launchctl"})
	if err == nil || !strings.Contains(err.Error(), "launchctl print") {
		t.Fatalf("expected launchctl print error, got %v", err)
	}
}

func TestWriteScheduleJSONReportsEncodeError(t *testing.T) {
	err := writeJSON(filepath.Join(t.TempDir(), "out.json"), make(chan int))
	if err == nil || !strings.Contains(err.Error(), "encode schedule evidence") {
		t.Fatalf("expected encode error, got %v", err)
	}
}
