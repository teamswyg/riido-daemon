package riidoapi

import (
	"context"
	"testing"
)

func TestServerEvaluatesReviewDemoMode(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var response ReviewDemoResponse
	err := client.Request(context.Background(), "review-demo", ReviewDemoRequest{
		DistributionChannel:      "mac-app-store",
		ReviewDemoConsentGranted: true,
	}, &response)
	if err != nil {
		t.Fatalf("review-demo request failed: %v", err)
	}
	assertReviewDemoEnabled(t, response)
}

func assertReviewDemoEnabled(t *testing.T, response ReviewDemoResponse) {
	t.Helper()
	if response.SchemaVersion != ReviewDemoSchemaVersion || !response.Enabled {
		t.Fatalf("unexpected review-demo response: %#v", response)
	}
	if response.ProviderExecutionAllowed || response.TelemetrySyncAllowed || !response.LocalOnly {
		t.Fatalf("unexpected review-demo guard flags: %#v", response)
	}
	if response.ProviderStatusMode != "synthetic-preview" {
		t.Fatalf("unexpected provider status mode: %s", response.ProviderStatusMode)
	}
	want := []string{"onboarding", "provider-status", "workspace-grant", "background-consent", "privacy-settings", "local-status"}
	if !sameStrings(response.Surfaces, want) {
		t.Fatalf("unexpected surfaces: got %#v want %#v", response.Surfaces, want)
	}
}
