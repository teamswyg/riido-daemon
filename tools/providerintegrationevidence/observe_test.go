package main

import "testing"

func TestAggregateStatusPartialForMixedPassAndSkip(t *testing.T) {
	providers := []providerEvidence{
		{IntegrationStatus: "passed"},
		{IntegrationStatus: "skipped"},
	}
	if got := aggregateStatus(providers); got != "partial" {
		t.Fatalf("aggregateStatus=%q, want partial", got)
	}
}

func TestAggregateStatusFailedWins(t *testing.T) {
	providers := []providerEvidence{
		{IntegrationStatus: "passed"},
		{IntegrationStatus: "failed"},
	}
	if got := aggregateStatus(providers); got != "failed" {
		t.Fatalf("aggregateStatus=%q, want failed", got)
	}
}
