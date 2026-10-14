package toolpolicy

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestClassifyToolUseSurfaceMapsProviderNeutralLabels(t *testing.T) {
	for _, tc := range toolClassificationCases() {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ClassifyToolUseSurface(tc.tool)
			if !ok {
				t.Fatalf("tool should classify: %+v", tc.tool)
			}
			if got != tc.want {
				t.Fatalf("surface = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestClassifyToolUseSurfaceUsesArgsToAvoidBroadShellClassification(t *testing.T) {
	tool := agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "go test ./..."}}
	if got, ok := ClassifyToolUseSurface(tool); ok {
		t.Fatalf("safe shell command must stay unclassified: %q", got)
	}
}

func TestClassifyToolUseSurfaceLeavesReadOnlyToolsUnclassified(t *testing.T) {
	if got, ok := ClassifyToolUseSurface(agentbridge.ToolRef{Kind: "read", Name: "Read"}); ok {
		t.Fatalf("read-only tool must not auto-classify as a risk surface: %q", got)
	}
}

type classificationCase struct {
	name string
	tool agentbridge.ToolRef
	want policy.ToolUseSurface
}
