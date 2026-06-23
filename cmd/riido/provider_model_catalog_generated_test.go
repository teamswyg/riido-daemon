package main

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/provider/claude"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
	"github.com/teamswyg/riido-daemon/internal/provider/cursor"
	"github.com/teamswyg/riido-daemon/internal/provider/openclaw"
)

func TestGeneratedProviderModelCatalogCoversRuntimeProviders(t *testing.T) {
	for _, provider := range []string{codex.Name, cursor.Name, openclaw.Name, claude.Name} {
		models := generatedProviderRuntimeModels(provider, "")
		if len(models) <= 1 {
			t.Fatalf("generated catalog provider=%s count=%d models=%+v", provider, len(models), models)
		}
	}
}
