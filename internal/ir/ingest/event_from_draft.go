package ingest

import (
	"fmt"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
)

func (i *Ingestor) eventFromDraft(draft Draft, occurredAt time.Time, eventType ir.EventType, payload, unknown map[string]any) (ir.CanonicalEvent, error) {
	eventID, err := i.cfg.NewEventID(occurredAt)
	if err != nil {
		return ir.CanonicalEvent{}, fmt.Errorf("ingest: new event id: %w", err)
	}
	return ir.CanonicalEvent{
		EventID:               eventID,
		OccurredAt:            occurredAt,
		EventSchemaVersion:    EventSchemaVersionV1,
		Scope:                 draft.Scope,
		Type:                  eventType,
		ActorKind:             i.cfg.ActorKind,
		ActorID:               i.cfg.ActorID,
		RiidoDaemonVersion:    i.cfg.RiidoDaemonVersion,
		PolicyBundleVersion:   i.cfg.PolicyBundleVersion,
		Payload:               copyMap(payload),
		Unknown:               copyMap(unknown),
		TaskID:                draft.TaskID,
		RunID:                 draft.RunID,
		RuntimeID:             draft.RuntimeID,
		CapabilityFingerprint: draft.CapabilityFingerprint,
		ProviderKind:          draft.ProviderKind,
		ProtocolKind:          draft.ProtocolKind,
		ProviderVersion:       draft.ProviderVersion,
		AdapterID:             draft.AdapterID,
		AdapterVersion:        draft.AdapterVersion,
		ProtocolVersion:       draft.ProtocolVersion,
		NativeConfigVersion:   draft.NativeConfigVersion,
		FSMVersion:            fsmVersionForEvent(eventType, draft.FSMVersion),
	}, nil
}
