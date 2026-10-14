package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoStringSliceLiteral(t *testing.T) {
	got, err := goStringSliceLiteral([]string{"alpha", "beta"})
	if err != nil {
		t.Fatal(err)
	}
	if got != `[]string{"alpha", "beta"}` {
		t.Fatalf("unexpected literal: %s", got)
	}
}

func TestGoStringSliceLiteralForEmptySlice(t *testing.T) {
	got, err := goStringSliceLiteral(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "nil" {
		t.Fatalf("unexpected empty literal: %s", got)
	}
}

func TestTemplateJSONLiteralReturnsExecuteError(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "bad.go.gotmpl")
	mustWriteTemplate(t, templatePath, "package generated\n\nvar _ = {{ json .Value }}\n")
	_, err := renderSpec(map[string]any{"Value": func() {}}, templatePath)
	if err == nil {
		t.Fatal("expected template execution error")
	}
	if !strings.Contains(err.Error(), "riidogen: execute template") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustWriteTemplate(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
