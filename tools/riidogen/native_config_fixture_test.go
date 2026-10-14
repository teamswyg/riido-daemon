package main

func validNativeConfigPlanSpec() nativeConfigPlanCatalogSpec {
	return nativeConfigPlanCatalogSpec{
		SchemaVersion:         "riido-native-config-plan.v1",
		ManifestSchemaVersion: "riido-native-config-manifest.v1",
		Default: nativeConfigProviderPlanSpec{
			PrimaryInstructionFile: "AGENTS.md",
			ManifestFile:           "manifest.json",
			HookMode:               "none",
		},
		Providers: []nativeConfigProviderPlanSpec{
			{
				ProviderKind:           "claude",
				PrimaryInstructionFile: "CLAUDE.md",
				ManifestFile:           "manifest.json",
				HookMode:               "audit",
			},
		},
	}
}
