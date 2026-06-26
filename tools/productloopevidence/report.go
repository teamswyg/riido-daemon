package main

func buildReport(root string, m manifest) (report, error) {
	reg, local, qa, schedule, err := loadSources(root, m)
	if err != nil {
		return report{}, err
	}
	routes, err := loadEntrypointRouteMap(root, m)
	if err != nil {
		return report{}, err
	}
	meta := buildMetaComplexity(root, m, reg, routes)
	product := buildProductAcceptance(m, local)
	qaSchedule := buildQASchedule(m, schedule)
	partial := buildPartialReduction(root, m, reg, qa)
	candidates := collectCandidates(meta, product, qaSchedule, partial)
	out := report{
		SchemaVersion:     reportSchema,
		ID:                m.ID,
		Status:            statusPassed,
		GeneratedDoc:      m.GeneratedDoc,
		Workflow:          m.Workflow,
		EvidenceArtifact:  m.EvidenceArtifact,
		MetaComplexity:    meta,
		ProductAcceptance: product,
		QASchedule:        qaSchedule,
		PartialReduction:  partial,
		Candidates:        candidates,
	}
	out.Status = aggregateStatus(meta.Status, product.Status, qaSchedule.Status, partial.Status)
	return out, nil
}

func loadSources(root string, m manifest) (registrySource, localAcceptanceSource, qaSystemSource, qaScheduleSource, error) {
	var reg registrySource
	var local localAcceptanceSource
	var qa qaSystemSource
	var schedule qaScheduleSource
	if err := loadJSON(repoPath(root, m.LoopRegistry), &reg); err != nil {
		return reg, local, qa, schedule, err
	}
	if err := loadJSON(repoPath(root, m.LocalAcceptanceManifest), &local); err != nil {
		return reg, local, qa, schedule, err
	}
	if err := loadJSON(repoPath(root, m.QASystemManifest), &qa); err != nil {
		return reg, local, qa, schedule, err
	}
	if err := loadJSON(repoPath(root, m.LocalQAScheduleManifest), &schedule); err != nil {
		return reg, local, qa, schedule, err
	}
	return reg, local, qa, schedule, nil
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
