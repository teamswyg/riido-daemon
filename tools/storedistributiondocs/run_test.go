package main

import (
	"strings"
	"testing"
)

func TestRenderedDocsUseStoreContractChannels(t *testing.T) {
	m, err := loadManifest("../..", "docs/30-architecture/store-distribution.riido.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := loadContract("../..", m.StoreContract)
	if err != nil {
		t.Fatal(err)
	}
	docs := renderedDocs(m, c)
	body := docs["docs/30-architecture/store-distribution/architecture/mac-app-store-acceptance.md"]
	for _, want := range []string{"mac-app-store", "sandboxed-login-item-helper", "provider-non-bundling-review-note"} {
		if !strings.Contains(body, want) {
			t.Fatalf("generated Mac App Store doc missing %q\n%s", want, body)
		}
	}
}
