package main

import "fmt"

func validateQAPlan(root string, m manifest) []string {
	var problems []string
	policy := m.ExecutionPolicy
	if policy.PublicRunnerMode == "" || policy.PaidRunnerMode == "" ||
		policy.Cadence != "weekly" || policy.Timezone == "" || policy.RunsPerMonth <= 0 {
		problems = append(problems, "execution_policy needs runner modes, weekly cadence, timezone, and runs_per_month")
	}
	budget := m.BudgetPolicy
	if budget.MonthlyCashUSD <= 0 || budget.MaxInputPerRun <= 0 ||
		budget.MaxOutputPerRun <= 0 || budget.MaxTotalPerMonth <= 0 ||
		!budget.FailClosed || budget.UsageAccounting == "" {
		problems = append(problems, "budget_policy must define positive fail-closed cash and token limits")
	}
	problems = append(problems, validateScenarios(m)...)
	problems = append(problems, mustExist(root, m.FailurePromotion.CandidateProducer)...)
	problems = append(problems, mustExist(root, m.FailurePromotion.DecisionVerifier)...)
	if m.FailurePromotion.AutomationMode == "" || len(m.FailurePromotion.FailureStates) == 0 ||
		m.FailurePromotion.MergePolicy != "pull_request_ci" {
		problems = append(problems, "failure_promotion needs automation, failure states, and pull_request_ci")
	}
	if plan := buildQAPlan(m); plan.BudgetStatus != "within_budget" {
		problems = append(problems, fmt.Sprintf("estimated provider QA plan is %s", plan.BudgetStatus))
	}
	return problems
}
