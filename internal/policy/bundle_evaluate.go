package policy

func EvaluateUnsafeBypassWithBundle(bundle PolicyBundle, input UnsafeBypassInput) Decision {
	input.BundleAllows = bundle.AllowsUnsafeBypass(input.TrustTier, input.Surface)
	input.PolicyVersion = bundle.Version
	return EvaluateUnsafeBypass(input)
}

func EvaluateNativeConfigHookWithBundle(bundle PolicyBundle, input NativeConfigHookInput) Decision {
	input.BundleAllows = bundle.AllowsNativeConfigHook(input.TrustTier, input.Surface)
	input.PolicyVersion = bundle.Version
	return EvaluateNativeConfigHook(input)
}

func EvaluateNativeConfigFileWithBundle(bundle PolicyBundle, input NativeConfigFileInput) Decision {
	input.BundleAllows = bundle.AllowsNativeConfigFile(input.TrustTier, input.Surface)
	input.PolicyVersion = bundle.Version
	return EvaluateNativeConfigFile(input)
}

func EvaluateToolUseWithBundle(bundle PolicyBundle, input ToolUseInput) ToolUseDecision {
	input.BundleAllows = bundle.AllowsToolUse(input.TrustTier, input.Surface)
	input.PolicyVersion = bundle.Version
	return EvaluateToolUse(input)
}
