package main

func validateMSIXSideloadSurfaces(item channel) []string {
	var problems []string
	problems = append(problems, requireRuntimeContract(
		item,
		"msix-packaged-helper-broker",
		"explicit-consent",
		"windows-named-pipe",
		"windows-package-local-data",
		"self-managed",
	)...)
	problems = append(problems, requireRequiredSurfaces(item, msixSideloadRequiredSurfaces...)...)
	return append(problems, requireForbiddenSurfaces(item, "windows-service-default")...)
}

var msixSideloadRequiredSurfaces = []string{
	"signed-msix-package",
	"package-identity",
	"windows-desktop-target-device-family",
	"named-pipe-local-ipc",
	"package-local-data",
	"user-consented-background-helper",
}
