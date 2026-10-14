package main

func msixStoreRequiredSurfaceFixture() []string {
	return []string{
		"package-identity",
		"windows-desktop-target-device-family",
		"named-pipe-local-ipc",
		"package-local-data",
		"runfulltrust-review-note",
		"store-managed-updates",
		"partner-center-review-notes",
		"review-demo-mode",
		"privacy-policy",
		"demo-review-account",
		"modified-apache-notice-legal-review-before-public-store",
		"privacy-metadata-allowlist",
		"provider-non-bundling-review-note",
	}
}

func msixStoreForbiddenSurfaceFixture() []string {
	return []string{
		"bundled-provider-cli",
		"silent-provider-install",
		"windows-service-default",
		"self-updater",
		"external-tcp-listener",
		"arbitrary-home-scan",
	}
}
