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
	if response.SchemaVersion != ReviewDemoSchemaVersion {
		t.Fatalf("unexpected review-demo schema: %s", response.SchemaVersion)
	}
	if !response.Enabled {
		t.Fatal("review demo mode should be enabled")
	}
	if response.ProviderExecutionAllowed {
		t.Fatal("review demo mode must not allow provider execution")
	}
	if response.TelemetrySyncAllowed {
		t.Fatal("review demo mode must not allow telemetry sync")
	}
	if !response.LocalOnly {
		t.Fatal("review demo mode should be reported as local-only")
	}
	if response.ProviderStatusMode != "synthetic-preview" {
		t.Fatalf("unexpected provider status mode: %s", response.ProviderStatusMode)
	}
	want := []string{"onboarding", "provider-status", "workspace-grant", "background-consent", "privacy-settings", "local-status"}
	if !sameStrings(response.Surfaces, want) {
		t.Fatalf("unexpected surfaces: got %#v want %#v", response.Surfaces, want)
	}
}

func TestServerReviewDemoRequiresConsentForStoreManagedChannel(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var response ReviewDemoResponse
	err := client.Request(context.Background(), "review-demo", ReviewDemoRequest{
		DistributionChannel: "msix-store",
	}, &response)
	if err == nil {
		t.Fatal("expected review-demo request without consent to fail")
	}
}

func TestServerReviewDemoIgnoresNonStoreManagedChannel(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var response ReviewDemoResponse
	err := client.Request(context.Background(), "review-demo", ReviewDemoRequest{
		DistributionChannel: "developer-id",
	}, &response)
	if err != nil {
		t.Fatalf("review-demo non-store request failed: %v", err)
	}
	if response.Enabled {
		t.Fatal("non-store channel should not enable review demo mode")
	}
	if response.ProviderStatusMode != "real-status" {
		t.Fatalf("unexpected provider status mode: %s", response.ProviderStatusMode)
	}
	if len(response.Surfaces) != 0 {
		t.Fatalf("non-store channel should not expose synthetic surfaces: %#v", response.Surfaces)
	}
}
