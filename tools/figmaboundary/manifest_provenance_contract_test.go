package figmaboundary

import "testing"

func TestFigmaBoundaryManifestUpstreamProvenance(t *testing.T) {
	provenance := loadBoundaryManifest(t).SourceCoverageManifestProvenance
	if provenance.Repo != "riido-contracts" ||
		provenance.SchemaVersion != "riido-figma-ai-agent-coverage.v1" ||
		provenance.ID != "figma-v1-22-ai-agent-ui-coverage" {
		t.Fatalf("upstream coverage provenance drifted: %#v", provenance)
	}
	if provenance.MirrorsSourceField != "stabilized_by" ||
		provenance.SourceFieldIntroducedBy != "teamswyg/riido-contracts#53" {
		t.Fatalf("upstream coverage provenance source field marker drifted: %#v", provenance)
	}
	requireSameStringSlice(t, provenance.StabilizedBy, expectedSourceProvenance())
}

func expectedSourceProvenance() []string {
	return []string{
		"teamswyg/riido-contracts#38",
		"teamswyg/riido-contracts#39",
		"teamswyg/riido-contracts#45",
		"teamswyg/riido-contracts#46",
		"teamswyg/riido-contracts#51",
		"teamswyg/riido-contracts#52",
		"teamswyg/riido-contracts#54",
	}
}
