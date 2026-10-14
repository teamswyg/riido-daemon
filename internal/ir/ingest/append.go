package ingest

import (
	"context"

	"github.com/teamswyg/riido-contracts/ir"
)

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
	if err := validateEvents(events); err != nil {
		return ir.CanonicalEvent{}, err
	}
	if err := i.appendEvents(ctx, events); err != nil {
		return ir.CanonicalEvent{}, err
	}
	return ev, nil
}

func validateEvents(events []ir.CanonicalEvent) error {
	for _, event := range events {
		if violations := ir.ValidateEnvelope(event); len(violations) > 0 {
			return EnvelopeError{Violations: violations}
		}
	}
	return nil
}

func (i *Ingestor) appendEvents(ctx context.Context, events []ir.CanonicalEvent) error {
	return i.cfg.Sink.AppendEvents(ctx, events)
}
