package figmaboundary

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestFigmaAIAgentDaemonBoundaryManifest(t *testing.T) {
	root := repoRoot(t)
	manifestPath := filepath.Join(root, "docs/30-architecture/figma-ai-agent-daemon-boundary.riido.json")
	docPath := filepath.Join(root, "docs/30-architecture/figma-ai-agent-daemon-boundary.md")

	var manifest boundaryManifest
	if err := json.Unmarshal(mustReadFile(t, manifestPath), &manifest); err != nil {
		t.Fatalf("decode daemon Figma boundary manifest: %v", err)
	}

	if manifest.SchemaVersion != "riido-figma-ai-agent-daemon-boundary.v1" {
		t.Fatalf("unexpected schema_version: %q", manifest.SchemaVersion)
	}
	if manifest.ID != "figma-v1-22-ai-agent-daemon-boundary" || manifest.RiidoTask != "RIID-4813" {
		t.Fatalf("manifest identity drifted: %#v", manifest)
	}
	requireSliceContains(t, manifest.HardeningTasks, "RIID-4843")
	requireSliceContains(t, manifest.HardeningTasks, "RIID-4859")
	if manifest.HumanDoc != "docs/30-architecture/figma-ai-agent-daemon-boundary.md" {
		t.Fatalf("human doc path drifted: %q", manifest.HumanDoc)
	}
	if !strings.Contains(manifest.SourceCoverageManifest, "riido-contracts") {
		t.Fatalf("source coverage manifest must point upstream to contracts: %q", manifest.SourceCoverageManifest)
	}
	if manifest.SourceCoverageManifestProvenance.Repo != "riido-contracts" ||
		manifest.SourceCoverageManifestProvenance.SchemaVersion != "riido-figma-ai-agent-coverage.v1" ||
		manifest.SourceCoverageManifestProvenance.ID != "figma-v1-22-ai-agent-ui-coverage" {
		t.Fatalf("upstream coverage provenance drifted: %#v", manifest.SourceCoverageManifestProvenance)
	}
	if manifest.SourceCoverageManifestProvenance.MirrorsSourceField != "stabilized_by" ||
		manifest.SourceCoverageManifestProvenance.SourceFieldIntroducedBy != "teamswyg/riido-contracts#53" {
		t.Fatalf("upstream coverage provenance source field marker drifted: %#v", manifest.SourceCoverageManifestProvenance)
	}
	expectedSourceProvenance := []string{
		"teamswyg/riido-contracts#38",
		"teamswyg/riido-contracts#39",
		"teamswyg/riido-contracts#45",
		"teamswyg/riido-contracts#46",
		"teamswyg/riido-contracts#51",
		"teamswyg/riido-contracts#52",
		"teamswyg/riido-contracts#54",
	}
	requireSameStringSlice(t, manifest.SourceCoverageManifestProvenance.StabilizedBy, expectedSourceProvenance)
	if manifest.Figma.FileKey != "MUOd9lctoEHASUStN3vUuK" || manifest.Figma.PageID != "129:5215" {
		t.Fatalf("Figma source drifted: %#v", manifest.Figma)
	}
	if manifest.BoundaryPolicy.TopDown == "" || manifest.BoundaryPolicy.BottomUp == "" {
		t.Fatalf("top-down/bottom-up policy must be explicit: %#v", manifest.BoundaryPolicy)
	}

	entries := map[string]boundaryEntry{}
	for _, entry := range manifest.Entries {
		if entry.NodeID == "" || entry.Name == "" || entry.DaemonScope == "" {
			t.Fatalf("entry must include node, name, and daemon_scope: %#v", entry)
		}
		if len(entry.UpstreamOwner) == 0 {
			t.Fatalf("entry %s must name upstream owners", entry.NodeID)
		}
		if entry.DaemonConsumedFacts == nil {
			t.Fatalf("entry %s must include daemon_consumed_facts, even when empty", entry.NodeID)
		}
		if len(entry.ClientOwnedFacts) == 0 {
			t.Fatalf("entry %s must separate client_owned_facts", entry.NodeID)
		}
		entries[entry.NodeID] = entry
	}

	limitation := requireToolLimitation(t, manifest.MirroredSupportingToolLimitations, "figma-metadata-page-list-underreports-pages.v1")
	if limitation.SourceOwner != "riido-contracts" || limitation.LocalRiidoTask != "RIID-4843" {
		t.Fatalf("metadata page-list limitation mirror drifted: %#v", limitation)
	}
	requireSameStringSlice(t, limitation.SourceStabilizedBy, []string{"teamswyg/riido-contracts#52"})
	for _, pageID := range []string{"129:5215", "42:3014", "0:1"} {
		requireSliceContains(t, limitation.RequiredAuthoritativePages, pageID)
	}
	requireContains(t, limitation.DaemonScope, "must not collapse")

	for _, nodeID := range []string{
		"153:12742",
		"153:15931",
		"153:15935",
		"156:19307",
		"162:23090",
		"432:37336",
		"134:6542",
		"432:35713",
		"42:3014",
		"137:6746",
		"138:7389",
		"432:46849",
		"164:26969",
		"164:30192",
		"164:30206",
		"435:60050",
		"236:29749",
		"275:22731",
	} {
		if _, ok := entries[nodeID]; !ok {
			t.Fatalf("daemon-relevant Figma node %s missing from manifest", nodeID)
		}
	}
	for _, nodeID := range limitation.MustPreserveNonUINodes {
		if _, ok := entries[nodeID]; !ok {
			t.Fatalf("metadata limitation requires preserving non-UI node %s", nodeID)
		}
	}

	requireSliceContains(t, entries["162:23090"].UpstreamOwner, "riido-daemon")
	requireSliceContains(t, entries["137:6746"].UpstreamOwner, "riido-daemon")
	requireContains(t, entries["432:37336"].DaemonScope, "Consumes the assigned runtime")
	requireContains(t, entries["138:7389"].DaemonScope, "No fixture catalog ownership")
	requireContains(t, entries["432:46849"].DaemonScope, "No onboarding draft execution ownership")
	requireContains(t, entries["432:46849"].DaemonScope, "workspace_id")
	requireContains(t, entries["432:46849"].DaemonScope, "runtime_id")

	humanDoc := string(mustReadFile(t, docPath))
	requireContains(t, humanDoc, manifest.SchemaVersion)
	requireContains(t, humanDoc, "RIID-4843")
	requireContains(t, humanDoc, "RIID-4847")
	requireContains(t, humanDoc, "RIID-4851")
	requireContains(t, humanDoc, "figma-metadata-page-list-underreports-pages.v1")
	requireContains(t, humanDoc, "teamswyg/riido-contracts#53")
	requireContains(t, humanDoc, "teamswyg/riido-contracts#54")
	requireContains(t, humanDoc, "`stabilized_by`")
	requireContains(t, humanDoc, "teamswyg/riido-contracts#38")
	requireContains(t, humanDoc, "teamswyg/riido-contracts#52")
	requireContains(t, humanDoc, "432:37336")
	requireContains(t, humanDoc, "432:46849")
	requireContains(t, humanDoc, "workspace-less create")
	requireContains(t, humanDoc, "fixture")
	requireContains(t, humanDoc, "Bottom-up")
}

func TestFigmaAIAgentDaemonBoundaryDocsStayLinked(t *testing.T) {
	root := repoRoot(t)
	for _, rel := range []string{
		"docs/README.md",
		"docs/20-domain/context-map.md",
		"docs/20-domain/provider-runtime.md",
		"docs/30-architecture/cli-surface.md",
		"docs/migration/daemon.md",
	} {
		body := string(mustReadFile(t, filepath.Join(root, rel)))
		requireContains(t, body, "figma-ai-agent-daemon-boundary")
	}
}
