package main

import "testing"

func TestParseCursorModels(t *testing.T) {
	rows := parseCursorModels("Available models\n\nauto - Auto (current)\ngpt-5.2 - GPT-5.2\n")
	if len(rows) != 2 || rows[0].ModelID != "cursor-auto" || rows[1].Label != "GPT-5.2" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestParseClaudeModels(t *testing.T) {
	body := "--model <model> use 'fable', 'opus', or 'sonnet' and 'claude-fable-5'.\n  -n, --name"
	rows := parseClaudeModels(body)
	if len(rows) != 4 || rows[2].ModelID != "sonnet" {
		t.Fatalf("rows = %+v", rows)
	}
}
