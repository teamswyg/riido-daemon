package openclaw

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type fullResultCoverage struct {
	session bool
	usage   bool
	text    bool
	result  bool
}

type ndjsonCoverage struct {
	session bool
	text    bool
	log     bool
	usage   bool
}

func assertGoldenFullResultCoverage(t *testing.T, raws []agentbridge.RawEvent) {
	t.Helper()
	saw := fullResultCoverage{}
	for _, ev := range goldenTranslatedEvents(raws) {
		switch ev.Kind {
		case agentbridge.EventSessionIdentified:
			saw.session = true
		case agentbridge.EventUsageDelta:
			saw.usage = true
		case agentbridge.EventTextDelta:
			saw.text = true
		case agentbridge.EventResult:
			if ev.Result.Status == agentbridge.ResultCompleted {
				saw.result = true
			}
		}
	}
	if !saw.session || !saw.usage || !saw.text || !saw.result {
		t.Fatalf("full_result coverage gap: %+v", saw)
	}
}

func assertGoldenNDJSONCoverage(t *testing.T, raws []agentbridge.RawEvent) {
	t.Helper()
	saw := ndjsonCoverage{}
	for _, ev := range goldenTranslatedEvents(raws) {
		switch ev.Kind {
		case agentbridge.EventSessionIdentified:
			saw.session = true
		case agentbridge.EventTextDelta:
			saw.text = true
		case agentbridge.EventLog:
			saw.log = true
		case agentbridge.EventUsageDelta:
			saw.usage = true
		}
	}
	if !saw.session || !saw.text || !saw.log || !saw.usage {
		t.Fatalf("ndjson_result coverage gap: %+v", saw)
	}
}
