package logging

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestWriterLoggerFormatsLine(t *testing.T) {
	var buf bytes.Buffer
	log := NewWriterLogger(&buf)
	log.now = func() time.Time {
		return time.Date(2026, 5, 24, 12, 0, 0, 123, time.UTC)
	}

	log.Printf("daemon starting id=%s", "d-1")

	got := buf.String()
	if !strings.HasPrefix(got, "[2026-05-24T12:00:00.000000123Z] ") {
		t.Fatalf("timestamp prefix: %q", got)
	}
	if !strings.Contains(got, "daemon starting id=d-1\n") {
		t.Fatalf("message: %q", got)
	}
}
