package main

import (
	"strings"
	"testing"
)

func TestValidateNativeConfigPlanRejectsDuplicateProvider(t *testing.T) {
	spec := validNativeConfigPlanSpec()
	spec.Providers = append(spec.Providers, spec.Providers[0])

	err := validateNativeConfigPlan(spec)
	if err == nil {
		t.Fatal("expected duplicate provider error")
	}
	if !strings.Contains(err.Error(), `duplicate provider_kind "claude"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSortNativeConfigProviders(t *testing.T) {
	providers := []nativeConfigProviderPlanSpec{
		{ProviderKind: "codex"},
		{ProviderKind: "claude"},
	}
	sortNativeConfigProviders(providers)
	if providers[0].ProviderKind != "claude" || providers[1].ProviderKind != "codex" {
		t.Fatalf("providers were not sorted: %#v", providers)
	}
}
