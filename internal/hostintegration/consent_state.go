package hostintegration

import (
	"slices"
	"sort"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

// ConsentState is the current view reduced from append-only records.
type ConsentState struct {
	BackgroundHelper bool
	TelemetrySync    bool
	ReviewDemoMode   bool
	ProviderExecute  map[capability.ProviderKind]bool
	WorkspaceAccess  map[string]bool
}

// ProviderExecutionAllowed reports whether a provider can be executed under
// the latest consent view.
func (s ConsentState) ProviderExecutionAllowed(provider capability.ProviderKind) bool {
	return s.ProviderExecute[provider]
}

// WorkspaceAccessAllowed reports whether a workspace root grant is active.
func (s ConsentState) WorkspaceAccessAllowed(workspaceID string) bool {
	return s.WorkspaceAccess[workspaceID]
}

// GrantedProviders returns active provider grants in deterministic order.
func (s ConsentState) GrantedProviders() []capability.ProviderKind {
	providers := make([]capability.ProviderKind, 0, len(s.ProviderExecute))
	for provider, granted := range s.ProviderExecute {
		if granted {
			providers = append(providers, provider)
		}
	}
	slices.Sort(providers)
	return providers
}

// GrantedWorkspaces returns active workspace grants in deterministic order.
func (s ConsentState) GrantedWorkspaces() []string {
	workspaces := make([]string, 0, len(s.WorkspaceAccess))
	for workspaceID, granted := range s.WorkspaceAccess {
		if granted {
			workspaces = append(workspaces, workspaceID)
		}
	}
	sort.Strings(workspaces)
	return workspaces
}
