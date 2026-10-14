package ingest

import (
	"time"

	"github.com/teamswyg/riido-contracts/ir"
)

// Draft is the authorized caller's event input. EventID, schema version,
// daemon/policy versions, and actor attribution are intentionally absent:
// those are assigned by Ingestor.
type Draft struct {
	OccurredAt time.Time
	Scope      ir.EventScope
	Type       ir.EventType
	Payload    map[string]any
	Unknown    map[string]any

	TaskID                string
	RunID                 string
	RuntimeID             string
	CapabilityFingerprint string
	ProviderKind          string
	ProtocolKind          string
	ProviderVersion       string
	AdapterID             string
	AdapterVersion        string
	ProtocolVersion       string
	NativeConfigVersion   string
	FSMVersion            int
}
