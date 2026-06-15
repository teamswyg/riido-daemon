// Package ingest implements the daemon-side C2 EventIngestor boundary.
//
// The ingestor is the single Append API for CanonicalEvent construction:
// callers provide a draft, the ingestor assigns event identity / schema /
// actor attribution / active daemon-policy versions, validates the
// scope-aware envelope, then writes through a Sink port.
//
// CanonicalEvent schema and envelope rules are owned by riido-contracts/ir.
// This package owns the local daemon's append-time completion, validation,
// and C7 policy redaction enforcement.
//
// This package is intentionally filesystem-free. Persistence adapters live
// outside core IR packages and implement Sink.
package ingest

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
)

const EventSchemaVersionV1 = 1

// Sink is the persistence port behind EventIngestor.
type Sink interface {
	AppendEvents(context.Context, []ir.CanonicalEvent) error
}

// Config fixes server-decided event envelope fields.
type Config struct {
	Sink                Sink
	RiidoDaemonVersion  string
	PolicyBundleVersion string
	ActorKind           ir.ActorKind
	ActorID             string
	Now                 func() time.Time
	NewEventID          func(time.Time) (string, error)
}

// Ingestor owns CanonicalEvent envelope completion and validation.
type Ingestor struct {
	cfg Config
}

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

func New(cfg Config) (*Ingestor, error) {
	if cfg.Sink == nil {
		return nil, errors.New("ingest: Sink is required")
	}
	if strings.TrimSpace(cfg.RiidoDaemonVersion) == "" {
		return nil, errors.New("ingest: RiidoDaemonVersion is required")
	}
	if strings.TrimSpace(cfg.PolicyBundleVersion) == "" {
		return nil, errors.New("ingest: PolicyBundleVersion is required")
	}
	if cfg.ActorKind == "" {
		cfg.ActorKind = ir.ActorDaemon
	}
	if cfg.ActorKind != ir.ActorSystem && strings.TrimSpace(cfg.ActorID) == "" {
		return nil, errors.New("ingest: ActorID is required unless ActorKind=system")
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.NewEventID == nil {
		cfg.NewEventID = NewUUID7EventID
	}
	return &Ingestor{cfg: cfg}, nil
}

func (i *Ingestor) Append(ctx context.Context, draft Draft) (ir.CanonicalEvent, error) {
	occurredAt := draft.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = i.cfg.Now().UTC()
	} else {
		occurredAt = occurredAt.UTC()
	}
	payload, unknown, redaction := redactDraftPayload(draft.Payload, draft.Unknown)
	ev, err := i.eventFromDraft(draft, occurredAt, draft.Type, payload, unknown)
	if err != nil {
		return ir.CanonicalEvent{}, err
	}
	events := []ir.CanonicalEvent{ev}
	if redaction.hasRedaction() {
		audit, err := i.eventFromDraft(draft, occurredAt, ir.EventPolicyViolationDetected, redaction.auditPayload(ev), nil)
		if err != nil {
			return ir.CanonicalEvent{}, err
		}
		events = []ir.CanonicalEvent{ev, audit}
	}
	for _, event := range events {
		if violations := ir.ValidateEnvelope(event); len(violations) > 0 {
			return ir.CanonicalEvent{}, EnvelopeError{Violations: violations}
		}
	}
	if err := i.appendEvents(ctx, events); err != nil {
		return ir.CanonicalEvent{}, err
	}
	return ev, nil
}

func (i *Ingestor) eventFromDraft(draft Draft, occurredAt time.Time, eventType ir.EventType, payload, unknown map[string]any) (ir.CanonicalEvent, error) {
	eventID, err := i.cfg.NewEventID(occurredAt)
	if err != nil {
		return ir.CanonicalEvent{}, fmt.Errorf("ingest: new event id: %w", err)
	}
	ev := ir.CanonicalEvent{
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
	}
	return ev, nil
}

func (i *Ingestor) appendEvents(ctx context.Context, events []ir.CanonicalEvent) error {
	return i.cfg.Sink.AppendEvents(ctx, events)
}

type EnvelopeError struct {
	Violations []ir.EnvelopeViolation
}

func (e EnvelopeError) Error() string {
	if len(e.Violations) == 0 {
		return "ingest: invalid envelope"
	}
	first := e.Violations[0]
	if first.Field == "" {
		return "ingest: invalid envelope: " + first.Code
	}
	return "ingest: invalid envelope: " + first.Code + " " + first.Field
}
