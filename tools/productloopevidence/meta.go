package main

func buildMetaComplexity(root string, m manifest, reg registrySource) metaComplexity {
	toolEntrypoints := countFiles(root, "tools/*/main.go")
	workflows := countFiles(root, ".github/workflows/*.yml")
	verifiers := countFiles(root, "**/*_test.go")
	generatedDocs := countGeneratedDocs(root)
	entrypoints := toolEntrypoints + workflows
	coverage := buildMappingCoverage(reg)
	out := metaComplexity{
		ToolEntrypointCount: toolEntrypoints,
		WorkflowCount:       workflows,
		VerifierFileCount:   verifiers,
		GeneratedDocCount:   generatedDocs,
		EntrypointCount:     entrypoints,
		EntrypointThreshold: m.Thresholds.MaxEntrypointsBeforePartial,
		MappingCoverage:     coverage,
		Status:              statusPassed,
	}
	if entrypoints > m.Thresholds.MaxEntrypointsBeforePartial {
		out.Status = statusPartial
		out.PartialReason = "entrypoint_count exceeds max_entrypoints_before_partial"
	}
	if coverage.ClaimCount == 0 || coverage.CoverageRatio < 1 {
		out.Status = statusPartial
		if out.PartialReason == "" {
			out.PartialReason = "claim mapping coverage is incomplete"
		}
	}
	return out
}

func buildMappingCoverage(reg registrySource) mappingCoverage {
	files := map[string]bool{}
	withVerifier := 0
	verifierCount := 0
	for _, claim := range reg.BusinessClaims {
		for _, file := range claim.Files {
			files[file] = true
		}
		if len(claim.Verifiers) > 0 {
			withVerifier++
		}
		verifierCount += len(claim.Verifiers)
	}
	return mappingCoverage{
		ClaimCount:             len(reg.BusinessClaims),
		ClaimWithVerifierCount: withVerifier,
		BoundFileCount:         len(files),
		DeclaredVerifierCount:  verifierCount,
		CoverageRatio:          ratio(withVerifier, len(reg.BusinessClaims)),
	}
}
