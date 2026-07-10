package main

import (
	"fmt"
	"strings"
)

func renderQAPlan(m manifest) string {
	plan := buildQAPlan(m)
	var b strings.Builder
	b.WriteString("\n## Paid QA Plan\n\n")
	fmt.Fprintf(&b, "- Paid execution: `%s`; public runner: `%s`.\n",
		m.ExecutionPolicy.PaidRunnerMode, m.ExecutionPolicy.PublicRunnerMode)
	fmt.Fprintf(&b, "- Cadence: `%s` in `%s`, %d runs per month.\n",
		m.ExecutionPolicy.Cadence, m.ExecutionPolicy.Timezone, m.ExecutionPolicy.RunsPerMonth)
	fmt.Fprintf(&b, "- Cash cap: `$%d/provider/month`; fleet cap: `$%d/month`.\n",
		m.BudgetPolicy.MonthlyCashUSD, plan.FleetCashUSD)
	fmt.Fprintf(&b, "- Estimated tokens: `%d input + %d output/provider/run`, `%d/provider/month`, `%d/fleet/month`.\n",
		plan.InputPerRun, plan.OutputPerRun, plan.TotalPerMonth, plan.FleetTokens)
	fmt.Fprintf(&b, "- Budget status: `%s`; usage accounting: `%s`.\n",
		plan.BudgetStatus, m.BudgetPolicy.UsageAccounting)
	b.WriteString("\n| Scenario | Execution | Paid | Input | Output | Timeout |\n")
	b.WriteString("| --- | --- | --- | ---: | ---: | ---: |\n")
	for _, scenario := range m.Scenarios {
		fmt.Fprintf(&b, "| `%s` | `%s` | %t | %d | %d | %ds |\n", scenario.ID,
			scenario.Execution, scenario.Paid, scenario.EstimatedInputTokens,
			scenario.EstimatedOutputTokens, scenario.MaxDurationSeconds)
	}
	fmt.Fprintf(&b, "\nFailures flow through `%s` then `%s`; merges require `%s`.\n",
		m.FailurePromotion.CandidateProducer, m.FailurePromotion.DecisionVerifier,
		m.FailurePromotion.MergePolicy)
	return b.String()
}
