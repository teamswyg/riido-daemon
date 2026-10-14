package main

import "sort"

func sortNativeConfigProviders(providers []nativeConfigProviderPlanSpec) {
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].ProviderKind < providers[j].ProviderKind
	})
}
