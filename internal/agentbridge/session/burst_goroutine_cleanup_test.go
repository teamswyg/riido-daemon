package session

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSessionGoroutineCleanupAfterCompletion(t *testing.T) {
	runtime.GC()
	time.Sleep(20 * time.Millisecond)
	baseline := runtime.NumGoroutine()

	for cycle := range 5 {
		sess := startBurstCleanupSession(t, cycle)
		go closeAfterEvents(sess, make(chan struct{}))
		go completeBurstSession(sess)
		<-sess.Result()
	}

	final := waitNumGoroutine(baseline+2, 2*time.Second)
	if final > baseline+2 {
		t.Fatalf("session goroutine leak: baseline=%d final=%d", baseline, final)
	}
}

func startBurstCleanupSession(t *testing.T, cycle int) *Session {
	t.Helper()
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	sess, err := Start(context.Background(), Config{
		TaskID:    "t-" + strings.Repeat("x", cycle+1),
		RuntimeID: "rt-1",
		Adapter:   &burstAdapter{},
		Process:   fake,
		Spawn:     process.Command{Executable: "x"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return sess
}

func completeBurstSession(sess *Session) {
	running := sess.runningForTest()
	running.EmitStdout([]byte("DONE"))
	running.EmitExit(0, nil)
}
