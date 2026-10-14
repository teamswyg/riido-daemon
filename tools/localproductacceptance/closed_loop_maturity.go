package main

import (
	"embed"
	"encoding/json"
)

//go:embed closed_loop_maturity.generated.json
var closedLoopMaturityFS embed.FS

func closedLoopMaturityScenario() scenario {
	body, err := closedLoopMaturityFS.ReadFile("closed_loop_maturity.generated.json")
	if err != nil {
		return failedClosedLoopMaturityScenario("read closed-loop maturity DSL", err)
	}
	spec := closedLoopMaturitySpec{}
	if err := json.Unmarshal(body, &spec); err != nil {
		return failedClosedLoopMaturityScenario("parse closed-loop maturity DSL", err)
	}
	root, err := repoRoot()
	if err != nil {
		return failedClosedLoopMaturityScenario("find repo root", err)
	}
	product := closedLoopMaturityProduct(root, spec)
	partial := closedLoopMaturityPartial(spec)
	status := statusPassed
	if partial["inference_required_count"] != 0 {
		status = statusPartial
	}
	if len(product["unlinked_metric_ids"].([]string)) > 0 {
		status = statusPartial
	}
	return scenario{
		ID:     "local.qa.closed_loop_maturity",
		Status: status,
		Observed: map[string]any{
			"schema_version":     spec.SchemaVersion,
			"id":                 spec.ID,
			"generated":          spec.Generated,
			"meta_complexity":    closedLoopMaturityMeta(root),
			"product_acceptance": product,
			"partial_reduction":  partial,
			"partial_when":       spec.PartialWhen,
		},
	}
}

func failedClosedLoopMaturityScenario(summary string, err error) scenario {
	return scenario{
		ID:             "local.qa.closed_loop_maturity",
		Status:         statusFailed,
		FailureSummary: summary + ": " + err.Error(),
	}
}
