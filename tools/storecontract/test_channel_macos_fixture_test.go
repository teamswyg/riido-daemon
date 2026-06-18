package main

func developerIDChannel() channel {
	return channel{
		ID:                "developer-id",
		Platform:          "macos",
		Status:            "preferred-first",
		RuntimeRole:       "local-helper-broker",
		BackgroundRule:    "explicit-consent",
		LocalIPCTransport: "unix-socket",
		DataRoot:          "user-application-support",
		UpdateMechanism:   "self-managed",
		RequiredSurfaces: []string{
			"developer-id-signing",
			"notarization",
			"user-consented-background-helper",
			"local-only-ipc",
		},
		ForbiddenSurfaces: []string{
			"bundled-provider-cli",
			"silent-provider-install",
			"external-tcp-listener",
			"arbitrary-home-scan",
		},
	}
}

func macAppStoreChannel() channel {
	return channel{
		ID:                "mac-app-store",
		Platform:          "macos",
		Status:            "requires-redesign",
		RuntimeRole:       "sandboxed-login-item-helper",
		BackgroundRule:    "explicit-consent-and-store-review",
		LocalIPCTransport: "unix-socket",
		DataRoot:          "app-group-or-sandbox-container",
		UpdateMechanism:   "app-store-managed",
		RequiredSurfaces:  macAppStoreRequiredSurfaceFixture(),
		ForbiddenSurfaces: macAppStoreForbiddenSurfaceFixture(),
	}
}
