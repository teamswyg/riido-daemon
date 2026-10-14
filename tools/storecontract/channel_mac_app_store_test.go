package main

import "testing"

func TestRunRejectsMacAppStoreRequiredSurfaces(t *testing.T) {
	cases := []contractMutationCase{
		{
			name: "sandbox review notes",
			mutate: removeChannelRequiredSurface(
				"mac-app-store",
				"app-sandbox-entitlement-review-notes",
			),
			error: `channel "mac-app-store" must require app-sandbox-entitlement-review-notes`,
		},
		{
			name:   "helper purpose review note",
			mutate: removeChannelRequiredSurface("mac-app-store", "helper-purpose-review-note"),
			error:  `channel "mac-app-store" must require helper-purpose-review-note`,
		},
	}
	expectContractMutationFailures(t, cases)
}
