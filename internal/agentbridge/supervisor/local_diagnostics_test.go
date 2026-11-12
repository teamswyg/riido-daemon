package supervisor

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/ir/ingest"
)

func TestProviderDiagnosticsStayLocalWhileAnswerIsDeliveredOnce(t *testing.T) {
	reporter := &cancellationDeliveryReporter{}
	actor, task, tasks := cancellationDeliveryFixture(reporter)
	sink := &deliveryJournal{}
	recorder, err := ingest.New(ingest.Config{Sink: sink, RiidoDaemonVersion: "test", PolicyBundleVersion: "test", ActorKind: ir.ActorAgent, ActorID: "test"})
	if err != nil {
		t.Fatal(err)
	}
	capability := failingWorkspaceCapability()
	capability.AdapterID, capability.AdapterVersion, capability.ProtocolVersion = "codex", "test", "v1"
	task.events = &workspaceEventContext{taskID: task.taskID, runID: "run", runtimeID: task.runtimeID, capability: capability, agentIngestor: recorder, ingestor: recorder, nativeConfigVersion: "test"}

	for range 300 {
		actor.handleTaskEvent(t.Context(), &taskEventMsg{taskID: task.taskID, event: agentbridge.Event{
			Kind: agentbridge.EventLog, LocalOnly: true, Text: "codex event decision=ignored reason=thread_mismatch",
		}}, tasks, nil)
	}
	if len(sink.events) != 300 || sink.events[0].Type != ir.EventLogLine || len(reporter.order) != 0 {
		t.Fatalf("journal records=%d, server reports=%d", len(sink.events), len(reporter.order))
	}
	actor.handleTaskEvent(t.Context(), &taskEventMsg{taskID: task.taskID, event: agentbridge.Event{
		Kind: agentbridge.EventWarning, Text: "actionable warning",
	}}, tasks, nil)
	if len(reporter.order) != 1 {
		t.Fatal("actionable warning did not reach the server")
	}
	deliverCancellationTestResult(actor, task, tasks, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "parent review"})
	actor.retryTerminalReports(t.Context(), tasks)
	if len(reporter.results) != 1 || reporter.results[0].Status != agentbridge.ResultCompleted || reporter.results[0].Output != "parent review" || len(tasks) != 0 {
		t.Fatalf("server results=%+v", reporter.results)
	}
}
