package processexec

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestRealStdinPipe(t *testing.T) {
	p := New()
	proc, _ := p.Start(context.Background(), process.Command{
		Executable: "/bin/cat",
	})
	if err := proc.WriteStdin([]byte("hi\n")); err != nil {
		t.Fatalf("WriteStdin: %v", err)
	}
	if err := proc.CloseStdin(); err != nil {
		t.Fatalf("CloseStdin: %v", err)
	}
	stdout := drainAll(proc.Stdout(), 2*time.Second)
	if string(stdout) != "hi\n" {
		t.Fatalf("stdout: %q", stdout)
	}
	<-proc.Exited()
}
