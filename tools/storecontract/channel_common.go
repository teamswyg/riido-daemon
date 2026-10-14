package main

func validateCommonForbiddenSurfaces(item channel) []string {
	return requireForbiddenSurfaces(
		item,
		"bundled-provider-cli",
		"silent-provider-install",
		"external-tcp-listener",
		"arbitrary-home-scan",
	)
}
