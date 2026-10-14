package hostintegration

func baseHelperRuntimePlan(in HelperRuntimePlanInput, endpoint LocalIPCEndpoint) HelperRuntimePlan {
	return HelperRuntimePlan{
		Channel:                       in.Channel,
		HostOS:                        in.HostOS,
		Endpoint:                      endpoint,
		AppDataRoot:                   in.AppDataRoot,
		BackgroundRule:                HelperBackgroundExplicitConsent,
		ProviderCLIBundlingAllowed:    false,
		WindowsServiceAllowed:         false,
		SharedLocationInstallAllowed:  false,
		StandaloneCodeDownloadAllowed: false,
		SelfUpdaterAllowed:            true,
	}
}
