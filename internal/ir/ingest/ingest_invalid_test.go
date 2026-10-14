package ingest

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAppendRejectsInvalidEnvelopeBeforeSink(t *testing.T) {
	sink := &memorySink{}
	ingestor, err := New(daemonTestConfig(sink, time.Time{}))
	if err != nil {
		t.Fatal(err)
	}

	_, err = ingestor.Append(context.Background(), invalidNativeConfigDraft())
	if err == nil {
		t.Fatal("expected validation error")
	}
	var envelopeErr EnvelopeError
	if !errors.As(err, &envelopeErr) {
		t.Fatalf("expected EnvelopeError, got %T %v", err, err)
	}
	if !strings.Contains(err.Error(), "NativeConfigVersion") {
		t.Fatalf("error should mention missing NCV: %v", err)
	}
	if len(sink.events) != 0 {
		t.Fatalf("invalid event must not reach sink: %+v", sink.events)
	}
	if sink.appendCalls != 0 {
		t.Fatalf("invalid event must not call sink: calls=%d", sink.appendCalls)
	}
}
