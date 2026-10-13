package openclaw

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestNativeCatalogUsesCLIHarnessDefaultWithoutLegacyProviderConfig(t *testing.T) {
	models, err := parseModelCatalog([]byte(`{"count":3,"models":[{"key":"claude-cli/claude-opus-5","name":"Claude CLI","tags":["default"]},{"key":"openai/example","available":true},{"key":"unavailable/model","available":false}]}`))
	if err != nil || len(models) != 2 || models[0].ModelID != "claude-cli/claude-opus-5" || !models[0].IsDefault {
		t.Fatalf("native routes lost: %+v %v", models, err)
	}
}

func TestNativeCatalogInvalidPinDoesNotFallBackToInstalledCLI(t *testing.T) {
	models, err := ModelCatalog(context.Background(), agentbridge.DetectEnv{EnvOverride: map[string]string{EnvOverride: "/nonexistent/openclaw"}})
	if err == nil || len(models) != 0 {
		t.Fatalf("invalid pin selected ambient CLI: %+v %v", models, err)
	}
}

func TestNativeCatalogBoundsOutputAndRejectsNoise(t *testing.T) {
	out := &catalogOutput{}
	body := make([]byte, 70*1024)
	n, err := out.Write(body)
	if n != len(body) || err != nil || !out.overflow || len(out.body) != 64*1024 {
		t.Fatal("unbounded catalog output")
	}
	if _, err := parseModelCatalog([]byte("provider warning\n{\"models\":[]}")); err == nil {
		t.Fatal("mixed diagnostic output accepted")
	}
}
