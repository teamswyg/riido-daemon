package validation

func providerRunID(provider, commandID string) string {
	return "provider-run:" + sanitizeID(provider) + ":" + sanitizeID(commandID)
}
