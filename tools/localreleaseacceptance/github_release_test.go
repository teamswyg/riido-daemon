package main

import (
	"net/http"
	"testing"
)

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

func TestApplyGitHubReleaseAuthUsesActionsTokenForGitHubAPIOnly(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "ghs_test")
	githubReq, err := http.NewRequest(http.MethodGet, defaultReleaseAPIURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	applyGitHubReleaseAuth(githubReq)
	if got := githubReq.Header.Get("Authorization"); got != "Bearer ghs_test" {
		t.Fatalf("github authorization = %q", got)
	}

	localReq, err := http.NewRequest(http.MethodGet, "https://example.invalid/releases", nil)
	if err != nil {
		t.Fatal(err)
	}
	applyGitHubReleaseAuth(localReq)
	if got := localReq.Header.Get("Authorization"); got != "" {
		t.Fatalf("non-github authorization leaked: %q", got)
	}
}
