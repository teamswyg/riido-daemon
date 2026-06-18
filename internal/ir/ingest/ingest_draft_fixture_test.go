package ingest

import "github.com/teamswyg/riido-contracts/ir"

func validNativeConfigDraft() Draft {
	draft := nativeConfigDraftBase()
	draft.NativeConfigVersion = "nc-1"
	draft.Payload = map[string]any{"files": []string{"AGENTS.md"}}
	return draft
}

func invalidNativeConfigDraft() Draft {
	return nativeConfigDraftBase()
}

func nativeConfigDraftBase() Draft {
	return Draft{
		Scope:                 ir.EventScopeRun,
		Type:                  ir.EventNativeConfigInjected,
		TaskID:                "task-1",
		RunID:                 "run-1",
		RuntimeID:             "runtime-1",
		CapabilityFingerprint: "cap-fp-1",
		ProviderKind:          "codex",
		ProtocolKind:          "codex-app-server",
		ProviderVersion:       "codex 1.0",
		AdapterID:             "codex",
		AdapterVersion:        "riido-agentbridge-adapter.v1",
		ProtocolVersion:       "v1",
	}
}
