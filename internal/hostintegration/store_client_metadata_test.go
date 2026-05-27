package hostintegration

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

func TestStoreClientMetadataCollectsOnlyRoutingFields(t *testing.T) {
	registry, err := NewExternalToolRegistry(
		externalToolRecordForProvider("codex", ToolLoginLoggedIn, capability.CompatSupported),
		externalToolRecordForProvider("claude", ToolLoginRequired, capability.CompatSupported),
	)
	if err != nil {
		t.Fatalf("registry create failed: %v", err)
	}

	metadata, err := BuildServerFacingClientMetadata(ServerFacingClientMetadataInput{
		Channel:    DistributionChannelMSIXStore,
		AppVersion: " 1.2.3 ",
		Registry:   registry,
	})
	if err != nil {
		t.Fatalf("metadata build failed: %v", err)
	}

	if metadata.DistributionChannel != DistributionChannelMSIXStore || metadata.AppVersion != "1.2.3" {
		t.Fatalf("unexpected envelope: %+v", metadata)
	}
	if got, want := metadata.ServerFacingProviderKinds(), []capability.ProviderKind{"claude", "codex"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("provider order = %v, want %v", got, want)
	}
	if metadata.Providers[0].ProviderAvailable {
		t.Fatalf("login-required provider should not be routable: %+v", metadata.Providers[0])
	}
	if metadata.Providers[0].RoutingStatus != ProviderRoutingLoginRequired {
		t.Fatalf("login-required routing status = %+v", metadata.Providers[0])
	}
	if !metadata.Providers[1].ProviderAvailable {
		t.Fatalf("logged-in supported provider should be routable: %+v", metadata.Providers[1])
	}
	if metadata.Providers[1].RoutingStatus != ProviderRoutingAvailable {
		t.Fatalf("available routing status = %+v", metadata.Providers[1])
	}
}

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

func externalToolRecordForProvider(provider capability.ProviderKind, login ToolLoginStatus, compat capability.CompatibilityStatus) ExternalToolRecord {
	record := validExternalToolRecord()
	record.Provider = provider
	record.ExecutablePath = "/usr/local/bin/" + string(provider)
	record.LoginStatus = login
	record.CompatibilityStatus = compat
	return record
}
