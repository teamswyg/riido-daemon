package main

func buildReport(root string, m manifest) (report, error) {
	reg, local, qa, err := loadSources(root, m)
	if err != nil {
		return report{}, err
	}
	routes, err := loadEntrypointRouteMap(root, m)
	if err != nil {
		return report{}, err
	}
	meta := buildMetaComplexity(root, m, reg, routes)
	product := buildProductAcceptance(m, local)
	partial := buildPartialReduction(root, m, reg, qa)
	candidates := collectCandidates(meta, product, partial)
	out := report{
		SchemaVersion:     reportSchema,
		ID:                m.ID,
		Status:            statusPassed,
		GeneratedDoc:      m.GeneratedDoc,
		Workflow:          m.Workflow,
		EvidenceArtifact:  m.EvidenceArtifact,
		MetaComplexity:    meta,
		ProductAcceptance: product,
		PartialReduction:  partial,
		Candidates:        candidates,
	}
	out.Status = aggregateStatus(meta.Status, product.Status, partial.Status)
	return out, nil
}

func loadSources(root string, m manifest) (registrySource, localAcceptanceSource, qaSystemSource, error) {
	var reg registrySource
	var local localAcceptanceSource
	var qa qaSystemSource
	if err := loadJSON(repoPath(root, m.LoopRegistry), &reg); err != nil {
		return reg, local, qa, err
	}
	if err := loadJSON(repoPath(root, m.LocalAcceptanceManifest), &local); err != nil {
		return reg, local, qa, err
	}
	if err := loadJSON(repoPath(root, m.QASystemManifest), &qa); err != nil {
		return reg, local, qa, err
	}
	return reg, local, qa, nil
}

func aggregateStatus(values ...string) string {
	status := statusPassed
	for _, value := range values {
		if value == statusFailed {
			return statusFailed
		}
		if value == statusPartial {
			status = statusPartial
		}
	}
	return status
}
