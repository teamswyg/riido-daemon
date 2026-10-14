package session

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
)

func TestBlockedProviderStdinCannotPreventCancelOrHardDeadline(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix process fixture")
	}
	for _, timeout := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		entered := make(chan struct{})
		driver := &fakeDriver{onStart: func(ctx context.Context, io ProtocolIO) error {
			close(entered)
			return io.WriteStdin(ctx, make([]byte, 2*1024*1024))
		}}
		cfg := Config{Adapter: &recordingAdapter{name: "blocked", parser: &recordingParser{}}, Process: processexec.New(), Spawn: process.Command{Executable: "/bin/sh", Args: []string{"-c", "sleep 30"}}, ProtocolDriver: driver}
		if timeout {
			cfg.HardTimeout = 100 * time.Millisecond
		}
		sess, err := Start(ctx, cfg)
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		<-entered
		if !timeout {
			time.Sleep(20 * time.Millisecond)
			sess.Cancel(nil)
		}
		select {
		case result := <-sess.Result():
			want := agentbridge.ResultCancelled
			if timeout {
				want = agentbridge.ResultTimeout
			}
			if result.Status != want {
				cancel()
				t.Fatalf("blocked input result=%+v", result)
			}
		case <-time.After(time.Second):
			cancel()
			t.Fatalf("blocked input prevented termination (timeout=%v)", timeout)
		}
		cancel()
	}
}
