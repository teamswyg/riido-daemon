package supervisor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestNativeHistoryRetainedBeyondOrdinaryProgressBuffer(t *testing.T) {
	task := &runningTask{}
	for i := range 2000 {
		retainReportEvent(task, agentbridge.Event{Kind: agentbridge.EventProgress, NativeHistory: &agentbridge.NativeHistoryMessage{Ordinal: i + 1}})
	}
	if len(task.pendingEvents) != 2000 {
		t.Fatal("history was silently dropped during reporting outage")
	}
	retainReportEvent(task, agentbridge.Event{Kind: agentbridge.EventSessionIdentified, SessionID: "bound"})
	if len(task.pendingEvents) != 2001 {
		t.Fatal("history exhausted the independent lifecycle retry buffer")
	}
	a := &Actor{}
	task.terminalResult = &agentbridge.Result{Status: agentbridge.ResultCompleted}
	if a.flushTerminalReport(t.Context(), map[string]*runningTask{}, task) {
		t.Fatal("reported success before imported history")
	}
}
