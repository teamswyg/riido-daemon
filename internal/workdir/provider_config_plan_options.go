package workdir

type ProviderConfigPlanOptions struct {
	NativeHookMode       string
	NativeConfigHomeMode string
}

func ResolveProviderConfigPlan(provider, nativeHookMode string) (ProviderNativeConfigPlan, error) {
	return ResolveProviderConfigPlanWithOptions(provider, ProviderConfigPlanOptions{
		NativeHookMode: nativeHookMode,
	})
}
