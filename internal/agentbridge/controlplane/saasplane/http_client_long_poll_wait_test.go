package saasplane

import (
	"context"
	"testing"
	"time"
)

func TestPlaneSendsLongPollWaitMsAndExtendsRequestTimeout(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane, err := New(Config{
		BaseURL:        fake.URL(),
		DaemonID:       "daemon-1",
		DeviceID:       "device-1",
		Agents:         []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}},
		RequestTimeout: time.Second,
		LongPollWait:   2500 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req != nil {
		t.Fatalf("empty fake queue should not claim task: %+v", req)
	}
	polls := fake.pollRequestsFor("jykim1")
	if len(polls) != 1 {
		t.Fatalf("poll requests = %+v", polls)
	}
	if polls[0].WaitMs != 2500 {
		t.Fatalf("wait_ms = %d, want 2500", polls[0].WaitMs)
	}
	if plane.cfg.RequestTimeout != 7500*time.Millisecond {
		t.Fatalf("request timeout = %s, want 7.5s", plane.cfg.RequestTimeout)
	}
}
