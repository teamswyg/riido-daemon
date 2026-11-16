package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesVerifiedEvidence(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	out := filepath.Join(repo, "out.json")
	err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest, WriteDoc: true, CheckDoc: true, EvidenceOut: out})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunDetectsMissingApprovalWaitGuard(t *testing.T) {
	for _, guard := range []string{"case <-r.hardC:", "r.idleTimer.Stop()", "r.resetIdle()", "case <-r.ctx.Done():", "case req := <-r.sess.cancel:"} {
		t.Run(guard, func(t *testing.T) {
			repo := t.TempDir()
			writeFixtureRepo(t, repo)
			const sourcePath = "internal/agentbridge/session/session_tool_approval_resolver.go"
			path := filepath.Join(repo, sourcePath)
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, repo, sourcePath, strings.Replace(string(source), guard, "", 1))
			if err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest}); err == nil {
				t.Fatal("expected source drift")
			}
		})
	}
}
