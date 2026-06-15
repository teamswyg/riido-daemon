package controlplane

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestTaskReportContextRoundTripFromMetadata(t *testing.T) {
	report, ok := TaskReportContextFromMetadata(map[string]string{
		MetadataRuntimeLeaseID:               "runtime-lease:t-1:3",
		MetadataRuntimeFencingToken:          "3",
		MetadataRuntimeCapabilityFingerprint: "fp-1",
	})
	if !ok {
		t.Fatal("TaskReportContextFromMetadata returned ok=false")
	}
	if report.RuntimeLeaseID != "runtime-lease:t-1:3" || report.RuntimeFencingToken != 3 || !report.RuntimeFencingTokenSet || report.RuntimeCapabilityFingerprint != "fp-1" {
		t.Fatalf("unexpected report context: %+v", report)
	}
	ctx := ContextWithTaskReport(context.Background(), report)
	got, ok := TaskReportContextFromContext(ctx)
	if !ok || got != report {
		t.Fatalf("context round trip = %+v, %v", got, ok)
	}
}

// --- TaskSourcePort: MemorySource ---

func TestMemorySourceClaimReturnsNoTaskWhenEmpty(t *testing.T) {
	src := NewMemorySource()
	req, err := src.ClaimTask(context.Background(), "rt-1")
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if req != nil {
		t.Fatalf("expected nil request, got %+v", req)
	}
}

func TestMemorySourceQueueAndClaim(t *testing.T) {
	src := NewMemorySource()
	src.Enqueue(bridge.TaskRequest{ID: "t-1", Provider: "claude", Prompt: "hi"})
	src.Enqueue(bridge.TaskRequest{ID: "t-2", Provider: "codex", Prompt: "yo"})

	first, _ := src.ClaimTask(context.Background(), "rt-1")
	second, _ := src.ClaimTask(context.Background(), "rt-1")
	third, _ := src.ClaimTask(context.Background(), "rt-1")

	if first == nil || first.ID != "t-1" {
		t.Fatalf("first: %+v", first)
	}
	if second == nil || second.ID != "t-2" {
		t.Fatalf("second: %+v", second)
	}
	if third != nil {
		t.Fatalf("expected empty queue, got %+v", third)
	}
}

func TestMemorySourceRegisterDeregisterHeartbeat(t *testing.T) {
	src := NewMemorySource()
	now := time.Now()
	src.now = func() time.Time { return now }

	reg := RuntimeRegistration{
		DaemonID:  "d-1",
		RuntimeID: "rt-1",
		Provider:  "claude",
	}
	if err := src.RegisterRuntime(context.Background(), reg); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if rts := src.Registered(); len(rts) != 1 || rts[0].RuntimeID != "rt-1" {
		t.Fatalf("registered: %+v", rts)
	}

	now = now.Add(15 * time.Second)
	if err := src.Heartbeat(context.Background(), RuntimeHeartbeat{RuntimeID: "rt-1", SlotLimit: 2, SlotsInUse: 1, RunningTaskIDs: []string{"task-1"}}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	hb := src.Registered()[0].LastHeartbeat
	if !hb.Equal(now) {
		t.Fatalf("heartbeat: %v", hb)
	}
	if got := src.Registered()[0].SlotsInUse; got != 1 {
		t.Fatalf("slots in use after heartbeat = %d", got)
	}

	if err := src.DeregisterRuntime(context.Background(), "rt-1"); err != nil {
		t.Fatalf("Deregister: %v", err)
	}
	if rts := src.Registered(); len(rts) != 0 {
		t.Fatalf("expected empty after deregister: %+v", rts)
	}
}

func TestMemorySourceWatchCancellation(t *testing.T) {
	src := NewMemorySource()
	src.Enqueue(bridge.TaskRequest{ID: "t-1", Provider: "claude"})
	_, _ = src.ClaimTask(context.Background(), "rt-1")

	ch, err := src.WatchCancellation(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	src.Cancel("t-1", errors.New("user cancel"))
	select {
	case cause := <-ch:
		if cause == nil || cause.Error() != "user cancel" {
			t.Fatalf("cause: %v", cause)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation not delivered")
	}
}

// --- TaskReporterPort: MemoryReporter ---

func TestMemoryReporterRoundTrip(t *testing.T) {
	rep := NewMemoryReporter()
	if err := rep.StartTask(context.Background(), "t-1"); err != nil {
		t.Fatal(err)
	}
	if err := rep.ReportEvent(context.Background(), "t-1", agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "hi"}); err != nil {
		t.Fatal(err)
	}
	if err := rep.CompleteTask(context.Background(), "t-1", agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "ok"}); err != nil {
		t.Fatal(err)
	}
	rec := rep.Recorded("t-1")
	if !rec.Started || len(rec.Events) != 1 || rec.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("record: %+v", rec)
	}
}
