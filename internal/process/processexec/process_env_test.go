package processexec

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestEnvOverridesPreserveParentEnvironment(t *testing.T) {
	p := New()
	proc, err := p.Start(context.Background(), process.Command{
		Executable: "/bin/sh",
		Args:       []string{"-c", `test -n "$PATH" && printf "%s" "$RIIDO_TEST_ENV"`},
		Env:        []string{"RIIDO_TEST_ENV=ok"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	stdout := drainAll(proc.Stdout(), 2*time.Second)
	if string(stdout) != "ok" {
		t.Fatalf("stdout: %q", stdout)
	}
	if status := <-proc.Exited(); status.Code != 0 {
		t.Fatalf("exit: %+v", status)
	}
}
