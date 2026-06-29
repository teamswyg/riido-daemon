package main

import (
	"os"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/provider/claude"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
	"github.com/teamswyg/riido-daemon/internal/provider/cursor"
	"github.com/teamswyg/riido-daemon/internal/provider/openclaw"
)

func TestProviderRuntimeModelsLiveEvidence(t *testing.T) {
	if os.Getenv("RIIDO_PROVIDER_MODEL_LIVE") != "1" {
		t.Skip("set RIIDO_PROVIDER_MODEL_LIVE=1 to validate local provider catalogs")
	}
	for _, provider := range []string{codex.Name, cursor.Name, openclaw.Name, claude.Name} {
		models := daemonRuntimeModels(provider)
		if len(models) <= 1 {
			t.Fatalf("%s model catalog too small: count=%d models=%+v",
				providerModelRule(provider), len(models), models)
		}
		if countRuntimeModelDefaults(models) != 1 {
			t.Fatalf("%s default model count invalid: count=%d models=%+v",
				providerModelRule(provider), countRuntimeModelDefaults(models), models)
		}
		t.Logf("%s model_count=%d default=%s", provider, len(models), runtimeDefaultModelID(models))
	}
}

func providerModelRule(provider string) string {
	return "rule=provider-runtime-model-catalog provider=" + provider +
		" requirement=\"model_count>1 and default_count==1\" source=cmd/riido"
}
