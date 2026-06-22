package main

import "testing"

func TestFirstTeamIDReadsCommonListShapes(t *testing.T) {
	for name, payload := range map[string]map[string]any{
		"teams": {"teams": []any{map[string]any{"teamId": "team-a"}}},
		"data":  {"data": []any{map[string]any{"team_id": "team-b"}}},
		"items": {"items": []any{map[string]any{"id": "team-c"}}},
		"root":  {"teamId": "team-d"},
	} {
		if got := firstTeamID(payload); got == "" {
			t.Fatalf("%s team id empty", name)
		}
	}
}

func TestFirstStringIgnoresMissingAndNonString(t *testing.T) {
	got := firstString(map[string]any{"a": 1, "b": "ok"}, "a", "b")
	if got != "ok" {
		t.Fatalf("first string=%q", got)
	}
}
