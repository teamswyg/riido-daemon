package ingest

import (
	"testing"
	"time"
)

func TestNewUUID7EventIDShape(t *testing.T) {
	id, err := NewUUID7EventID(time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 36 || id[14] != '7' {
		t.Fatalf("uuid7 shape mismatch: %q", id)
	}
}
