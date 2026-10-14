package hostintegration

func applyMacAppStoreHelperRuntimePlan(plan *HelperRuntimePlan, in HelperRuntimePlanInput) {
	plan.Role = HelperRuntimeRoleSandboxedLoginItem
	plan.BackgroundRule = HelperBackgroundConsentAndStoreReview
	plan.StartupRegistration = HelperStartupServiceManagementLoginItem
	plan.BackgroundAllowed = in.Consent.BackgroundHelper && in.StoreReviewApproved
	plan.RequiresStoreReviewApproval = true
	plan.RequiresWorkspaceGrant = WorkspaceGrantSecurityScopedBookmark
	plan.SelfUpdaterAllowed = false
	plan.StoreManagedUpdates = true
	plan.ReviewNoteSurfaces = macAppStoreHelperReviewSurfaces()
}

func macAppStoreHelperReviewSurfaces() []string {
	return []string{
		"app-sandbox-entitlement-review-notes",
		"service-management-login-item-consent",
		"security-scoped-workspace-grant",
		"helper-purpose-review-note",
		"provider-non-bundling-review-note",
		"review-demo-mode",
		"privacy-metadata-allowlist",
	}
}
