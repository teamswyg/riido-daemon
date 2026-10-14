package figmaboundary

import "testing"

func TestFigmaBoundaryManifestMirrorsToolLimitation(t *testing.T) {
	manifest := loadBoundaryManifest(t)
	limitation := requireToolLimitation(
		t,
		manifest.MirroredSupportingToolLimitations,
		"figma-metadata-page-list-underreports-pages.v1",
	)
	if limitation.SourceOwner != "riido-contracts" || limitation.LocalRiidoTask != "RIID-4843" {
		t.Fatalf("metadata page-list limitation mirror drifted: %#v", limitation)
	}
	requireSameStringSlice(t, limitation.SourceStabilizedBy, []string{"teamswyg/riido-contracts#52"})
	for _, pageID := range []string{"129:5215", "42:3014", "0:1"} {
		requireSliceContains(t, limitation.RequiredAuthoritativePages, pageID)
	}
	requireContains(t, limitation.DaemonScope, "must not collapse")
	requirePreservedNonUINodes(t, manifest, limitation)
}

func requirePreservedNonUINodes(
	t *testing.T,
	manifest boundaryManifest,
	limitation toolLimitation,
) {
	t.Helper()
	entries := boundaryEntriesByNodeID(t, manifest)
	for _, nodeID := range limitation.MustPreserveNonUINodes {
		if _, ok := entries[nodeID]; !ok {
			t.Fatalf("metadata limitation requires preserving non-UI node %s", nodeID)
		}
	}
}
