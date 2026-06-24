package main

import "strings"

func stagingFixtureHandoffProofFor(cfg config, domainRows []scenario) stagingFixtureHandoffProof {
	required := domainEntityKeys(domainEntityDefs())
	passed := passedDomainEntities(domainRows)
	return stagingFixtureHandoffProof{
		CachePath:       *cfg.domainCache,
		Remote:          domainRemoteEnvironment(*cfg.riidoAPIHost, *cfg.agentHost),
		Verification:    domainVerificationSource(*cfg.baseURL),
		HasToken:        strings.TrimSpace(*cfg.apiToken) != "",
		HasWorkspace:    strings.TrimSpace(*cfg.workspaceID) != "",
		Required:        required,
		PassedEntities:  passed,
		MissingEntities: missingDomainEntities(required, passed),
	}
}

func passedDomainEntities(rows []scenario) []string {
	var out []string
	for _, row := range rows {
		if strings.HasPrefix(row.ID, "domain.fixture.") && row.Status == statusPassed {
			out = append(out, strings.TrimPrefix(row.ID, "domain.fixture."))
		}
	}
	return out
}

func missingDomainEntities(required, passed []string) []string {
	seen := map[string]bool{}
	for _, key := range passed {
		seen[key] = true
	}
	var missing []string
	for _, key := range required {
		if !seen[key] {
			missing = append(missing, key)
		}
	}
	return missing
}
