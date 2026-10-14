package main

func validContract() contract {
	return contract{
		SchemaVersion:            contractSchemaVersion,
		Product:                  "riido_daemon",
		ProviderCLIBundling:      "forbidden",
		ExternalProviderCLINames: []string{"claude", "codex", "openclaw", "cursor-agent"},
		StoreArtifactRoots:       []string{"packaging/store"},
		RequiredDocs:             requiredDocPaths(),
		RequiredNoticeTerms:      requiredNoticeTerms(),
		Channels: []channel{
			developerIDChannel(),
			macAppStoreChannel(),
			msixSideloadChannel(),
			msixStoreChannel(),
		},
	}
}

func requiredDocPaths() []string {
	return []string{
		"docs/20-domain/distribution-host-integration.md",
		"docs/30-architecture/store-distribution.md",
		"NOTICE.md",
	}
}

func requiredNoticeTerms() []string {
	return []string{
		"No source code from any third-party project is directly incorporated",
		"Modified Apache License, Version 2.0",
		"do not redistribute any vendor code or bundle provider CLI executables",
		"No vendored third-party code",
	}
}
