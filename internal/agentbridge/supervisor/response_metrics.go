package supervisor

import (
	"encoding/json"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Actor-owned observations. Unknown durations are -1, never a zero-latency success.
type responseMetrics struct {
	providerResults, queueObservations                            int64
	claimedAt, firstTextAt, providerStartedAt, providerFinishedAt time.Time
	textEvents, textCharacters, reportAttempts, reportFailures    int64
	queueMax, reportTime                                          time.Duration
}

func (m *responseMetrics) observe(msg *taskEventMsg, now time.Time) {
	if !msg.queuedAt.IsZero() && !now.Before(msg.queuedAt) {
		m.queueObservations++
		m.queueMax = max(m.queueMax, now.Sub(msg.queuedAt))
	}
	events := msg.originalEvents
	if len(events) == 0 {
		events = []agentbridge.Event{msg.event}
	}
	for i, ev := range events {
		if ev.Kind != agentbridge.EventTextDelta {
			continue
		}
		m.textEvents++
		m.textCharacters += int64(utf8.RuneCountInString(ev.Text))
		if m.firstTextAt.IsZero() && strings.TrimSpace(ev.Text) != "" {
			m.firstTextAt = msg.queuedAt
			if i < len(msg.originalQueuedAt) {
				m.firstTextAt = msg.originalQueuedAt[i]
			}
		}
	}
}

func (m *responseMetrics) report(send func() error) error {
	start := time.Now()
	err := send()
	m.reportAttempts++
	m.reportTime += time.Since(start)
	if err != nil {
		m.reportFailures++
	}
	return err
}

func elapsedMS(start, end time.Time) int64 {
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return -1
	}
	return end.Sub(start).Milliseconds()
}

func (m *responseMetrics) summary(provider, version string, status agentbridge.ResultStatus, ack time.Time) map[string]any {
	providerElapsed := elapsedMS(m.providerStartedAt, m.providerFinishedAt)
	if m.providerResults > 1 {
		providerElapsed = -1 // Explicit continuations are not one model turn.
	}
	queueMS, reportMS := int64(-1), int64(-1)
	if m.queueObservations > 0 {
		queueMS = m.queueMax.Milliseconds()
	}
	if m.reportAttempts > 0 {
		reportMS = m.reportTime.Milliseconds()
	}
	return map[string]any{
		"event": "agent_execution_timing", "schema_version": "agent-execution-timing.v1",
		"provider": boundedTimingProvider(provider), "daemon_version": boundedTimingVersion(version),
		"status": boundedTimingStatus(status), "claim_to_terminal_ms": elapsedMS(m.claimedAt, ack),
		"claim_to_first_text_ms": elapsedMS(m.claimedAt, m.firstTextAt),
		"provider_elapsed_ms":    providerElapsed, "provider_result_count": m.providerResults,
		"provider_clock_scope": "adapter_including_backpressure", "native_turn_ms": int64(-1),
		"post_provider_ms":    elapsedMS(m.providerFinishedAt, ack),
		"mailbox_wait_max_ms": queueMS, "report_elapsed_ms": reportMS,
		"report_attempts": m.reportAttempts, "report_failures": m.reportFailures,
		"text_delta_events": m.textEvents, "text_characters": m.textCharacters,
	}
}

func (a *Actor) logResponseMetrics(task *runningTask, ack time.Time) {
	body, err := json.Marshal(task.timing.summary(task.provider, a.cfg.RiidoDaemonVersion, task.terminalResult.Status, ack))
	if err == nil {
		log.Print(string(body))
	}
}
