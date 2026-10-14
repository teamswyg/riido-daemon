package main

func buildEvidenceGraph(manifest Manifest, changed []string, results []bindingResult) []evidenceGraphEntry {
	changedSet := stringSet(changed)
	bindings := bindingByID(manifest)
	var graph []evidenceGraphEntry
	for _, result := range results {
		if !result.Triggered {
			continue
		}
		binding := bindings[result.ID]
		graph = append(graph, evidenceGraphEntry{
			BindingID:   result.ID,
			Observation: result.ChangedTriggers,
			Hypothesis:  result.Claim,
			Change:      changedSemanticPeers(binding, changedSet),
			Verifier:    binding.Verifiers,
			Evidence:    binding.EvidenceIDs,
			Decision:    graphDecision(result),
			NextLoop:    manifest.LoopSource,
		})
	}
	return graph
}

func bindingByID(manifest Manifest) map[string]Binding {
	out := map[string]Binding{}
	for _, binding := range manifest.Bindings {
		out[binding.ID] = binding
	}
	return out
}

func changedSemanticPeers(binding Binding, changed map[string]bool) []string {
	var out []string
	seen := map[string]bool{}
	for _, path := range binding.RequiredWithTriggers {
		if changed[path] && !seen[path] {
			out = append(out, path)
			seen[path] = true
		}
	}
	for _, path := range binding.GeneratedDocs {
		if changed[path] && !seen[path] {
			out = append(out, path)
			seen[path] = true
		}
	}
	return out
}

func graphDecision(result bindingResult) string {
	if len(result.MissingRequired) > 0 {
		return "failed_missing_semantic_peers"
	}
	return "verified"
}
