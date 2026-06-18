package saasplane

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestPlaneRetriesTransientPoll(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.failNext("/v1/agents/jykim1/poll", 1, http.StatusServiceUnavailable)
	fake.enqueue(queuedHTTPAssignment("asn-1", "hello"))
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask should retry transient poll: %v", err)
	}
	if req == nil || req.ID != "asn-1" {
		t.Fatalf("request = %+v", req)
	}
	if got := fake.requestCount("/v1/agents/jykim1/poll"); got != 2 {
		t.Fatalf("poll request count = %d, want 2", got)
	}
}

func TestPlaneDoesNotRetryPermanentPollFailure(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.failNext("/v1/agents/jykim1/poll", 1, http.StatusUnauthorized)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	_, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("ClaimTask should return permanent auth failure without retry, got %v", err)
	}
	if got := fake.requestCount("/v1/agents/jykim1/poll"); got != 1 {
		t.Fatalf("poll request count = %d, want 1", got)
	}
}
