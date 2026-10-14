package ingest

import (
	"strings"

	"github.com/teamswyg/riido-contracts/ir"
)

func redactionDraft() Draft {
	return Draft{
		Scope:                 ir.EventScopeRun,
		Type:                  ir.EventTextDelta,
		TaskID:                "task-1",
		RunID:                 "run-1",
		RuntimeID:             "runtime-1",
		CapabilityFingerprint: "cap-fp-1",
		ProviderKind:          "claude",
		ProtocolKind:          "claude-jsonl",
		ProviderVersion:       "claude 1.0",
		AdapterID:             "claude",
		AdapterVersion:        "riido-agentbridge-adapter.v1",
		ProtocolVersion:       "v1",
		NativeConfigVersion:   "nc-1",
		Payload:               redactionPayload(),
		Unknown: map[string]any{
			"raw": "RIIDO_TOKEN=" + strings.Repeat("b", 12),
		},
	}
}

func redactionPayload() map[string]any {
	return map[string]any{
		"text": "token ghp_" + strings.Repeat("a", 20),
		"nested": map[string]any{
			"url": "https://user:pass@example.com/path",
		},
	}
}
