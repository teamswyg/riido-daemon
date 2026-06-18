package hostintegration

import "time"

func consentRecord(kind ConsentKind, decision ConsentDecision) ConsentRecord {
	return ConsentRecord{
		Kind:       kind,
		Decision:   decision,
		Actor:      "user:tester",
		RecordedAt: time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
	}
}
