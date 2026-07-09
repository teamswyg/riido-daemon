package session

import (
	"errors"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestCommandExecutorWriteProviderInputEdges(t *testing.T) {
	proc := process.NewFakeRunning()
	exec := commandExecutor{proc: proc}
	if got := exec.writeProviderInput(nil); len(got) != 0 {
		t.Fatalf("empty input events = %+v", got)
	}
	if got := exec.writeProviderInput([]byte("hello\n")); len(got) != 0 {
		t.Fatalf("write events = %+v", got)
	}
	if string(<-proc.StdinRecv()) != "hello\n" {
		t.Fatal("stdin chunk mismatch")
	}
	fillFakeStdin(t, proc)
	got := exec.writeProviderInput([]byte("overflow"))
	assertWarning(t, got, "provider input write failed")
}

func TestCommandExecutorApprovalCommandEdges(t *testing.T) {
	proc := process.NewFakeRunning()
	cmd := agentbridge.Command{Kind: agentbridge.CommandApproveTool, ToolID: "tool-1"}
	missing := commandExecutor{proc: proc, adapter: &burstAdapter{}}
	assertWarning(t, missing.writeApprovalCommand(cmd), "provider approval command has no input builder")
	builderErr := commandExecutor{proc: proc, adapter: &approvalInputAdapter{err: errors.New("nope")}}
	assertWarning(t, builderErr.writeApprovalCommand(cmd), "provider approval command build failed")
	empty := commandExecutor{proc: proc, adapter: &approvalInputAdapter{}}
	if got := empty.writeApprovalCommand(cmd); len(got) != 0 {
		t.Fatalf("empty approval input events = %+v", got)
	}
	okProc := process.NewFakeRunning()
	ok := commandExecutor{proc: okProc, adapter: &approvalInputAdapter{input: []byte("approve\n")}}
	if got := ok.writeApprovalCommand(cmd); len(got) != 0 {
		t.Fatalf("approval write events = %+v", got)
	}
	if string(<-okProc.StdinRecv()) != "approve\n" {
		t.Fatal("approval stdin mismatch")
	}
	fillFakeStdin(t, okProc)
	assertWarning(t, ok.writeApprovalCommand(cmd), "provider approval command write failed")
}

func TestToolDecisionReasonEdges(t *testing.T) {
	if got := toolBlockReason(agentbridge.ToolStartDecision{Code: "blocked"}); got != "blocked" {
		t.Fatalf("code-only reason = %q", got)
	}
	got := toolBlockReason(agentbridge.ToolStartDecision{Code: "blocked", Reason: "unsafe"})
	if got != "blocked: unsafe" {
		t.Fatalf("combined reason = %q", got)
	}
}
