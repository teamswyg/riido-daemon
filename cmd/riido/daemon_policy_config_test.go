package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestLoadDaemonSettingsLoadsPolicyBundleFile(t *testing.T) {
	path := writePolicyBundleFile(t, "policy-bundle.file.v1")
	env := map[string]string{envPolicyBundlePath: path}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.PolicyBundle != "policy-bundle.file.v1" || settings.PolicyBundlePath != path {
		t.Fatalf("policy bundle settings = %+v", settings)
	}
	if settings.PolicyBundleDoc.Version != "policy-bundle.file.v1" {
		t.Fatalf("policy bundle doc = %+v", settings.PolicyBundleDoc)
	}
}

func TestLoadDaemonSettingsRejectsPolicyBundleVersionMismatch(t *testing.T) {
	path := writePolicyBundleFile(t, "policy-bundle.file.v1")
	env := map[string]string{
		envPolicyBundlePath: path,
		envPolicyBundle:     "policy-bundle.env.v1",
	}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected policy bundle version mismatch error")
	}
}

func TestLoadDaemonSettingsRejectsInvalidPolicyBundleFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy-bundle.riido.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":"wrong"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{envPolicyBundlePath: path}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected invalid policy bundle file error")
	}
}

func TestDaemonToolAutoApproverUsesActivePolicyBundle(t *testing.T) {
	settings := daemonSettings{PolicyBundleDoc: policy.PolicyBundle{
		SchemaVersion:  policy.BundleSchemaVersion,
		Version:        "policy-bundle.tool-auto.v1",
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[policy.TrustTier]policy.TrustTierPolicy{
			policy.TrustTierHost: {
				AllowedSurfaces: policy.AllowedSurfaceSet{
					ToolUse: []policy.ToolUseSurface{policy.ToolUseDestructiveCommand},
				},
			},
		},
	}}
	approver := daemonToolAutoApprover(settings)

	if !approver(agentbridge.ToolRef{Kind: "shell"}) {
		t.Fatal("daemon policy auto approver should approve explicitly allowed shell surface")
	}
	if approver(agentbridge.ToolRef{Kind: "patch_apply"}) {
		t.Fatal("daemon policy auto approver must not approve unallowed patch surface")
	}
}

func TestDaemonToolStartGateUsesActivePolicyBundle(t *testing.T) {
	settings := daemonSettings{PolicyBundleDoc: policy.PolicyBundle{
		SchemaVersion:  policy.BundleSchemaVersion,
		Version:        "policy-bundle.tool-start.v1",
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[policy.TrustTier]policy.TrustTierPolicy{
			policy.TrustTierHost: {
				AllowedSurfaces: policy.AllowedSurfaceSet{
					ToolUse: []policy.ToolUseSurface{policy.ToolUseNetworkEgress},
				},
			},
		},
	}}
	gate := daemonToolStartGate(settings)

	if decision := gate(agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "curl https://example.com"}}); decision.Block {
		t.Fatalf("allowed network surface should not block: %+v", decision)
	}
	if decision := gate(agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "terraform destroy"}}); !decision.Block {
		t.Fatalf("unallowed destructive command should block: %+v", decision)
	}
}

func TestDaemonToolApprovalGateUsesActivePolicyBundle(t *testing.T) {
	settings := daemonSettings{PolicyBundleDoc: policy.PolicyBundle{
		SchemaVersion:  policy.BundleSchemaVersion,
		Version:        "policy-bundle.tool-approval.v1",
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[policy.TrustTier]policy.TrustTierPolicy{
			policy.TrustTierHost: {
				AllowedSurfaces: policy.AllowedSurfaceSet{
					ToolUse: []policy.ToolUseSurface{policy.ToolUseNetworkEgress},
				},
			},
		},
	}}
	gate := daemonToolApprovalGate(settings)

	if decision := gate(agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "curl https://example.com"}}); decision.Block {
		t.Fatalf("allowed network approval should not block: %+v", decision)
	}
	if decision := gate(agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "cat .env.local"}}); !decision.Block {
		t.Fatalf("unallowed secret exposure approval should block: %+v", decision)
	} else if decision.Code != "approval_timeout" {
		t.Fatalf("unallowed secret exposure approval code = %q", decision.Code)
	}
}

func writePolicyBundleFile(t *testing.T, version string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "policy-bundle.riido.json")
	body := `{
		"schema_version": "riido-policy-bundle.v1",
		"version": "` + version + `",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {}
	}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadDaemonSettingsRejectsInvalidWorkspaceCount(t *testing.T) {
	env := map[string]string{envWorkspaceCount: "nope"}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected invalid workspace count error")
	}
}

func TestLoadDaemonSettingsDefaultsWorkdirCleanupInterval(t *testing.T) {
	env := map[string]string{envWorkdirRetentionSeconds: "7200"}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.WorkdirRetention != 2*time.Hour || settings.WorkdirCleanupEvery != time.Hour {
		t.Fatalf("workdir cleanup settings = %+v", settings)
	}
}

func TestLoadDaemonSettingsRejectsCleanupIntervalWithoutRetention(t *testing.T) {
	env := map[string]string{envWorkdirCleanupIntervalSeconds: "60"}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected cleanup interval without retention error")
	}
}

func TestBuildDaemonControlPlaneUsesMemoryByDefault(t *testing.T) {
	source, reporter, kind, err := buildDaemonControlPlane(daemonSettings{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if kind != "memory" {
		t.Fatalf("kind = %q", kind)
	}
	if _, ok := source.(*controlplane.MemorySource); !ok {
		t.Fatalf("source type = %T", source)
	}
	if _, ok := reporter.(*controlplane.MemoryReporter); !ok {
		t.Fatalf("reporter type = %T", reporter)
	}
}
