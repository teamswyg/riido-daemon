package main

func validateMacAppStoreSurfaces(item channel) []string {
	if item.ID != "mac-app-store" {
		return nil
	}
	var problems []string
	problems = append(problems, requireRuntimeContract(
		item,
		"sandboxed-login-item-helper",
		"explicit-consent-and-store-review",
		"unix-socket",
		"app-group-or-sandbox-container",
		"app-store-managed",
	)...)
	problems = append(problems, requireRequiredSurfaces(item, macAppStoreRequiredSurfaces...)...)
	problems = append(problems, requireRequiredSurfaces(item, storeReviewSubmissionRequiredSurfaces...)...)
	return append(problems, requireForbiddenSurfaces(item, macAppStoreForbiddenSurfaces...)...)
}

var macAppStoreRequiredSurfaces = []string{
	"app-sandbox",
	"app-group-or-container-ipc",
	"security-scoped-workspace-grant",
	"service-management-login-item-consent",
	"helper-purpose-review-note",
	"app-sandbox-entitlement-review-notes",
	"app-store-managed-updates",
	"privacy-policy",
	"review-demo-mode",
}

var macAppStoreForbiddenSurfaces = []string{
	"direct-launchagent-install",
	"self-updater",
	"third-party-installer",
	"shared-location-code-install",
	"standalone-code-download",
	"root-privilege-escalation",
}
