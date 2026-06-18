package hostintegration

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStoreClientMetadataRejectsInvalidChannel(t *testing.T) {
	_, err := BuildServerFacingClientMetadata(ServerFacingClientMetadataInput{
		Channel: "side-channel",
	})
	if err == nil {
		t.Fatal("expected invalid channel error")
	}
}

func TestStoreClientMetadataDoesNotLeakLocalPathsOrSecrets(t *testing.T) {
	registry, err := NewExternalToolRegistry(validExternalToolRecord())
	if err != nil {
		t.Fatalf("registry create failed: %v", err)
	}

	metadata, err := BuildServerFacingClientMetadata(ServerFacingClientMetadataInput{
		Channel:    DistributionChannelMacAppStore,
		AppVersion: "1.2.3",
		Registry:   registry,
	})
	if err != nil {
		t.Fatalf("metadata build failed: %v", err)
	}

	payload, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	encoded := string(payload)
	for _, forbidden := range []string{
		"executable_path",
		"workspace_root_path",
		"token",
		"api_key",
		"/usr/local/bin/codex",
	} {
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("metadata leaked %q in %s", forbidden, encoded)
		}
	}
}
