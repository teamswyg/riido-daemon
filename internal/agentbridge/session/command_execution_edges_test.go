package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestExecuteCommandsCancelProviderKillsProcess(t *testing.T) {
	running := process.NewFakeRunning()
	got := executeCommands(
		context.Background(),
		running,
		&recordingAdapter{},
		[]agentbridge.Command{{Kind: agentbridge.CommandCancelProvider}},
		0,
	)
	if len(got) != 0 {
		t.Fatalf("cancel events=%+v", got)
	}
	select {
	case <-running.KillRecv():
	case <-time.After(time.Second):
		t.Fatal("cancel command did not kill provider")
	}
}

func TestExecuteCommandsIgnoresSessionOnlyCommands(t *testing.T) {
	running := process.NewFakeRunning()
	commands := []agentbridge.Command{
		{Kind: agentbridge.CommandStartProvider},
		{Kind: agentbridge.CommandPersistSession},
		{Kind: agentbridge.CommandFlushEvents},
		{Kind: agentbridge.CommandKind("unknown")},
	}
	got := executeCommands(context.Background(), running, &recordingAdapter{}, commands, time.Second)
	if len(got) != 0 {
		t.Fatalf("ignored command events=%+v", got)
	}
	select {
	case <-running.KillRecv():
		t.Fatal("ignored commands should not kill provider")
	case <-time.After(20 * time.Millisecond):
	}
	select {
	case input := <-running.StdinRecv():
		t.Fatalf("ignored commands wrote stdin=%q", input)
	default:
	}
}
