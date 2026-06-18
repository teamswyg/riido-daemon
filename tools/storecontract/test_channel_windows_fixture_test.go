package main

func msixSideloadChannel() channel {
	return channel{
		ID:                "msix-sideload",
		Platform:          "windows",
		Status:            "preferred-first",
		RuntimeRole:       "msix-packaged-helper-broker",
		BackgroundRule:    "explicit-consent",
		LocalIPCTransport: "windows-named-pipe",
		DataRoot:          "windows-package-local-data",
		UpdateMechanism:   "self-managed",
		RequiredSurfaces: []string{
			"signed-msix-package",
			"package-identity",
			"windows-desktop-target-device-family",
			"named-pipe-local-ipc",
			"package-local-data",
			"user-consented-background-helper",
		},
		ForbiddenSurfaces: []string{
			"bundled-provider-cli",
			"silent-provider-install",
			"windows-service-default",
			"external-tcp-listener",
			"arbitrary-home-scan",
		},
	}
}

func msixStoreChannel() channel {
	return channel{
		ID:                "msix-store",
		Platform:          "windows",
		Status:            "requires-policy-gate",
		RuntimeRole:       "msix-packaged-full-trust-helper-tray",
		BackgroundRule:    "explicit-consent-and-store-review",
		LocalIPCTransport: "windows-named-pipe",
		DataRoot:          "windows-package-local-data",
		UpdateMechanism:   "store-managed",
		RequiredSurfaces:  msixStoreRequiredSurfaceFixture(),
		ForbiddenSurfaces: msixStoreForbiddenSurfaceFixture(),
	}
}
