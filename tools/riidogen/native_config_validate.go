package main

import (
	"errors"
	"fmt"
)

func validateNativeConfigPlan(spec nativeConfigPlanCatalogSpec) error {
	if spec.SchemaVersion == "" {
		return errors.New("riidogen: native config plan schema_version is required")
	}
	if spec.ManifestSchemaVersion == "" {
		return errors.New("riidogen: native config manifest schema version is required")
	}
	if !hasCompleteProviderPlan(spec.Default) {
		return errors.New("riidogen: native config default plan is incomplete")
	}
	return validateNativeConfigProviders(spec.Providers)
}

func validateNativeConfigProviders(providers []nativeConfigProviderPlanSpec) error {
	seen := map[string]struct{}{}
	for _, provider := range providers {
		if provider.ProviderKind == "" {
			return errors.New("riidogen: provider_kind is required")
		}
		if _, ok := seen[provider.ProviderKind]; ok {
			return fmt.Errorf("riidogen: duplicate provider_kind %q", provider.ProviderKind)
		}
		seen[provider.ProviderKind] = struct{}{}
		if !hasCompleteProviderPlan(provider) {
			return fmt.Errorf("riidogen: provider %q plan is incomplete", provider.ProviderKind)
		}
	}
	return nil
}

func hasCompleteProviderPlan(plan nativeConfigProviderPlanSpec) bool {
	return plan.PrimaryInstructionFile != "" && plan.ManifestFile != "" && plan.HookMode != ""
}
