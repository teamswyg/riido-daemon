package failure

import (
	"errors"
	"io"
	"testing"
)

func TestSentinelClassificationSurvivesWrapping(t *testing.T) {
	sentinel := NewSentinel("daemon", "socket")
	err := Wrap(sentinel, "listen", "bind local socket", io.ErrClosedPipe)

	if !errors.Is(err, sentinel) {
		t.Fatalf("errors.Is() = false, want true")
	}
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("wrapped cause was not preserved")
	}
}

func TestClassifiedAndOperationalInterfaces(t *testing.T) {
	sentinel := NewSentinel("taskdbplane", "lease")
	err := New(sentinel, "claim", "runtime lease mismatch")

	classified, ok := AsClassified(err)
	if !ok {
		t.Fatal("AsClassified() = false, want true")
	}
	if classified.Layer() != "taskdbplane" || classified.Kind() != "lease" {
		t.Fatalf("classification = %s/%s", classified.Layer(), classified.Kind())
	}

	operational, ok := AsOperational(err)
	if !ok {
		t.Fatal("AsOperational() = false, want true")
	}
	if operational.Op() != "claim" {
		t.Fatalf("Op() = %q, want %q", operational.Op(), "claim")
	}
}

func TestSentinelMatchesClassifiedTarget(t *testing.T) {
	sentinel := NewSentinel("daemon", "usage")
	err := Wrap(sentinel, "parse", "bad flag", nil)

	if !errors.Is(sentinel, err) {
		t.Fatalf("sentinel should match classified error target")
	}
}
