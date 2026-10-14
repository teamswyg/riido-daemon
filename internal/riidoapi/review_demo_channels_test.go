package riidoapi

import (
	"context"
	"testing"
)

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
	if response.Enabled || response.ProviderStatusMode != "real-status" || len(response.Surfaces) != 0 {
		t.Fatalf("unexpected non-store response: %#v", response)
	}
}
