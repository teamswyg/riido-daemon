package supervisor

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/ir/ingest"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestCancelledDeliveryPreservesLocalSupersessionJournal(t *testing.T) {
	reporter := &cancellationDeliveryReporter{failEvent: true}
	a, task, tasks := cancellationDeliveryFixture(reporter)
	sink := &deliveryJournal{}
	recorder, err := ingest.New(ingest.Config{Sink: sink, RiidoDaemonVersion: "test", PolicyBundleVersion: "test", ActorKind: ir.ActorAgent, ActorID: "test"})
	if err != nil {
		t.Fatal(err)
	}
	task.events = &workspaceEventContext{taskID: task.taskID, runID: "run", runtimeID: task.runtimeID, capability: failingWorkspaceCapability(), agentIngestor: recorder, nativeConfigVersion: "test"}
	task.events.capability.AdapterID = "codex"
	task.events.capability.AdapterVersion = "test"
	task.events.capability.ProtocolVersion = "v1"
	ev := agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}
	a.appendTaskProviderEvent(context.Background(), task, ev)
	a.reportTaskEvent(context.Background(), task, ev)
	cancelDeliveryClaim(a, task, tasks)
	if len(sink.events) != 2 || sink.events[0].Type != ir.EventStatusUpdate || sink.events[1].Type != ir.EventLogLine {
		t.Fatalf("original/supersession journal events=%+v", sink.events)
	}
	if len(task.supersededEvents) != 1 || len(task.pendingEvents) != 0 {
		t.Fatal("superseded evidence not preserved independently")
	}
	if len(reporter.order) != 1 {
		t.Fatal("local diagnostic made an extra HTTP call")
	}
}

func TestCancelledDeliveryRetainsSessionPinsAndOtherEventOrder(t *testing.T) {
	reporter := &cancellationDeliveryReporter{}
	a, task, tasks := cancellationDeliveryFixture(reporter)
	task.pendingEvents = []agentbridge.Event{
		{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning},
		{Kind: agentbridge.EventSessionIdentified, SessionID: "native"},
		{Kind: agentbridge.EventWarning, Text: "diagnostic"},
	}
	cancelDeliveryClaim(a, task, tasks)
	deliverCancellationTestResult(a, task, tasks, agentbridge.Result{Status: agentbridge.ResultCancelled})
	if len(task.pendingEvents) != 2 || task.pendingEvents[0].Kind != agentbridge.EventSessionIdentified || len(reporter.results) != 0 {
		t.Fatal("non-obsolete records discarded or terminal overtook them")
	}
	a.retryEventReports(context.Background(), tasks)
	a.retryTerminalReports(context.Background(), tasks)
	if len(reporter.results) != 0 {
		t.Fatal("terminal overtook warning")
	}
	a.retryEventReports(context.Background(), tasks)
	a.retryTerminalReports(context.Background(), tasks)
	if len(tasks) != 0 || len(reporter.accepted) != 1 {
		t.Fatal("retained records did not drain")
	}
}

func TestCancelledDeliveryBoundsLateSupersededNotifications(t *testing.T) {
	reporter := &cancellationDeliveryReporter{}
	a, task, tasks := cancellationDeliveryFixture(reporter)
	cancelDeliveryClaim(a, task, tasks)
	for range 64 {
		a.reportTaskEvent(context.Background(), task, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning})
	}
	if len(task.supersededEvents) != maxPendingReportEvents || len(task.pendingEvents) != 0 || len(reporter.order) != 0 {
		t.Fatal("late cancelled notifications were unbounded or sent to the server")
	}
}

func TestShutdownPreservesPreparedUnknownTerminal(t *testing.T) {
	reporter := &cancellationDeliveryReporter{failTerminal: true}
	a, task, tasks := cancellationDeliveryFixture(reporter)
	deliverCancellationTestResult(a, task, tasks, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "possibly committed answer"})
	a.cancelInFlightTask(context.Background(), tasks, task, time.Now())
	if len(reporter.results) != 2 || reporter.results[0].Status != agentbridge.ResultCompleted || reporter.results[1].Status != agentbridge.ResultCompleted || reporter.results[0].Output != reporter.results[1].Output || len(tasks) != 0 {
		t.Fatal("shutdown replaced a prepared unknown-ACK terminal")
	}
}

func TestDuplicateResultCannotRewritePreparedUnknownTerminal(t *testing.T) {
	reporter := &cancellationDeliveryReporter{failTerminal: true}
	a, task, tasks := cancellationDeliveryFixture(reporter)
	deliverCancellationTestResult(a, task, tasks, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "possibly committed answer"})
	deliverCancellationTestResult(a, task, tasks, agentbridge.Result{Status: agentbridge.ResultFailed, Error: "late duplicate result"})
	if len(reporter.results) != 2 || reporter.results[1].Status != agentbridge.ResultCompleted || reporter.results[1].Output != "possibly committed answer" || len(tasks) != 0 {
		t.Fatal("duplicate result replaced immutable delivery outcome")
	}
}

func TestLateCancellationAfterTerminalACKDoesNotReportAgain(t *testing.T) {
	reporter := &cancellationDeliveryReporter{}
	a, task, tasks := cancellationDeliveryFixture(reporter)
	deliverCancellationTestResult(a, task, tasks, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "acknowledged answer"})
	cancelDeliveryClaim(a, task, tasks)
	a.retryTerminalReports(context.Background(), tasks)
	if len(reporter.results) != 1 || reporter.results[0].Status != agentbridge.ResultCompleted || len(tasks) != 0 {
		t.Fatal("late cancellation rewrote an acknowledged completion")
	}
}

func TestSupersessionDiagnosticDoesNotLeakUnknownPhase(t *testing.T) {
	reporter := &cancellationDeliveryReporter{}
	a, task, tasks := cancellationDeliveryFixture(reporter)
	cancelDeliveryClaim(a, task, tasks)
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previous) })
	ev := agentbridge.Event{Kind: agentbridge.EventToolApprovalNeeded, Phase: "private-unvalidated-phase-content"}
	a.reportTaskEvent(context.Background(), task, ev)
	kind, phase := supersededNotificationClassification(ev)
	if kind != "tool_approval_needed" || phase != "unknown" || strings.Contains(output.String(), "private-unvalidated-phase-content") {
		t.Fatalf("unbounded diagnostic: kind=%s phase=%s log=%q", kind, phase, output.String())
	}
	if len(task.supersededEvents) != 1 || task.supersededEvents[0].Phase != ev.Phase {
		t.Fatal("original private event evidence changed")
	}
}

func TestUnattemptedCancellationCorrectsJournalAndArchive(t *testing.T) {
	reporter := &cancellationDeliveryReporter{failEvent: true}
	a, task, tasks := cancellationDeliveryFixture(reporter)
	store := workdir.NewFSAdapter(t.TempDir())
	ws, err := store.Prepare(workdir.TaskID{Workspace: "workspace", Task: "task", Run: "run"})
	if err != nil {
		t.Fatal(err)
	}
	a.cfg.Workdir = store
	task.workspace = &ws
	sink := &deliveryJournal{}
	recorder, err := ingest.New(ingest.Config{Sink: sink, RiidoDaemonVersion: "test", PolicyBundleVersion: "test", ActorKind: ir.ActorDaemon, ActorID: "daemon"})
	if err != nil {
		t.Fatal(err)
	}
	task.events = &workspaceEventContext{taskID: task.taskID, runID: "run", runtimeID: task.runtimeID, capability: failingWorkspaceCapability(), ingestor: recorder, agentIngestor: recorder, nativeConfigVersion: "test"}
	task.events.capability.AdapterID = "codex"
	task.events.capability.AdapterVersion = "test"
	task.events.capability.ProtocolVersion = "v1"
	a.reportTaskEvent(context.Background(), task, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning})
	deliverCancellationTestResult(a, task, tasks, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "original provider evidence", Usage: agentbridge.Usage{PromptTokens: 13}})
	readArchive := func() workdir.ArchiveRecord {
		t.Helper()
		data, readErr := os.ReadFile(filepath.Join(ws.Root, "archive.json"))
		if readErr != nil {
			t.Fatal(readErr)
		}
		var record workdir.ArchiveRecord
		if decodeErr := json.Unmarshal(data, &record); decodeErr != nil {
			t.Fatal(decodeErr)
		}
		return record
	}
	if readArchive().ResultStatus != "completed" || len(reporter.results) != 0 || task.terminalDeliveryAttempted {
		t.Fatal("expected locally recorded but unattempted provider completion")
	}
	reporter.cancelling = true
	cancelDeliveryClaim(a, task, tasks)
	a.retryTerminalReports(context.Background(), tasks)
	if len(reporter.accepted) != 1 || reporter.accepted[0].Status != agentbridge.ResultCancelled || readArchive().ResultStatus != "cancelled" || len(tasks) != 0 {
		t.Fatal("wire/archive did not settle cancellation")
	}
	if reporter.accepted[0].Usage.PromptTokens != 13 || reporter.accepted[0].Output != "original provider evidence" {
		t.Fatal("original output/usage were discarded")
	}
	done, cancelled := -1, -1
	for i, event := range sink.events {
		if event.Type == ir.EventRunReportedDone {
			done = i
		}
		if event.Type == ir.EventTaskCancelled {
			cancelled = i
		}
	}
	if done < 0 || cancelled <= done {
		t.Fatal("journal erased provider evidence or lacks later cancellation correction")
	}
}
