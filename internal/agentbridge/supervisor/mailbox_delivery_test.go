package supervisor

import (
	"context"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/ir/ingest"
)

func deliveryEnvelope(task, text string) envelope {
	return envelope{taskEvent: &taskEventMsg{taskID: task, event: agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: text}}}
}

func TestDeliveryGroupsPreserveBoundariesAndFirstToken(t *testing.T) {
	for _, boundary := range []envelope{
		deliveryEnvelope("other", "B"),
		{taskEvent: &taskEventMsg{taskID: "one", event: agentbridge.Event{Kind: agentbridge.EventWarning, Text: "warning"}}},
		{taskEvent: &taskEventMsg{taskID: "one", event: agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateWaitingUserInput}}},
		{taskEvent: &taskEventMsg{taskID: "one", event: agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "B", ProgressKey: "structured"}}},
		{taskResult: &taskResultMsg{taskID: "one"}},
		{cancel: &cancelMsg{taskID: "one"}},
	} {
		a := &Actor{mailbox: make(chan envelope, 3)}
		a.mailbox <- deliveryEnvelope("one", "나")
		a.mailbox <- boundary
		a.mailbox <- deliveryEnvelope("one", "after")
		group, next := a.coalesceQueuedDelivery(deliveryEnvelope("one", "가"))
		if group.taskEvent.event.Text != "가나" || len(group.taskEvent.originalEvents) != 2 || next == nil {
			t.Fatal("invalid group or missing boundary")
		}
		if len(a.mailbox) != 1 {
			t.Fatal("consumed past boundary")
		}
	}
	a := &Actor{mailbox: make(chan envelope)}
	group, next := a.coalesceQueuedDelivery(deliveryEnvelope("one", "first"))
	if next != nil || group.taskEvent.event.Text != "first" {
		t.Fatal("isolated first text changed")
	}
}

func TestDeliveryGroupsAreBounded(t *testing.T) {
	for _, chunk := range []string{"x", strings.Repeat("x", 20000)} {
		a := &Actor{mailbox: make(chan envelope, 100)}
		for range 100 {
			a.mailbox <- deliveryEnvelope("one", chunk)
		}
		group, _ := a.coalesceQueuedDelivery(deliveryEnvelope("one", chunk))
		if len(group.taskEvent.originalEvents) > maxDeliveryGroupEvents || len(group.taskEvent.event.Text) > maxDeliveryGroupBytes || len(a.mailbox) == 0 {
			t.Fatal("unbounded delivery group")
		}
	}
}

type deliveryJournal struct{ events []ir.CanonicalEvent }

func (s *deliveryJournal) AppendEvents(_ context.Context, events []ir.CanonicalEvent) error {
	s.events = append(s.events, events...)
	return nil
}

func TestDeliveryGroupingPreservesOriginalJournal(t *testing.T) {
	sink := &deliveryJournal{}
	ingestor, err := ingest.New(ingest.Config{Sink: sink, RiidoDaemonVersion: "test", PolicyBundleVersion: "test", ActorKind: ir.ActorAgent, ActorID: "test"})
	if err != nil {
		t.Fatal(err)
	}
	reporter := newReporterProbe()
	a := &Actor{cfg: Config{Reporter: reporter}, mailbox: make(chan envelope, 2)}
	task := &runningTask{taskID: "one", events: &workspaceEventContext{taskID: "one", runID: "run", runtimeID: "runtime", capability: failingWorkspaceCapability(), agentIngestor: ingestor}}
	task.events.nativeConfigVersion = "test"
	task.events.capability.AdapterID = "codex"
	task.events.capability.AdapterVersion = "test"
	task.events.capability.ProtocolVersion = "v1"
	a.mailbox <- deliveryEnvelope("one", "나")
	pending := []detachedReport{}
	a.handleMailboxDelivery(context.Background(), deliveryEnvelope("one", "가"), map[string]*runningTask{"one": task}, &pending)
	if len(sink.events) != 2 {
		t.Fatalf("journal records=%d, want original 2; reports=%+v", len(sink.events), <-reporter.events)
	}
	if len(reporter.events) != 1 {
		t.Fatalf("network reports=%d, want 1", len(reporter.events))
	}
	if ev := <-reporter.events; ev.Text != "가나" {
		t.Fatalf("text=%q", ev.Text)
	}
}
