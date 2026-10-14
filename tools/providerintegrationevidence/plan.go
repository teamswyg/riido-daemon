package main

func buildQAPlan(m manifest) qaPlanEvidence {
	input, output := 0, 0
	for _, scenario := range m.Scenarios {
		input += scenario.EstimatedInputTokens
		output += scenario.EstimatedOutputTokens
	}
	monthly := (input + output) * m.ExecutionPolicy.RunsPerMonth
	status := "within_budget"
	if input > m.BudgetPolicy.MaxInputPerRun ||
		output > m.BudgetPolicy.MaxOutputPerRun ||
		monthly > m.BudgetPolicy.MaxTotalPerMonth {
		status = "over_budget"
	}
	return qaPlanEvidence{
		ExecutionPolicy: m.ExecutionPolicy,
		BudgetPolicy:    m.BudgetPolicy,
		Scenarios:       m.Scenarios,
		Promotion:       m.FailurePromotion,
		ProviderCount:   len(m.Providers),
		InputPerRun:     input,
		OutputPerRun:    output,
		TotalPerMonth:   monthly,
		FleetTokens:     monthly * len(m.Providers),
		FleetCashUSD:    m.BudgetPolicy.MonthlyCashUSD * len(m.Providers),
		BudgetStatus:    status,
		UsageStatus:     "estimated_before_run",
	}
}
