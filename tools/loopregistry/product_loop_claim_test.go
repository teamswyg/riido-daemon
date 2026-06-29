package main

import (
	"slices"
	"testing"
)

func TestProductLoopClaimBindsOperationalReadinessEvidence(t *testing.T) {
	reg, err := loadRegistry(repoPath(repoRoot(), defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	claim := findClaim(reg, "product_loop_evidence_must_measure_success_closure")
	for _, file := range []string{
		"tools/productloopevidence/ops_readiness_test.go",
		"tools/productloopevidence/ops_scenarios_test.go",
		".github/workflows/local-qa-runner.yml",
	} {
		if !slices.Contains(claim.Files, file) {
			t.Fatalf("claim missing file %s", file)
		}
	}
	if !slices.Contains(claim.Docs, "docs/30-architecture/local-acceptance-coverage.riido.json") {
		t.Fatal("claim missing local acceptance coverage doc")
	}
	if !slices.Contains(claim.Evidence, "client_surface_anomaly_monitoring") {
		t.Fatal("claim missing operational readiness evidence signal")
	}
	if !hasCheck(claim.Verifiers, "product-loop-operational-readiness-test") {
		t.Fatal("claim missing operational readiness verifier")
	}
	if !hasCheck(claim.Contracts, "local-qa-runner-operational-coverage") {
		t.Fatal("claim missing local QA runner operational contract")
	}
}

func findClaim(reg registry, id string) businessClaim {
	for _, claim := range reg.BusinessClaims {
		if claim.ID == id {
			return claim
		}
	}
	return businessClaim{}
}

func hasCheck(checks []sourceCheck, name string) bool {
	for _, check := range checks {
		if check.Name == name {
			return true
		}
	}
	return false
}
