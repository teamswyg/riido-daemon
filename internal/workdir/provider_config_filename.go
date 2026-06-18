package workdir

// ProviderConfigFilename returns the native config filename for a provider.
// Unknown providers fall back to AGENTS.md through the generated
// native-config plan catalog.
func ProviderConfigFilename(provider string) string {
	return ProviderConfigPlan(provider).PrimaryInstructionFile
}
