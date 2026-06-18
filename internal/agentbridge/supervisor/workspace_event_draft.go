package supervisor

import (
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/ir/ingest"
)

func (e *workspaceEventContext) draft(eventType ir.EventType, nativeConfigVersion string, payload map[string]any) ingest.Draft {
	return ingest.Draft{
		Scope:                 ir.EventScopeRun,
		Type:                  eventType,
		Payload:               payload,
		TaskID:                e.taskID,
		RunID:                 e.runID,
		RuntimeID:             e.runtimeID,
		CapabilityFingerprint: e.capability.CapabilityFingerprint,
		ProviderKind:          e.capability.Provider,
		ProtocolKind:          e.capability.ProtocolKind,
		ProviderVersion:       e.capability.Version,
		AdapterID:             e.capability.AdapterID,
		AdapterVersion:        e.capability.AdapterVersion,
		ProtocolVersion:       e.capability.ProtocolVersion,
		NativeConfigVersion:   nativeConfigVersion,
	}
}

func (e *workspaceEventContext) transitionDraft(eventType ir.EventType, nativeConfigVersion string, payload map[string]any) ingest.Draft {
	draft := e.draft(eventType, nativeConfigVersion, payload)
	draft.FSMVersion = task.FSMSchemaVersion
	return draft
}

func eventNativeConfigVersion(events *workspaceEventContext) string {
	if events == nil {
		return ""
	}
	return events.nativeConfigVersion
}
