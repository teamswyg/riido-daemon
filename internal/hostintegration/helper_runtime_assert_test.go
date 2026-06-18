package hostintegration

import (
	"slices"
	"testing"
)

func assertMacAppStorePlanShape(t *testing.T, plan HelperRuntimePlan) {
	t.Helper()
	if plan.Role != HelperRuntimeRoleSandboxedLoginItem {
		t.Fatalf("role = %q, want %q", plan.Role, HelperRuntimeRoleSandboxedLoginItem)
	}
	if plan.StartupRegistration != HelperStartupServiceManagementLoginItem {
		t.Fatalf("startup registration = %q", plan.StartupRegistration)
	}
	if plan.Endpoint.EndpointKind != LocalIPCEndpointUnixSocket {
		t.Fatalf("endpoint kind = %q", plan.Endpoint.EndpointKind)
	}
	if plan.Endpoint.Owner != LocalIPCOwnerHelper {
		t.Fatalf("endpoint owner = %q", plan.Endpoint.Owner)
	}
	if plan.Endpoint.Path != "/Users/tester/Library/Group Containers/group.io.riido/agentd.sock" {
		t.Fatalf("endpoint path = %q", plan.Endpoint.Path)
	}
	if plan.RequiresWorkspaceGrant != WorkspaceGrantSecurityScopedBookmark {
		t.Fatalf("workspace grant = %q", plan.RequiresWorkspaceGrant)
	}
}

func assertReviewSurfaces(t *testing.T, plan HelperRuntimePlan, surfaces ...string) {
	t.Helper()
	for _, surface := range surfaces {
		if !slices.Contains(plan.ReviewNoteSurfaces, surface) {
			t.Fatalf("review surfaces missing %s: %v", surface, plan.ReviewNoteSurfaces)
		}
	}
}
