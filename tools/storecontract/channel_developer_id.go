package main

func validateDeveloperIDSurfaces(item channel) []string {
	if item.ID != "developer-id" {
		return nil
	}
	var problems []string
	problems = append(problems, requireRuntimeContract(
		item,
		"local-helper-broker",
		"explicit-consent",
		"unix-socket",
		"user-application-support",
		"self-managed",
	)...)
	return append(problems, requireRequiredSurfaces(
		item,
		"developer-id-signing",
		"notarization",
		"user-consented-background-helper",
		"local-only-ipc",
	)...)
}
