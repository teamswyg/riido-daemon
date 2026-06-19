package main

import "fmt"

func validateKinds(kinds []evidenceKind) []string {
	var problems []string
	if len(kinds) == 0 {
		return []string{"evidence_kinds must not be empty"}
	}
	seen := map[string]bool{}
	for _, kind := range kinds {
		if kind.ID == "" || kind.ActualID == "" {
			problems = append(problems, "evidence kind id and actual_id are required")
		}
		if len(kind.Proves) == 0 || len(kind.Forbids) == 0 {
			problems = append(problems, fmt.Sprintf("%s must include proves and forbids", kind.ID))
		}
		if seen[kind.ID] {
			problems = append(problems, fmt.Sprintf("duplicate evidence kind %q", kind.ID))
		}
		seen[kind.ID] = true
	}
	return problems
}

func validatePacketFields(m manifest) []string {
	var problems []string
	if len(m.AllowedPacketFields) == 0 {
		problems = append(problems, "allowed_packet_fields must not be empty")
	}
	for _, field := range m.ForbiddenFieldNames {
		if contains(m.AllowedPacketFields, field) {
			problems = append(problems, fmt.Sprintf("forbidden field %q is also allowed", field))
		}
	}
	problems = append(problems, requireAll("forbidden_field_names", m.ForbiddenFieldNames, requiredForbiddenFields)...)
	return problems
}
