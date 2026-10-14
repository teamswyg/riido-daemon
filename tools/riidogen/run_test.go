package main

import (
	"strings"
	"testing"
)

func TestValidateRequestRejectsUnsupportedKind(t *testing.T) {
	err := validateRequest("enum-gen", "spec.json", "template.gotmpl", "out.go")
	if err == nil {
		t.Fatal("expected unsupported kind error")
	}
	if !strings.Contains(err.Error(), `unsupported kind "enum-gen"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRequestRequiresPaths(t *testing.T) {
	err := validateRequest(nativeConfigPlanKind, "", "template.gotmpl", "out.go")
	if err == nil {
		t.Fatal("expected missing path error")
	}
	if err.Error() != "riidogen: -spec, -template, and -out are required" {
		t.Fatalf("unexpected error: %v", err)
	}
}
