package main

func dashboardRenderExpectations() []string {
	return []string{
		"Coverage Status",
		"freshness-status",
		`data-expires="2999-06-23T01:00:00Z"`,
		">fresh</div>",
		"Deployment Gate",
		"blocked",
		"-strict-coverage",
		"provider_auth_required",
		"cursor-agent login",
		"Closed-Loop Candidates",
		"coverage.product",
		"12h",
		"stale 2026-06-25T00:00:00Z",
		"candidate_for_promotion",
		"closed-loop.coverage-product",
		"coverage.local-qa-daily-freshness",
		"promoted_to_closed_loop",
		"promote to verifier",
		"figma.onboarding",
		"figma.json",
		"expires 2999-06-23T01:00:00Z",
		`<img class="shot"`,
		"passed",
	}
}
