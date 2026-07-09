package main

import (
	"strings"
	"testing"
)

func TestRunIntegrationTestReportsPassed(t *testing.T) {
	status, summary := runIntegrationTest("../..", provider{
		GoPackage: "./tools/providerintegrationevidence",
		TestRegex: "^TestRunIntegrationPassFixture$",
	})
	if status != "passed" || summary != "" {
		t.Fatalf("status=%q summary=%q", status, summary)
	}
}

func TestRunIntegrationTestReportsSkipped(t *testing.T) {
	status, summary := runIntegrationTest("../..", provider{
		GoPackage: "./tools/providerintegrationevidence",
		TestRegex: "^TestRunIntegrationSkipFixture$",
	})
	if status != "skipped" || !strings.Contains(summary, "fixture skip") {
		t.Fatalf("status=%q summary=%q", status, summary)
	}
}

func TestRunIntegrationPassFixture(t *testing.T) {
	if !strings.Contains(t.Name(), "PassFixture") {
		t.Fatal("unexpected fixture name")
	}
}

func TestRunIntegrationSkipFixture(t *testing.T) {
	t.Skip("fixture skip")
}
