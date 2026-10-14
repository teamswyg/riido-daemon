package main

import "testing"

func TestRunRejectsDeveloperIDWithoutNotarization(t *testing.T) {
	expectContractMutationFailure(
		t,
		removeChannelRequiredSurface("developer-id", "notarization"),
		`channel "developer-id" must require notarization`,
	)
}
