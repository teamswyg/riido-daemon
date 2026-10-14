package providervalidation

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func loadManifest(t *testing.T, path string) providerValidationManifest {
	t.Helper()
	var manifest providerValidationManifest
	decodeJSONFile(t, path, &manifest)
	base := filepath.Dir(path)
	for _, providerFile := range manifest.ProviderFiles {
		var provider providerEvidence
		decodeJSONFile(t, filepath.Join(base, providerFile), &provider)
		manifest.Providers = append(manifest.Providers, provider)
	}
	if len(manifest.Providers) == 0 {
		t.Fatalf("manifest must include providers or provider_files: %s", path)
	}
	return manifest
}

func decodeJSONFile(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}
