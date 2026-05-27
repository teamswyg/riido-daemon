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
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/policy"
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

func (i *Ingestor) eventFromDraft(draft Draft, occurredAt time.Time, eventType ir.EventType, payload map[string]any, unknown map[string]any) (ir.CanonicalEvent, error) {
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

func NewUUID7EventID(now time.Time) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	ms := uint64(now.UTC().UnixNano() / int64(time.Millisecond))
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)
	b[6] = (b[6] & 0x0f) | 0x70
	b[8] = (b[8] & 0x3f) | 0x80
	hexed := make([]byte, 32)
	hex.Encode(hexed, b[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", hexed[0:8], hexed[8:12], hexed[12:16], hexed[16:20], hexed[20:32]), nil
}

func copyMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func fsmVersionForEvent(eventType ir.EventType, source int) int {
	if eventType.IsTransition() {
		return source
	}
	return 0
}

type redactionSummary struct {
	patternIDs map[string]struct{}
	fields     map[string]struct{}
}

func (s *redactionSummary) add(path string, patternIDs []string) {
	if len(patternIDs) == 0 {
		return
	}
	if s.patternIDs == nil {
		s.patternIDs = map[string]struct{}{}
	}
	if s.fields == nil {
		s.fields = map[string]struct{}{}
	}
	for _, patternID := range patternIDs {
		s.patternIDs[patternID] = struct{}{}
	}
	if path != "" {
		s.fields[path] = struct{}{}
	}
}

func (s redactionSummary) hasRedaction() bool {
	return len(s.patternIDs) > 0
}

func (s redactionSummary) auditPayload(source ir.CanonicalEvent) map[string]any {
	return map[string]any{
		"category":        "SECRET_LEAK_ATTEMPTED",
		"subject":         strings.Join(sortedKeys(s.patternIDs), ","),
		"severity":        "high",
		"sourceEventID":   source.EventID,
		"sourceEventType": string(source.Type),
		"redactedFields":  sortedKeys(s.fields),
	}
}

func redactDraftPayload(payload map[string]any, unknown map[string]any) (map[string]any, map[string]any, redactionSummary) {
	var summary redactionSummary
	redactedPayload, payloadSummary := redactMap(payload, "payload")
	redactedUnknown, unknownSummary := redactMap(unknown, "unknown")
	summary.merge(payloadSummary)
	summary.merge(unknownSummary)
	return redactedPayload, redactedUnknown, summary
}

func (s *redactionSummary) merge(other redactionSummary) {
	for patternID := range other.patternIDs {
		s.add("", []string{patternID})
	}
	if len(other.fields) == 0 {
		return
	}
	if s.fields == nil {
		s.fields = map[string]struct{}{}
	}
	for field := range other.fields {
		s.fields[field] = struct{}{}
	}
}

func redactMap(in map[string]any, prefix string) (map[string]any, redactionSummary) {
	if len(in) == 0 {
		return nil, redactionSummary{}
	}
	out := make(map[string]any, len(in))
	var summary redactionSummary
	keys := make([]string, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value, child := redactValue(in[key], joinPath(prefix, key))
		out[key] = value
		summary.merge(child)
	}
	return out, summary
}

func redactValue(value any, path string) (any, redactionSummary) {
	switch v := value.(type) {
	case string:
		redacted, patternIDs := policy.RedactSecretPatterns(v, policy.SecretRedactionMarker)
		var summary redactionSummary
		summary.add(path, patternIDs)
		return redacted, summary
	case map[string]any:
		return redactMap(v, path)
	case map[string]string:
		out := make(map[string]string, len(v))
		var summary redactionSummary
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			redacted, patternIDs := policy.RedactSecretPatterns(v[key], policy.SecretRedactionMarker)
			out[key] = redacted
			summary.add(joinPath(path, key), patternIDs)
		}
		return out, summary
	case []any:
		out := make([]any, len(v))
		var summary redactionSummary
		for idx, item := range v {
			redacted, child := redactValue(item, fmt.Sprintf("%s.%d", path, idx))
			out[idx] = redacted
			summary.merge(child)
		}
		return out, summary
	case []string:
		out := make([]string, len(v))
		var summary redactionSummary
		for idx, item := range v {
			redacted, patternIDs := policy.RedactSecretPatterns(item, policy.SecretRedactionMarker)
			out[idx] = redacted
			summary.add(fmt.Sprintf("%s.%d", path, idx), patternIDs)
		}
		return out, summary
	default:
		return value, redactionSummary{}
	}
}

func joinPath(prefix string, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
