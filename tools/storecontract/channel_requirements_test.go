package main

import "testing"

func TestRunRejectsStoreChannelWithoutReviewDemoMode(t *testing.T) {
	expectContractMutationFailure(
		t,
		removeChannelRequiredSurface("mac-app-store", "review-demo-mode"),
		`channel "mac-app-store" must require review-demo-mode`,
	)
}

func TestRunRejectsStoreChannelWithoutReviewSubmissionSurface(t *testing.T) {
	tests := []contractMutationCase{
		{
			name:   "mac app store requires demo review account",
			mutate: removeChannelRequiredSurface("mac-app-store", "demo-review-account"),
			error:  `channel "mac-app-store" must require demo-review-account`,
		},
		{
			name:   "mac app store requires privacy metadata allowlist",
			mutate: removeChannelRequiredSurface("mac-app-store", "privacy-metadata-allowlist"),
			error:  `channel "mac-app-store" must require privacy-metadata-allowlist`,
		},
		{
			name:   "mac app store requires provider non bundling review note",
			mutate: removeChannelRequiredSurface("mac-app-store", "provider-non-bundling-review-note"),
			error:  `channel "mac-app-store" must require provider-non-bundling-review-note`,
		},
		{
			name:   "microsoft store requires demo review account",
			mutate: removeChannelRequiredSurface("msix-store", "demo-review-account"),
			error:  `channel "msix-store" must require demo-review-account`,
		},
		{
			name:   "microsoft store requires privacy metadata allowlist",
			mutate: removeChannelRequiredSurface("msix-store", "privacy-metadata-allowlist"),
			error:  `channel "msix-store" must require privacy-metadata-allowlist`,
		},
		{
			name:   "microsoft store requires provider non bundling review note",
			mutate: removeChannelRequiredSurface("msix-store", "provider-non-bundling-review-note"),
			error:  `channel "msix-store" must require provider-non-bundling-review-note`,
		},
	}
	expectContractMutationFailures(t, tests)
}
