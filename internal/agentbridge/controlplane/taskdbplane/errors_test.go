package taskdbplane

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/pkg/failure"
)

func TestTaskDBPlaneInputErrorIsClassified(t *testing.T) {
	_, err := New(" ")
	if err == nil {
		t.Fatal("expected input error")
	}
	if !errors.Is(err, ErrTaskDBPlaneInput) {
		t.Fatalf("errors.Is(err, ErrTaskDBPlaneInput) = false for %v", err)
	}
}

func TestTaskDBPlaneRegistryErrorIsClassified(t *testing.T) {
	path := filepath.Join(t.TempDir(), "registry.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":"bad"}`), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}

	_, err := loadRuntimeRegistryOrEmpty(path)
	if err == nil {
		t.Fatal("expected registry error")
	}
	if !errors.Is(err, ErrTaskDBPlaneRegistry) {
		t.Fatalf("errors.Is(err, ErrTaskDBPlaneRegistry) = false for %v", err)
	}

	classified, ok := failure.AsClassified(err)
	if !ok {
		t.Fatal("expected classified error")
	}
	if classified.Layer() != taskDBPlaneErrorLayer {
		t.Fatalf("Layer() = %q, want %q", classified.Layer(), taskDBPlaneErrorLayer)
	}
}
