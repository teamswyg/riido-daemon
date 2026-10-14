package main

type executionPolicy struct {
	PublicRunnerMode string `json:"public_runner_mode"`
	PaidRunnerMode   string `json:"paid_runner_mode"`
	Cadence          string `json:"cadence"`
	Timezone         string `json:"timezone"`
	RunsPerMonth     int    `json:"runs_per_month"`
}

type budgetPolicy struct {
	MonthlyCashUSD   int    `json:"monthly_cash_budget_usd_per_provider"`
	MaxInputPerRun   int    `json:"max_estimated_input_tokens_per_run"`
	MaxOutputPerRun  int    `json:"max_estimated_output_tokens_per_run"`
	MaxTotalPerMonth int    `json:"max_estimated_total_tokens_per_provider_month"`
	FailClosed       bool   `json:"fail_closed"`
	UsageAccounting  string `json:"usage_accounting"`
}

type qaScenario struct {
	ID                    string `json:"id"`
	Execution             string `json:"execution"`
	Paid                  bool   `json:"paid"`
	EstimatedInputTokens  int    `json:"estimated_input_tokens"`
	EstimatedOutputTokens int    `json:"estimated_output_tokens"`
	MaxDurationSeconds    int    `json:"max_duration_seconds"`
}

type failurePromotion struct {
	CandidateProducer string   `json:"candidate_producer"`
	DecisionVerifier  string   `json:"decision_verifier"`
	AutomationMode    string   `json:"automation_mode"`
	FailureStates     []string `json:"failure_states"`
	MergePolicy       string   `json:"merge_policy"`
}

type qaPlanEvidence struct {
	ExecutionPolicy executionPolicy  `json:"execution_policy"`
	BudgetPolicy    budgetPolicy     `json:"budget_policy"`
	Scenarios       []qaScenario     `json:"scenarios"`
	Promotion       failurePromotion `json:"failure_promotion"`
	ProviderCount   int              `json:"provider_count"`
	InputPerRun     int              `json:"estimated_input_tokens_per_provider_run"`
	OutputPerRun    int              `json:"estimated_output_tokens_per_provider_run"`
	TotalPerMonth   int              `json:"estimated_total_tokens_per_provider_month"`
	FleetTokens     int              `json:"estimated_fleet_tokens_per_month"`
	FleetCashUSD    int              `json:"fleet_monthly_cash_budget_usd"`
	BudgetStatus    string           `json:"budget_status"`
	UsageStatus     string           `json:"usage_status"`
}
