package hostintegration

import "testing"

func TestExternalToolRegistryPreservesStrongestProvenance(t *testing.T) {
	weak := validExternalToolRecord()
	weak.Provenance = ToolProvenanceAutoDetected
	weak.ExecutablePath = "/usr/local/bin/codex"

	strong := validExternalToolRecord()
	strong.Provenance = ToolProvenanceUserSelected
	strong.ExecutablePath = "/Applications/Codex.app/Contents/MacOS/codex"

	registry, err := NewExternalToolRegistry(strong)
	if err != nil {
		t.Fatalf("registry create failed: %v", err)
	}
	effective, accepted, err := registry.Register(weak)
	if err != nil {
		t.Fatalf("register weak failed: %v", err)
	}
	if accepted {
		t.Fatal("weaker provenance should not replace user-selected path")
	}
	if effective.ExecutablePath != strong.ExecutablePath {
		t.Fatalf("effective path changed: %+v", effective)
	}
}

func TestExternalToolRegistryAllowsStrongerProvenanceToReplace(t *testing.T) {
	weak := validExternalToolRecord()
	weak.Provenance = ToolProvenanceAutoDetected
	weak.ExecutablePath = "/usr/local/bin/codex"

	strong := validExternalToolRecord()
	strong.Provenance = ToolProvenanceEnvOverride
	strong.ExecutablePath = "/opt/homebrew/bin/codex"

	registry, err := NewExternalToolRegistry(weak)
	if err != nil {
		t.Fatalf("registry create failed: %v", err)
	}
	effective, accepted, err := registry.Register(strong)
	if err != nil {
		t.Fatalf("register strong failed: %v", err)
	}
	if !accepted {
		t.Fatal("stronger provenance should replace auto-detected path")
	}
	if effective.ExecutablePath != strong.ExecutablePath {
		t.Fatalf("effective path = %q, want %q", effective.ExecutablePath, strong.ExecutablePath)
	}
}
