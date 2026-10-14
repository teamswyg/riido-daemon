package hostintegration

func applyMSIXSideloadHelperRuntimePlan(plan *HelperRuntimePlan, in HelperRuntimePlanInput) {
	plan.Role = HelperRuntimeRoleMSIXPackagedBroker
	plan.StartupRegistration = HelperStartupMSIXPackagedStartupTask
	plan.BackgroundAllowed = in.Consent.BackgroundHelper
}

func applyMSIXStoreHelperRuntimePlan(plan *HelperRuntimePlan, in HelperRuntimePlanInput) {
	plan.Role = HelperRuntimeRoleMSIXFullTrustTray
	plan.BackgroundRule = HelperBackgroundConsentAndStoreReview
	plan.StartupRegistration = HelperStartupMSIXPackagedStartupTask
	plan.BackgroundAllowed = in.Consent.BackgroundHelper && in.StoreReviewApproved
	plan.RequiresStoreReviewApproval = true
	plan.SelfUpdaterAllowed = false
	plan.StoreManagedUpdates = true
	plan.ReviewNoteSurfaces = msixStoreHelperReviewSurfaces()
}

func msixStoreHelperReviewSurfaces() []string {
	return []string{
		"runfulltrust-review-note",
		"partner-center-review-notes",
		"provider-non-bundling-review-note",
		"review-demo-mode",
		"privacy-metadata-allowlist",
	}
}
