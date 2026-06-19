package providervalidation

import (
	"os"
	"path/filepath"
	"testing"
)

type providerValidationContext struct {
	manifest      providerValidationManifest
	docText       string
	securityText  string
	runtimeText   string
	migrationText string
	providers     map[string]providerEvidence
}

func loadProviderValidationContext(t *testing.T) providerValidationContext {
	t.Helper()
	root := filepath.Join("..", "..")
	ctx := providerValidationContext{
		manifest:      loadManifest(t, filepath.Join(root, "docs", "30-architecture", "provider-validation-matrix.riido.json")),
		docText:       readText(t, filepath.Join(root, "docs", "30-architecture", "integration-matrix.md")),
		securityText:  readText(t, filepath.Join(root, "docs", "20-domain", "security.md")),
		runtimeText:   readText(t, filepath.Join(root, "docs", "20-domain", "provider-runtime.md")),
		migrationText: readText(t, filepath.Join(root, "docs", "migration", "daemon.md")),
		providers:     map[string]providerEvidence{},
	}
	for _, provider := range ctx.manifest.Providers {
		if provider.Provider == "" {
			t.Fatalf("provider row has empty provider: %+v", provider)
		}
		if _, exists := ctx.providers[provider.Provider]; exists {
			t.Fatalf("duplicate provider row %q", provider.Provider)
		}
		ctx.providers[provider.Provider] = provider
	}
	return ctx
}

func readText(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
