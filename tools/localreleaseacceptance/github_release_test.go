package main

import "testing"

func TestVerifyGitHubLatestRequiresChecksum(t *testing.T) {
	api := releaseAPIServer(t, releaseBody(expectedReleaseAsset()))
	scenario, _ := verifyGitHubLatest(t.Context(), api)
	if scenario.Status != statusFailed {
		t.Fatalf("expected failure: %+v", scenario)
	}
	if scenario.ID != "release.github.latest_assets" {
		t.Fatalf("scenario id=%q", scenario.ID)
	}
}
