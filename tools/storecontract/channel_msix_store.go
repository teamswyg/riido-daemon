package main

func validateMSIXStoreSurfaces(item channel) []string {
	var problems []string
	problems = append(problems, requireRuntimeContract(
		item,
		"msix-packaged-full-trust-helper-tray",
		"explicit-consent-and-store-review",
		"windows-named-pipe",
		"windows-package-local-data",
		"store-managed",
	)...)
	problems = append(problems, requireRequiredSurfaces(item, msixStoreRequiredSurfaces...)...)
	problems = append(problems, requireRequiredSurfaces(item, storeReviewSubmissionRequiredSurfaces...)...)
	return append(problems, requireForbiddenSurfaces(item, "windows-service-default", "self-updater")...)
}

var msixStoreRequiredSurfaces = []string{
	"package-identity",
	"windows-desktop-target-device-family",
	"named-pipe-local-ipc",
	"package-local-data",
	"runfulltrust-review-note",
	"store-managed-updates",
	"partner-center-review-notes",
	"review-demo-mode",
	"privacy-policy",
}
