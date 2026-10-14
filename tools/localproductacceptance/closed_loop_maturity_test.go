package main

import (
	"bytes"
	"testing"
)

const (
	closedLoopMaturityDSLPath = "../../docs/30-architecture/closed-loop-maturity.dsl.json"
	closedLoopMaturityGenPath = "closed_loop_maturity.generated.json"
)

func TestClosedLoopMaturityGeneratedFileFresh(t *testing.T) {
	dsl := readFeatureUIJSON(t, closedLoopMaturityDSLPath)
	generated := readFeatureUIJSON(t, closedLoopMaturityGenPath)
	if !bytes.Equal(canonicalJSON(t, dsl), canonicalJSON(t, generated)) {
		t.Fatal("closed_loop_maturity.generated.json is stale; run go generate ./tools/localproductacceptance")
	}
}

func TestClosedLoopMaturityScenarioSurfacesPartialEvidence(t *testing.T) {
	got := closedLoopMaturityScenario()
	if got.ID != "local.qa.closed_loop_maturity" || got.Status != statusPartial {
		t.Fatalf("closed-loop maturity scenario = %+v", got)
	}
	if got.Observed["meta_complexity"] == nil || got.Observed["partial_reduction"] == nil {
		t.Fatalf("maturity evidence missing: %+v", got.Observed)
	}
	product := got.Observed["product_acceptance"].(map[string]any)
	if product["product_metric_count"] != 4 || product["linked_metric_count"] != 4 {
		t.Fatalf("product evidence linkage mismatch: %+v", product)
	}
}
