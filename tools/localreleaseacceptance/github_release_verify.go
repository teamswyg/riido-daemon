package main

import (
	"context"
	"fmt"
)

func verifyGitHubLatest(ctx context.Context, url string) (scenario, githubRelease) {
	const id = "release.github.latest_assets"
	release, err := fetchLatestRelease(ctx, url)
	if err != nil {
		return failedScenario(id, "latest release unavailable: "+err.Error()), release
	}
	if release.TagName == "" {
		return failedScenario(id, "latest release tag is empty"), release
	}
	if release.Draft {
		return failedScenario(id, "latest release is draft: "+release.TagName), release
	}
	expected := expectedReleaseAsset()
	if !hasAsset(release, expected) {
		return failedScenario(id, fmt.Sprintf("missing release asset: %s", expected)), release
	}
	if !hasAsset(release, "SHA256SUMS") {
		return failedScenario(id, "missing release asset: SHA256SUMS"), release
	}
	return scenario{ID: id, Status: statusPassed}, release
}
