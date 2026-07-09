package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchLatestReleaseFailureShapes(t *testing.T) {
	cases := []struct {
		body string
		code int
		want string
	}{
		{body: `[]`, code: http.StatusOK, want: "latest release missing"},
		{body: `{`, code: http.StatusOK, want: "parse release"},
		{body: `nope`, code: http.StatusServiceUnavailable, want: "fetch release status"},
	}
	for _, tc := range cases {
		server := releaseAPIStatusServer(t, tc.code, tc.body)
		_, err := fetchLatestRelease(t.Context(), server)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("err=%v want %q", err, tc.want)
		}
	}
}

func TestFetchLatestReleaseRejectsBadURL(t *testing.T) {
	_, err := fetchLatestRelease(t.Context(), "://bad")
	if err == nil || !strings.Contains(err.Error(), "build release request") {
		t.Fatalf("expected request build error, got %v", err)
	}
}

func TestVerifyGitHubLatestRejectsDraftEmptyTagAndMissingAsset(t *testing.T) {
	cases := []struct {
		body string
		want string
	}{
		{body: releaseJSON("", false, expectedReleaseAsset(), "SHA256SUMS"), want: "tag is empty"},
		{body: releaseJSON("v-test", true, expectedReleaseAsset(), "SHA256SUMS"), want: "latest release is draft"},
		{body: releaseJSON("v-test", false, "SHA256SUMS"), want: "missing release asset"},
	}
	for _, tc := range cases {
		scenario, _ := verifyGitHubLatest(t.Context(), releaseAPIServer(t, tc.body))
		if scenario.Status != statusFailed || !strings.Contains(scenario.FailureSummary, tc.want) {
			t.Fatalf("scenario=%+v want %q", scenario, tc.want)
		}
	}
}

func releaseAPIStatusServer(t *testing.T, code int, body string) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(code)
		fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func releaseJSON(tag string, draft bool, assets ...string) string {
	return fmt.Sprintf(`[{"tag_name":%q,"draft":%v,"assets":[%s]}]`, tag, draft, assetJSON(assets...))
}

func assetJSON(assets ...string) string {
	out := make([]string, 0, len(assets))
	for _, asset := range assets {
		out = append(out, fmt.Sprintf(`{"name":%q}`, asset))
	}
	return strings.Join(out, ",")
}
