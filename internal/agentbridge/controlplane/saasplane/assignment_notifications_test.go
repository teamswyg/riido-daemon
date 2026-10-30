package saasplane

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestAssignmentNotificationsOneStreamNoIdleFanout(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	events := make(chan assignmentNotification, 10)
	var streams atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/daemon/assignment-events" {
			fake.handle(w, r)
			return
		}
		streams.Add(1)
		if r.Header.Get("X-Riido-Device-Secret") != "secret" {
			t.Error("stream missing device auth")
			http.Error(w, "auth", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.(http.Flusher).Flush()
		for {
			select {
			case <-r.Context().Done():
				return
			case event := <-events:
				b, _ := json.Marshal(event)
				fmt.Fprintf(w, "event: assignments\ndata: %s\n\n", b)
				w.(http.Flusher).Flush()
			}
		}
	}))
	defer server.Close()
	p, err := New(Config{BaseURL: server.URL, DaemonID: "daemon-1", DeviceID: "device-1", DeviceSecret: "secret", AssignmentEvents: true})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wake := p.AssignmentNotifications(ctx)
	if wake != p.AssignmentNotifications(ctx) {
		t.Fatal("multiple notification channels")
	}
	event := assignmentNotification{SchemaVersion: assignmentEventsSchema, Bindings: []assignmentcontract.AgentRuntimeBinding{codexRuntimeBinding("a"), codexRuntimeBinding("b"), cursorRuntimeBinding("c")}, AgentIDs: []string{"a", "b", "c"}}
	events <- event
	awaitAssignmentNotice(t, wake)
	for _, runtime := range []string{"daemon-1:codex", "daemon-1:cursor"} {
		if _, err := p.ClaimTask(ctx, runtime); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 100; i++ {
		_, _ = p.ClaimTask(ctx, "daemon-1:codex")
		_, _ = p.ClaimTask(ctx, "daemon-1:cursor")
	}
	for _, id := range []string{"a", "b", "c"} {
		if n := fake.requestCount("/v1/agents/" + id + "/poll"); n != 1 {
			t.Fatalf("idle agent %s polled %d times", id, n)
		}
	}
	if n := fake.requestCount("/v1/daemon/agent-bindings"); n != 0 {
		t.Fatalf("binding fanout=%d", n)
	}
	if streams.Load() != 1 {
		t.Fatalf("streams=%d", streams.Load())
	}
	event.AgentIDs = []string{"b"}
	events <- event
	awaitAssignmentNotice(t, wake)
	_, _ = p.ClaimTask(ctx, "daemon-1:codex")
	if fake.requestCount("/v1/agents/a/poll") != 1 || fake.requestCount("/v1/agents/b/poll") != 2 {
		t.Fatal("notification polled unrelated agent")
	}
	for _, req := range fake.pollRequestsFor("b") {
		if req.WaitMs != 0 {
			t.Fatal("notification claim long-polled")
		}
	}
}

func TestAssignmentNotificationsRetryAndNewHintDuringClaim(t *testing.T) {
	ctx := context.Background()
	var polls atomic.Int32
	var plane *Plane
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if polls.Add(1) == 1 {
			http.Error(w, "busy", http.StatusBadGateway)
			return
		}
		if polls.Load() == 2 {
			plane.acceptAssignmentNotification(ctx, assignmentNotification{Bindings: []assignmentcontract.AgentRuntimeBinding{codexRuntimeBinding("a")}, AgentIDs: []string{"a"}})
		}
		if err := json.NewEncoder(w).Encode(assignmentcontract.PollResponse{Action: assignmentcontract.PollNone}); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	plane, _ = New(Config{BaseURL: server.URL, DaemonID: "daemon-1", DeviceSecret: "secret"})
	defer plane.Close()
	plane.notices.started = true
	plane.notices.wake = make(chan struct{}, 1)
	plane.notices.pending = map[string]pendingAssignmentNotice{}
	plane.acceptAssignmentNotification(ctx, assignmentNotification{Bindings: []assignmentcontract.AgentRuntimeBinding{codexRuntimeBinding("a")}, AgentIDs: []string{"a"}})
	// 502 is retried by the existing JSON transport. A newer hint arriving during
	// the successful attempt must remain pending, despite the previous claim ACK.
	_, err := plane.ClaimTask(ctx, "daemon-1:codex")
	if err != nil {
		t.Fatal(err)
	}
	plane.notices.mu.Lock()
	_, pending := plane.notices.pending["a"]
	plane.notices.mu.Unlock()
	if !pending {
		t.Fatal("concurrent hint was lost")
	}
	_, _ = plane.ClaimTask(ctx, "daemon-1:codex")
	if polls.Load() != 3 {
		t.Fatalf("polls=%d", polls.Load())
	}
}

func TestAssignmentNotificationsAuthFailureRetainsRetryWithoutBurst(t *testing.T) {
	ctx := context.Background()
	var polls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		polls.Add(1)
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()
	p, _ := New(Config{BaseURL: server.URL, DaemonID: "daemon-1", DeviceSecret: "secret"})
	defer p.Close()
	p.notices.started = true
	p.notices.wake = make(chan struct{}, 1)
	p.notices.pending = map[string]pendingAssignmentNotice{}
	p.acceptAssignmentNotification(ctx, assignmentNotification{Bindings: []assignmentcontract.AgentRuntimeBinding{codexRuntimeBinding("a")}, AgentIDs: []string{"a"}})
	if _, err := p.ClaimTask(ctx, "daemon-1:codex"); err == nil {
		t.Fatal("expected failure")
	}
	for i := 0; i < 100; i++ {
		_, _ = p.ClaimTask(ctx, "daemon-1:codex")
	}
	if polls.Load() != 1 {
		t.Fatalf("retry burst=%d", polls.Load())
	}
	p.notices.mu.Lock()
	notice, exists := p.notices.pending["a"]
	p.notices.mu.Unlock()
	if !exists || notice.retryAfter.IsZero() {
		t.Fatal("failed notification not retained")
	}
}

func TestAssignmentNotificationsExplicitLegacyFallback(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	p, _ := New(Config{BaseURL: fake.URL(), DaemonID: "daemon-1", DeviceSecret: "secret", AssignmentEvents: true})
	defer p.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wake := p.AssignmentNotifications(ctx)
	awaitAssignmentNotice(t, wake)
	p.notices.mu.Lock()
	legacy := p.notices.legacy
	p.notices.mu.Unlock()
	if !legacy {
		t.Fatal("missing legacy fallback")
	}
}

func awaitAssignmentNotice(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("missing notification")
	}
}

func TestAssignmentNotificationsCapacityDefersWithoutLosingWork(t *testing.T) {
	ctx := context.Background()
	fake := newFakeAssignmentServer(t)
	p, err := New(Config{BaseURL: fake.URL(), DaemonID: "daemon-1", DeviceSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	p.notices.started = true
	p.notices.wake = make(chan struct{}, 1)
	p.notices.pending = map[string]pendingAssignmentNotice{}
	p.acceptAssignmentNotification(ctx, assignmentNotification{Bindings: []assignmentcontract.AgentRuntimeBinding{codexRuntimeBinding("a")}, AgentIDs: []string{"a"}})
	_, err = p.ClaimTask(controlplane.ContextWithClaimCapacity(ctx, false), "daemon-1:codex")
	if err != nil {
		t.Fatal(err)
	}
	if fake.requestCount("/v1/agents/a/poll") != 0 {
		t.Fatal("claimed with no capacity")
	}
	_, err = p.ClaimTask(ctx, "daemon-1:codex")
	if err != nil {
		t.Fatal(err)
	}
	if fake.requestCount("/v1/agents/a/poll") != 1 {
		t.Fatal("hint lost while busy")
	}
}

func TestAssignmentNotificationsReconnectRestoresMissedHint(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	var streams atomic.Int32
	firstRead := make(chan struct{})
	releaseFirst := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/daemon/assignment-events" {
			fake.handle(w, r)
			return
		}
		n := streams.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		event := assignmentNotification{SchemaVersion: assignmentEventsSchema, Bindings: []assignmentcontract.AgentRuntimeBinding{codexRuntimeBinding("jykim1")}, AgentIDs: []string{"jykim1"}}
		b, _ := json.Marshal(event)
		fmt.Fprintf(w, "event: assignments\ndata: %s\n\n", b)
		w.(http.Flusher).Flush()
		if n == 1 {
			close(firstRead)
			select {
			case <-releaseFirst:
			case <-r.Context().Done():
			}
			return
		}
		<-r.Context().Done()
	}))
	defer server.Close()
	p, err := New(Config{BaseURL: server.URL, DaemonID: "daemon-1", DeviceID: "device-1", DeviceSecret: "secret", AssignmentEvents: true})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wake := p.AssignmentNotifications(ctx)
	<-firstRead
	awaitAssignmentNotice(t, wake)
	if req, err := p.ClaimTask(ctx, "daemon-1:codex"); err != nil || req != nil {
		t.Fatalf("initial: %v %v", req, err)
	}
	// Work arrives while disconnected; its individual notification is never sent.
	fake.enqueue(assignmentLifecycleFixture("missed", "hello", "lease"))
	close(releaseFirst)
	awaitAssignmentNotice(t, wake)
	req, err := p.ClaimTask(ctx, "daemon-1:codex")
	if err != nil || req == nil {
		t.Fatalf("lost reconnect work: %v %v", req, err)
	}
	if streams.Load() != 2 {
		t.Fatalf("streams=%d", streams.Load())
	}
	for i := 0; i < 20; i++ {
		again, err := p.ClaimTask(ctx, "daemon-1:codex")
		if err != nil || again != nil {
			t.Fatalf("duplicate claim: %v %v", again, err)
		}
	}
	if fake.requestCount("/v1/agents/jykim1/poll") != 2 {
		t.Fatal("idle duplicate poll")
	}
}
