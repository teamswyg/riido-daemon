package main

import "testing"

func TestAPIReviewDemoCommandUsesLocalControlSurface(t *testing.T) {
	socketPath, stop := serveReviewDemoCLIAPI(t)
	defer stop()

	err := run([]string{
		"api", "review-demo",
		"--socket", socketPath,
		"--channel", "mac-app-store",
		"--review-demo-consent-granted", "true",
	})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}
}
