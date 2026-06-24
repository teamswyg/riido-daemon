package main

func inferredExecutionCandidates(item scenario) []evidenceGapCandidate {
	if item.ID != "local.qa.dsl_system_audit" {
		return nil
	}
	ids := inferredExecutionIDs(item.Observed)
	out := make([]evidenceGapCandidate, 0, len(ids))
	for _, id := range ids {
		out = append(out, closedLoopCandidate(
			"close-inferred-"+id,
			item.ID,
			"inferred_execution_loop",
			"QA execution inventory still requires human or Codex inference.",
			"Replace execution_inventory entry "+id+" with a system-owned command/verifier evidence path.",
		))
	}
	return out
}

func inferredExecutionIDs(observed map[string]any) []string {
	counts, ok := observed["execution_counts"].(map[string]any)
	if !ok {
		return nil
	}
	return stringList(counts["inference_required_ids"])
}

func stringList(value any) []string {
	switch ids := value.(type) {
	case []string:
		return ids
	case []any:
		return anyStringList(ids)
	default:
		return nil
	}
}

func anyStringList(values []any) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if text, ok := value.(string); ok {
			out = append(out, text)
		}
	}
	return out
}
