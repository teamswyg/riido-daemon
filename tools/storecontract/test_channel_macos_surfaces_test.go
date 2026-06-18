package main

func macAppStoreRequiredSurfaceFixture() []string {
	return []string{
		"app-sandbox",
		"app-group-or-container-ipc",
		"security-scoped-workspace-grant",
		"service-management-login-item-consent",
		"helper-purpose-review-note",
		"app-sandbox-entitlement-review-notes",
		"app-store-managed-updates",
		"privacy-policy",
		"review-demo-mode",
		"demo-review-account",
		"privacy-metadata-allowlist",
		"provider-non-bundling-review-note",
	}
}

func macAppStoreForbiddenSurfaceFixture() []string {
	return []string{
		"bundled-provider-cli",
		"silent-provider-install",
		"direct-launchagent-install",
		"self-updater",
		"external-tcp-listener",
		"arbitrary-home-scan",
		"third-party-installer",
		"shared-location-code-install",
		"standalone-code-download",
		"root-privilege-escalation",
	}
}
