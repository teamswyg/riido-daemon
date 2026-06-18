package hostintegration

import (
	"reflect"
	"testing"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

func TestConsentStateGrantedSubjectsAreDeterministic(t *testing.T) {
	state := ConsentState{
		ProviderExecute: map[capability.ProviderKind]bool{
			"codex":  true,
			"claude": true,
			"cursor": false,
		},
		WorkspaceAccess: map[string]bool{
			"workspace-z": true,
			"workspace-a": true,
			"workspace-b": false,
		},
	}

	wantProviders := []capability.ProviderKind{"claude", "codex"}
	if got := state.GrantedProviders(); !reflect.DeepEqual(got, wantProviders) {
		t.Fatalf("providers = %v, want %v", got, wantProviders)
	}
	wantWorkspaces := []string{"workspace-a", "workspace-z"}
	if got := state.GrantedWorkspaces(); !reflect.DeepEqual(got, wantWorkspaces) {
		t.Fatalf("workspaces = %v, want %v", got, wantWorkspaces)
	}
}
