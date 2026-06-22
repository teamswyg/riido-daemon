package main

import (
	"encoding/json"
	"testing"
)

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

func TestAPIStatusCommandIncludesAppVersion(t *testing.T) {
	socketPath, stop := serveReviewDemoCLIAPI(t)
	defer stop()

	out, err := runCapturingStdout(t, func() error {
		return run([]string{"api", "status", "--socket", socketPath})
	})
	if err != nil {
		t.Fatalf("api status returned error: %v\n%s", err, out)
	}
	var status struct {
		AppVersion string `json:"app_version"`
	}
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		t.Fatalf("parse status: %v\n%s", err, out)
	}
	if status.AppVersion != "riido-daemon test.v1" {
		t.Fatalf("app_version = %q\n%s", status.AppVersion, out)
	}
}
