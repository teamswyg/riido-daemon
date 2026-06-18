package ingest

import (
	"strings"

	"github.com/teamswyg/riido-contracts/ir"
)

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
