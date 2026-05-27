package hostintegration

import (
	"reflect"
	"slices"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/jsontest"
)

func TestPrivacyMetadataAllowlistLoadsRequiredSurfaces(t *testing.T) {
	allowlist, err := LoadPrivacyMetadataAllowlist()
	if err != nil {
		t.Fatalf("LoadPrivacyMetadataAllowlist: %v", err)
	}
	if allowlist.SchemaVersion != PrivacyMetadataAllowlistSchemaVersion {
		t.Fatalf("unexpected schema: %s", allowlist.SchemaVersion)
	}
	for _, id := range []string{PrivacySurfaceServerFacingClientMetadata, PrivacySurfaceProviderStatusSyncRequest} {
		if _, ok := allowlist.Surface(id); !ok {
			t.Fatalf("missing surface %s", id)
		}
	}
}

func TestPrivacyMetadataAllowlistMatchesServerFacingMetadataShape(t *testing.T) {
	allowlist, err := LoadPrivacyMetadataAllowlist()
	if err != nil {
		t.Fatalf("LoadPrivacyMetadataAllowlist: %v", err)
	}
	surface, ok := allowlist.Surface(PrivacySurfaceServerFacingClientMetadata)
	if !ok {
		t.Fatalf("missing surface %s", PrivacySurfaceServerFacingClientMetadata)
	}

	got := jsontest.StructJSONPaths(reflect.TypeOf(ServerFacingClientMetadata{}))
	if !reflect.DeepEqual(got, surface.AllowedJSONPaths) {
		t.Fatalf("server-facing metadata paths = %#v, want %#v", got, surface.AllowedJSONPaths)
	}
	for _, forbidden := range []string{"provider_executable_path", "workspace_root_path", "token", "api_key", "raw_environment"} {
		if !surface.Forbids(forbidden) {
			t.Fatalf("surface should forbid %s", forbidden)
		}
		if slices.Contains(got, forbidden) {
			t.Fatalf("struct path leaked forbidden field %s", forbidden)
		}
	}
}
