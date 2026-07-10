package main

import (
	"strings"
	"testing"
)

func TestCurrentProviderQAPlanFitsBudget(t *testing.T) {
	m := mustLoad(t, "../../docs/30-architecture/provider-real-cli-observation.riido.json")
	plan := buildQAPlan(m)
	if plan.BudgetStatus != "within_budget" || plan.TotalPerMonth != 24000 {
		t.Fatalf("plan=%+v", plan)
	}
	if plan.FleetTokens != 96000 || plan.FleetCashUSD != 80 {
		t.Fatalf("fleet plan=%+v", plan)
	}
}

func TestProviderQAPlanFailsClosedOverBudget(t *testing.T) {
	m := mustLoad(t, "../../docs/30-architecture/provider-real-cli-observation.riido.json")
	m.Scenarios[1].EstimatedInputTokens = m.BudgetPolicy.MaxInputPerRun + 1
	problems := validateQAPlan("../..", m)
	if !strings.Contains(strings.Join(problems, "\n"), "over_budget") {
		t.Fatalf("problems=%v", problems)
	}
}
