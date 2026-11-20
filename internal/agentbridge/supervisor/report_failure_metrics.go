package supervisor

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
)

// Keep failure evidence without error strings, request bodies or credentials.
func (m *responseMetrics) recordReportFailure(err error) {
	kind, status := runtimeSyncFailure(err)
	if status != 0 {
		kind = fmt.Sprintf("http_%d", status)
	}
	if m.reportFailureKinds == nil {
		m.reportFailureKinds = map[string]int64{}
	}
	m.reportFailureKinds[kind]++
	var correlation interface{ RequestCorrelationID() string }
	if errors.As(err, &correlation) {
		id := correlation.RequestCorrelationID()
		if _, decodeErr := hex.DecodeString(id); len(id) == 32 && decodeErr == nil {
			m.lastReportRequestID = id
		}
	}
}

func (m *responseMetrics) mergeReports(other responseMetrics) {
	m.reportAttempts += other.reportAttempts
	m.reportFailures += other.reportFailures
	m.reportTime += other.reportTime
	m.reportSuperseded += other.reportSuperseded
	if m.reportFailureKinds == nil {
		m.reportFailureKinds = map[string]int64{}
	}
	for kind, count := range other.reportFailureKinds {
		m.reportFailureKinds[kind] += count
	}
	if other.lastReportRequestID != "" {
		m.lastReportRequestID = other.lastReportRequestID
	}
}

func (m *responseMetrics) reportProgress(ctx context.Context, send func() error) {
	beforeFailures := m.reportFailures
	err := m.report(send)
	// A deliberate cancellation at a state/result boundary is not a transport
	// failure. Count it separately so operators can distinguish both outcomes.
	if errors.Is(err, context.Canceled) && ctx.Err() != nil {
		m.reportFailures = beforeFailures
		m.reportFailureKinds["cancelled"]--
		if m.reportFailureKinds["cancelled"] == 0 {
			delete(m.reportFailureKinds, "cancelled")
		}
		m.reportSuperseded++
	}
}
