package providervalidation

import "slices"

func hasString(items []string, want string) bool {
	return slices.Contains(items, want)
}

func hasAny(items []string, wants ...string) bool {
	for _, want := range wants {
		if hasString(items, want) {
			return true
		}
	}
	return false
}
