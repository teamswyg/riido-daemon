package agentbridge

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCodexContextCLIUsesDaemonExecutable(t *testing.T) {
	req := StartRequest{Prompt: "task", ContextCLIEnabled: true, Env: map[string]string{
		"RIIDO_RUN_TOKEN": "cap", "RIIDO_DAEMON_SOCKET": "/tmp/fixture.sock", "RIIDO_TASK_ID": "task-1", "RIIDO_CLI_PATH": "/caller/wrong-riido",
	}}
	req.ContextCLIExecutable = filepath.Join(t.TempDir(), "Riido QA's", "riido")
	quote := "'\"'\"'"
	if runtime.GOOS == "windows" {
		quote = "''"
	}
	got := ApplyContextCLIContract("codex", req)
	for _, want := range []string{quote, "context get --json --no-input", "work get", "meeting list", "meeting get"} {
		if !strings.Contains(got.SystemPrompt, want) {
			t.Errorf("missing daemon executable instruction %q", want)
		}
	}
	for _, unwanted := range []string{"Begin with riido context", "riido meeting", "/caller/wrong-riido"} {
		if strings.Contains(got.SystemPrompt, unwanted) {
			t.Errorf("untrusted or PATH-selected command: %q", unwanted)
		}
	}
	claude := ApplyContextCLIContract("claude", req)
	if !strings.Contains(claude.SystemPrompt, "Begin with riido context get") {
		t.Fatal("Claude allowlist command changed")
	}
}

func TestContextCLIExecutableWindowsQuoting(t *testing.T) {
	if got := quoteContextCLIExecutable(`C:\Riido QA's\riido.exe`, "windows"); got != `& 'C:\Riido QA''s\riido.exe'` {
		t.Fatalf("PowerShell literal quoting: %s", got)
	}
}

func TestCodexContextCLIQuotedExecutableSurvivesPATHReplacement(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX execution; PowerShell quoting has a separate contract test")
	}
	dir := t.TempDir()
	current := filepath.Join(dir, "Riido's $QA `literal`")
	old := filepath.Join(dir, "riido")
	for path, script := range map[string]string{
		current: "#!/bin/sh\nprintf 'CURRENT:%s\\n' \"$*\"\n",
		old:     "#!/bin/sh\nprintf 'OLD\\n'\n",
	} {
		if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	section := contextCLIInstruction("codex", current)
	var command string
	for line := range strings.SplitSeq(section, "\n") {
		if rest, ok := strings.CutPrefix(line, "- Begin with "); ok {
			command, _, _ = strings.Cut(rest, " once.")
		}
	}
	if command == "" {
		t.Fatal("CLI instruction has no executable command")
	}
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Env = []string{"PATH=" + dir}
	out, err := cmd.CombinedOutput()
	if err != nil || string(out) != "CURRENT:context get --json --no-input\n" {
		t.Fatalf("replaced PATH selected wrong executable or broke quoting: %q (%v)", out, err)
	}
}
