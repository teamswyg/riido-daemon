package execution

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) Binding {
	t.Helper()
	root := t.TempDir()
	for _, argv := range [][]string{{"init", "-q", root}, {"-C", root, "config", "user.email", "test@example.invalid"}, {"-C", root, "config", "user.name", "Test"}} {
		if out, err := exec.Command("git", argv...).CombinedOutput(); err != nil {
			t.Fatalf("git: %s %v", out, err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "answer.txt"), []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("Keep repository instructions"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, argv := range [][]string{{"-C", root, "add", "."}, {"-C", root, "commit", "-qm", "fixture"}} {
		if out, err := exec.Command("git", argv...).CombinedOutput(); err != nil {
			t.Fatalf("git: %s %v", out, err)
		}
	}
	return Binding{WorkspaceID: "workspace", TaskID: "task", Repository: root, ChangedFiles: []string{"answer.txt"}, Checks: [][]string{{"git", "diff", "--check"}}}
}
func TestExecutionRequiresActualChangedOutputAndSuccessfulCheck(t *testing.T) {
	binding := fixture(t)
	prepared, err := Prepare(context.Background(), binding, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if evidence := prepared.Verify(context.Background()); !strings.Contains(evidence.Error, "not changed") {
		t.Fatalf("summary-only accepted: %+v", evidence)
	}
	if err := os.WriteFile(filepath.Join(prepared.Directory, "answer.txt"), []byte("new\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if evidence := prepared.Verify(context.Background()); evidence.Error != "" || len(evidence.Checks) != 1 || !evidence.Checks[0].Passed {
		t.Fatalf("changed output rejected: %+v", evidence)
	}
	source, _ := os.ReadFile(filepath.Join(binding.Repository, "answer.txt"))
	if string(source) != "old" {
		t.Fatal("source checkout changed")
	}
	instructions, _ := os.ReadFile(filepath.Join(prepared.Directory, "AGENTS.md"))
	if string(instructions) != "Keep repository instructions" {
		t.Fatal("repository instructions lost")
	}
	prepared.Binding.Checks = [][]string{{"git", "not-a-command"}}
	if evidence := prepared.Verify(context.Background()); !strings.Contains(evidence.Error, "verification git not-a-command failed") {
		t.Fatalf("failed check accepted: %+v", evidence)
	}
}
func TestRequiredOutputRejectsSymlinkEscape(t *testing.T) {
	binding := fixture(t)
	prepared, err := Prepare(context.Background(), binding, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(prepared.Directory, "answer.txt")
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(binding.Repository, "answer.txt"), file); err != nil {
		t.Fatal(err)
	}
	if evidence := prepared.Verify(context.Background()); !strings.Contains(evidence.Error, "escapes repository") {
		t.Fatalf("escape accepted: %+v", evidence)
	}
}
