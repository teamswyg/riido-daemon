package execution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBindingBoundsContinuationLimit(t *testing.T) {
	for _, limit := range []int{-1, 0, 1, 2, 3} {
		binding := fixture(t)
		binding.ContinuationLimit = limit
		data, _ := json.Marshal([]Binding{binding})
		p := filepath.Join(t.TempDir(), "binding.json")
		if err := os.WriteFile(p, data, 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := Load(p)
		if (err == nil) != (limit >= 0 && limit <= 2) {
			t.Fatalf("continuation limit %d: %v", limit, err)
		}
	}
}

func TestVerificationClassifiesMissingAndInvalidOutput(t *testing.T) {
	p, err := Prepare(context.Background(), fixture(t), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if e := p.Verify(context.Background()); e.ErrorKind != "unchanged_output" {
		t.Fatalf("unchanged: %+v", e)
	}
	file := filepath.Join(p.Directory, "answer.txt")
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if e := p.Verify(context.Background()); e.ErrorKind != "missing_output" {
		t.Fatalf("missing: %+v", e)
	}
	if err := os.Mkdir(file, 0o700); err != nil {
		t.Fatal(err)
	}
	if e := p.Verify(context.Background()); e.ErrorKind != "invalid_output" {
		t.Fatalf("invalid: %+v", e)
	}
}
