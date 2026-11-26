package saasplane

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestHeartbeatRenewsSiblingAgentWhileFirstAgentIsHeld(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	client := fake.server.Client()
	transport := &heldAgentHeartbeatTransport{
		next: client.Transport, held: make(chan struct{}), release: make(chan struct{}), siblingRenewed: make(chan struct{}),
	}
	client.Transport = transport
	plane := newRuntimeBindingPlane(t, fake, func(cfg *Config) {
		cfg.RequestTimeout = 5 * time.Second
		cfg.HTTPClient = client
	})
	const runtimeID = "daemon-1:codex"
	executions := []string{"exec-a", "exec-b"}
	fake.assignmentsByID["asn-a"] = assignmentcontract.Assignment{ID: "asn-a", AgentID: "agent-a"}
	fake.assignmentsByID["asn-b"] = assignmentcontract.Assignment{ID: "asn-b", AgentID: "agent-b"}
	if err := plane.withState(t.Context(), func(s *planeState) {
		s.assignmentsByExecution[executions[0]] = fake.assignmentsByID["asn-a"]
		s.assignmentsByExecution[executions[1]] = fake.assignmentsByID["asn-b"]
		for _, id := range executions {
			s.runtimeIDsByExecution[id] = runtimeID
		}
	}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	var heartbeatErr error
	go func() {
		defer close(done)
		heartbeatErr = plane.Heartbeat(ctx, controlplane.RuntimeHeartbeat{RuntimeID: runtimeID, RunningTaskIDs: executions})
	}()
	t.Cleanup(func() { cancel(); <-done })
	select {
	case <-transport.held:
	case <-time.After(time.Second):
		t.Fatal("heartbeat never reached the held agent")
	}
	select {
	case <-transport.siblingRenewed:
	case <-time.After(time.Second):
		t.Fatal("live sibling agent lease renewal stalled behind the held agent")
	}
	close(transport.release)
	<-done
	var status httpStatusError
	if !errors.As(heartbeatErr, &status) || status.StatusCode != http.StatusForbidden {
		t.Fatalf("Heartbeat error = %v, want held agent's rejection", heartbeatErr)
	}
	want := "asn-a"
	if transport.first == "/v1/agents/agent-a/heartbeat" {
		want = "asn-b"
	}
	if len(fake.heartbeats) != 1 {
		t.Fatalf("heartbeat calls reaching server = %d, want 1", len(fake.heartbeats))
	}
	ids := fake.heartbeats[0].ActiveAssignmentIDs
	if len(ids) != 1 || ids[0] != want {
		t.Fatalf("active assignment IDs = %v, want [%s]", ids, want)
	}
}

type heldAgentHeartbeatTransport struct {
	next           http.RoundTripper
	first          string
	once           sync.Once
	held           chan struct{}
	release        chan struct{}
	siblingRenewed chan struct{}
}

func (t *heldAgentHeartbeatTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !strings.HasSuffix(req.URL.Path, "/heartbeat") {
		return t.next.RoundTrip(req)
	}
	t.once.Do(func() {
		t.first = req.URL.Path
		close(t.held)
	})
	if req.URL.Path == t.first {
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-t.release:
		}
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Status:     "403 Forbidden",
			Header:     make(http.Header),
			Body:       http.NoBody,
		}, nil
	}
	resp, err := t.next.RoundTrip(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		close(t.siblingRenewed)
	}
	return resp, err
}
