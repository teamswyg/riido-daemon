package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func fetchLatestRelease(ctx context.Context, url string) (githubRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return githubRelease{}, fmt.Errorf("build release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	applyGitHubReleaseAuth(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return githubRelease{}, fmt.Errorf("fetch release: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return githubRelease{}, fmt.Errorf("fetch release status: %s", res.Status)
	}
	var releases []githubRelease
	if err := json.NewDecoder(res.Body).Decode(&releases); err != nil {
		return githubRelease{}, fmt.Errorf("parse release: %w", err)
	}
	if len(releases) == 0 {
		return githubRelease{}, fmt.Errorf("latest release missing")
	}
	return releases[0], nil
}

func applyGitHubReleaseAuth(req *http.Request) {
	if req.URL.Host != "api.github.com" {
		return
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}
