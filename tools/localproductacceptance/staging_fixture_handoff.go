package main

func stagingFixtureHandoffScenario(cfg config, domainRows []scenario) scenario {
	proof := stagingFixtureHandoffProofFor(cfg, domainRows)
	status := statusPassed
	var repairEvidence *repair
	if len(proof.MissingEntities) > 0 || !proof.HasToken || !proof.HasWorkspace {
		status = statusPartial
		repairEvidence = &repair{
			Class:   "staging_fixture_handoff_required",
			Owner:   "local-qa",
			Mode:    "automated-with-token",
			Summary: "Run local product acceptance with staging token/workspace inputs or import the domain fixture cache.",
		}
	}
	return scenario{
		ID:       "local.qa.staging_fixture_handoff",
		Status:   status,
		Observed: proof.Observed(),
		Repair:   repairEvidence,
	}
}

type stagingFixtureHandoffProof struct {
	CachePath       string
	Remote          string
	Verification    string
	HasToken        bool
	HasWorkspace    bool
	Required        []string
	PassedEntities  []string
	MissingEntities []string
}

func (p stagingFixtureHandoffProof) Observed() map[string]any {
	return map[string]any{
		"mode":                 "system",
		"replaces_inferred_id": "staging-fixture-handoff",
		"cache_path":           p.CachePath,
		"remote_environment":   p.Remote,
		"verification_source":  p.Verification,
		"has_token":            p.HasToken,
		"has_workspace":        p.HasWorkspace,
		"required_entities":    p.Required,
		"passed_entities":      p.PassedEntities,
		"missing_entities":     p.MissingEntities,
		"entrypoint":           "go run ./tools/localproductacceptance -run-task-mutations",
	}
}
