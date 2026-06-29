package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func closedLoopMaturityProduct(root string, spec closedLoopMaturitySpec) map[string]any {
	coverage := closedLoopCoverage{}
	body, err := os.ReadFile(filepath.Join(root,
		"tools/localproductacceptance/local_acceptance_coverage.generated.json"))
	if err == nil {
		_ = json.Unmarshal(body, &coverage)
	}
	scenarioIDs := map[string]bool{}
	for _, row := range coverage.Scenarios {
		scenarioIDs[row.ID] = true
	}
	unlinked := []string{}
	linked := 0
	total := 0
	for _, metric := range spec.Metrics {
		if metric.Class != "product-acceptance" {
			continue
		}
		total++
		if scenarioIDs[metric.Evidence] {
			linked++
			continue
		}
		unlinked = append(unlinked, metric.ID)
	}
	return map[string]any{
		"product_metric_count": total,
		"linked_metric_count":  linked,
		"unlinked_metric_ids":  unlinked,
	}
}

func closedLoopMaturityPartial(spec closedLoopMaturitySpec) map[string]any {
	inferred := []string{}
	partialMetrics := 0
	for _, metric := range spec.Metrics {
		if metric.Mode == "inferred" {
			inferred = append(inferred, metric.ID)
		}
		if metric.Class == "partial-reduction" {
			partialMetrics++
		}
	}
	return map[string]any{
		"inference_required_ids":      inferred,
		"inference_required_count":    len(inferred),
		"closed_loop_candidate_count": partialMetrics,
		"stale_after_days":            spec.StaleAfter,
		"candidate_age_days":          "inferred",
		"promoted_count":              0,
	}
}
