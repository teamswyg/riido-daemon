package claude

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
)

func TestClaudeStartupExitHelper(t *testing.T) {
	if os.Getenv("GO_WANT_CLAUDE_STARTUP_EXIT") != "1" {
		return
	}
	_ = os.Stdin.Close()
	switch os.Getenv("RIIDO_TEST_STARTUP_OUTPUT") {
	case "stderr":
		fmt.Fprint(os.Stderr, "fixture: invalid launch setting")
	case "secret":
		fmt.Fprint(os.Stderr, "fixture: invalid launch setting ANTHROPIC_API_KEY=sk-ant-api03-"+strings.Repeat("a", 60))
	case "success":
		fmt.Fprintln(os.Stdout, `{"type":"result","subtype":"success","result":"OK"}`)
	case "empty":
	default:
		fmt.Fprintln(os.Stdout, `{"type":"result","subtype":"error_during_execution","is_error":true,"result":"fixture: Claude rejected the launch configuration"}`)
	}
	time.Sleep(100 * time.Millisecond)
	os.Exit(2)
}

func TestClaudeStartupExitPreservesProviderReason(t *testing.T) {
	for _, tc := range []struct{ name, want string }{
		{"result", "fixture: Claude rejected the launch configuration"},
		{"stderr", "fixture: invalid launch setting"},
		{"secret", "fixture: invalid launch setting"},
		{"success", "claude driver: write user frame"},
		{"empty", "claude driver: write user frame"},
	} {
		t.Run(tc.name, func(t *testing.T) { testClaudeStartupExit(t, tc.name, tc.want) })
	}
}

func testClaudeStartupExit(t *testing.T, output, want string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	driver, err := NewProtocolDriver(agentbridge.StartRequest{Prompt: strings.Repeat("x", 2*1024*1024)})
	if err != nil {
		t.Fatal(err)
	}
	sess, err := session.Start(ctx, session.Config{
		Adapter: integrationAdapter{}, Process: processexec.New(), ProtocolDriver: driver,
		Spawn: process.Command{Executable: os.Args[0], Args: []string{"-test.run=^TestClaudeStartupExitHelper$"}, Env: []string{"GO_WANT_CLAUDE_STARTUP_EXIT=1", "RIIDO_TEST_STARTUP_OUTPUT=" + output}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for event := range sess.Events() {
		if strings.Contains(event.Err+event.Text+event.Result.Error, strings.Repeat("a", 60)) {
			t.Fatal("startup event exposed the test credential")
		}
	}
	result := <-sess.Result()
	if result.Status != agentbridge.ResultFailed || !strings.Contains(result.Error, want) {
		t.Fatalf("startup error hid the child failure: status=%s error=%q", result.Status, result.Error)
	}
	if strings.Contains(result.Error, strings.Repeat("a", 60)) {
		t.Fatal("startup result exposed the test credential")
	}
}
