package supervisor

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTimingUsesProducerObservationAndPreservesOriginalCounts(t *testing.T) {
	start := time.Now()
	a := &Actor{mailbox: make(chan envelope, 2)}
	first := deliveryEnvelope("task", " \n")
	first.taskEvent.queuedAt = start.Add(time.Second)
	second := deliveryEnvelope("task", "한글")
	second.taskEvent.queuedAt = start.Add(2 * time.Second)
	a.mailbox <- second
	group, boundary := a.coalesceQueuedDelivery(first)
	if boundary != nil {
		t.Fatal("unexpected boundary")
	}
	m := responseMetrics{claimedAt: start, providerStartedAt: start.Add(100 * time.Millisecond), providerFinishedAt: start.Add(3 * time.Second)}
	m.observe(group.taskEvent, start.Add(4*time.Second))
	s := m.summary("codex", "v0.0.94", agentbridge.ResultCompleted, start.Add(5*time.Second))
	for key, want := range map[string]int64{"claim_to_first_text_ms": 2000, "claim_to_terminal_ms": 5000, "provider_elapsed_ms": 2900, "post_provider_ms": 2000, "mailbox_wait_max_ms": 3000, "text_delta_events": 2, "text_characters": 4} {
		if s[key] != want {
			t.Fatalf("%s=%v, want %d", key, s[key], want)
		}
	}
}

func TestTimingUnknownValuesAndPrivacy(t *testing.T) {
	m := responseMetrics{}
	s := m.summary("customer-provider-secret", "credential-secret", "raw-error-secret", time.Now())
	for _, key := range []string{"claim_to_terminal_ms", "provider_elapsed_ms", "post_provider_ms", "claim_to_first_text_ms", "mailbox_wait_max_ms", "report_elapsed_ms", "native_turn_ms"} {
		if s[key] != int64(-1) {
			t.Fatalf("missing %s incorrectly reported as success", key)
		}
	}
	body, _ := json.Marshal(s)
	if strings.Contains(string(body), "secret") || strings.Contains(string(body), "task_id") {
		t.Fatal("unbounded metadata leaked")
	}
	if elapsedMS(time.Now(), time.Now().Add(-time.Hour)) != -1 {
		t.Fatal("clock inversion accepted")
	}
}

func TestTimingReportFailuresAreCountedWithoutErrorContents(t *testing.T) {
	m := responseMetrics{}
	_ = m.report(func() error { return errors.New("token-secret") })
	_ = m.report(func() error { return nil })
	if m.reportAttempts != 2 || m.reportFailures != 1 {
		t.Fatal("retry accounting lost")
	}
	body, _ := json.Marshal(m.summary("claude", "unknown", agentbridge.ResultFailed, time.Now()))
	if strings.Contains(string(body), "token-secret") {
		t.Fatal("error leaked")
	}
}
