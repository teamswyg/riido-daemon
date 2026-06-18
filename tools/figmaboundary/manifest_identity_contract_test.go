package figmaboundary

import (
	"strings"
	"testing"
)

func TestFigmaAIAgentDaemonBoundaryManifestIdentity(t *testing.T) {
	manifest := loadBoundaryManifest(t)
	if manifest.SchemaVersion != "riido-figma-ai-agent-daemon-boundary.v1" {
		t.Fatalf("unexpected schema_version: %q", manifest.SchemaVersion)
	}
	if manifest.ID != "figma-v1-22-ai-agent-daemon-boundary" || manifest.RiidoTask != "RIID-4813" {
		t.Fatalf("manifest identity drifted: %#v", manifest)
	}
	requireSliceContains(t, manifest.HardeningTasks, "RIID-4843")
	requireSliceContains(t, manifest.HardeningTasks, "RIID-4859")
	if manifest.HumanDoc != boundaryHumanDocRelPath {
		t.Fatalf("human doc path drifted: %q", manifest.HumanDoc)
	}
	if !strings.Contains(manifest.SourceCoverageManifest, "riido-contracts") {
		t.Fatalf("source coverage manifest must point upstream to contracts: %q", manifest.SourceCoverageManifest)
	}
	if manifest.Figma.FileKey != "MUOd9lctoEHASUStN3vUuK" || manifest.Figma.PageID != "129:5215" {
		t.Fatalf("Figma source drifted: %#v", manifest.Figma)
	}
	if manifest.BoundaryPolicy.TopDown == "" || manifest.BoundaryPolicy.BottomUp == "" {
		t.Fatalf("top-down/bottom-up policy must be explicit: %#v", manifest.BoundaryPolicy)
	}
}
