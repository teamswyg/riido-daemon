package main

import "github.com/teamswyg/riido-daemon/internal/policy"

var policyTableSurfaces = []policySurfaceSpec{
	{ID: string(policy.StoreSurfaceProviderCLIBundling), Label: "Provider CLI bundling"},
	{ID: string(policy.StoreSurfaceProviderCLIUserSelectedPath), Label: "Provider CLI user-selected path"},
	{ID: string(policy.StoreSurfaceSilentProviderAutoInstall), Label: "Silent provider auto-install"},
	{ID: string(policy.StoreSurfaceBackgroundHelper), Label: "Background helper"},
	{ID: string(policy.StoreSurfaceDirectLaunchAgentInstall), Label: "Direct LaunchAgent install"},
	{ID: string(policy.StoreSurfaceWindowsServiceInstall), Label: "Windows service install"},
	{ID: string(policy.StoreSurfaceExternalTCPListener), Label: "External TCP listener"},
	{ID: string(policy.StoreSurfaceLocalIPC), Label: "Local IPC"},
	{ID: string(policy.StoreSurfaceSelfUpdater), Label: "Self-updater"},
	{ID: string(policy.StoreSurfaceArbitraryHomeScan), Label: "Arbitrary home scan"},
}

var policyFactScenarios = []policyFactScenario{
	{},
	{Facts: []string{"explicit-consent"}, Consent: true},
	{Facts: []string{"os-grant"}, OSGrant: true},
	{Facts: []string{"store-review"}, StoreReview: true},
	{Facts: []string{"os-grant", "store-review"}, OSGrant: true, StoreReview: true},
	{Facts: []string{"explicit-consent", "store-review"}, Consent: true, StoreReview: true},
	{Facts: []string{"explicit-consent", "os-grant"}, Consent: true, OSGrant: true},
	{Facts: []string{"explicit-consent", "os-grant", "store-review"}, Consent: true, OSGrant: true, StoreReview: true},
}
