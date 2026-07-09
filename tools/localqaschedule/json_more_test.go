package main

import (
	"encoding/json"
	"os"
	"testing"
)

func readScheduleJSON(t *testing.T, path string, v any) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, v); err != nil {
		t.Fatal(err)
	}
}
