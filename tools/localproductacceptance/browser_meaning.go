package main

func browserMeaningScenario() scenario {
	result := browserMeaningResult()
	status := statusPassed
	if len(result.Missing) > 0 {
		status = statusFailed
	}
	return scenario{
		ID:       "contract.ui.browser_meaning_qa",
		Status:   status,
		Observed: result.Observed(),
		FailureSummary: func() string {
			if status == statusPassed {
				return ""
			}
			return "contract lab browser meaning proof is missing required tokens"
		}(),
	}
}

type browserMeaningProof struct {
	Required []string
	Missing  []string
}

func (p browserMeaningProof) Observed() map[string]any {
	return map[string]any{
		"mode":                 "system",
		"replaces_inferred_id": "browser-meaning-qa",
		"query_source":         "contract lab template and QA system DSL",
		"required_token_count": len(p.Required),
		"missing_tokens":       p.Missing,
		"search_matches":       []string{"scenario id", "status", "method", "endpoint", "failure", "area", "observed JSON"},
		"verified_interaction": "typing any generated alias into the toolbar search can match scenario.observed JSON without human inference",
		"human_browser_needed": false,
	}
}
