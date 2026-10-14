package controlplane

import (
	"context"
	"errors"
	"testing"
)

func TestControlPlaneRuntimeErrorIsClassified(t *testing.T) {
	src := NewMemorySource()

	err := src.RegisterRuntime(context.Background(), RuntimeRegistration{})
	if err == nil {
		t.Fatal("expected runtime error")
	}
	if !errors.Is(err, ErrControlPlaneRuntime) {
		t.Fatalf("errors.Is(err, ErrControlPlaneRuntime) = false for %v", err)
	}
}

func TestControlPlaneQueueErrorIsClassified(t *testing.T) {
	_, err := NewFileQueueSource(t.TempDir() + "/missing")
	if err == nil {
		t.Fatal("expected queue error")
	}
	if !errors.Is(err, ErrControlPlaneQueue) {
		t.Fatalf("errors.Is(err, ErrControlPlaneQueue) = false for %v", err)
	}
}
