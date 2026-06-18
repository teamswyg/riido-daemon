package hostintegration

func applyHelperRuntimeChannelPlan(plan *HelperRuntimePlan, in HelperRuntimePlanInput) bool {
	switch in.Channel {
	case DistributionChannelDevLocal, DistributionChannelDeveloperID:
		applyLooseHelperRuntimePlan(plan, in)
	case DistributionChannelMacAppStore:
		applyMacAppStoreHelperRuntimePlan(plan, in)
	case DistributionChannelMSIXSideload:
		applyMSIXSideloadHelperRuntimePlan(plan, in)
	case DistributionChannelMSIXStore:
		applyMSIXStoreHelperRuntimePlan(plan, in)
	default:
		return false
	}
	return true
}

func applyLooseHelperRuntimePlan(plan *HelperRuntimePlan, in HelperRuntimePlanInput) {
	plan.Role = HelperRuntimeRoleLocalBroker
	plan.StartupRegistration = HelperStartupLaunchAgentOrLoginItem
	plan.BackgroundAllowed = in.Consent.BackgroundHelper
	plan.DirectLaunchAgentAllowed = true
}
