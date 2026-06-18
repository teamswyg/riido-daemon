package figmaboundary

import "testing"

func TestFigmaBoundaryManifestKeepsDaemonRelevantEntries(t *testing.T) {
	entries := boundaryEntriesByNodeID(t, loadBoundaryManifest(t))
	for _, nodeID := range daemonRelevantFigmaNodes() {
		if _, ok := entries[nodeID]; !ok {
			t.Fatalf("daemon-relevant Figma node %s missing from manifest", nodeID)
		}
	}
	requireSliceContains(t, entries["162:23090"].UpstreamOwner, "riido-daemon")
	requireSliceContains(t, entries["137:6746"].UpstreamOwner, "riido-daemon")
	requireContains(t, entries["432:37336"].DaemonScope, "Consumes the assigned runtime")
	requireContains(t, entries["138:7389"].DaemonScope, "No fixture catalog ownership")
	requireContains(t, entries["432:46849"].DaemonScope, "No onboarding draft execution ownership")
	requireContains(t, entries["432:46849"].DaemonScope, "workspace_id")
	requireContains(t, entries["432:46849"].DaemonScope, "runtime_id")
}

func daemonRelevantFigmaNodes() []string {
	return []string{
		"153:12742", "153:15931", "153:15935", "156:19307",
		"162:23090", "432:37336", "134:6542", "432:35713",
		"42:3014", "137:6746", "138:7389", "432:46849",
		"164:26969", "164:30192", "164:30206", "435:60050",
		"236:29749", "275:22731",
	}
}
