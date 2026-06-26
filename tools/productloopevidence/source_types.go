package main

type registrySource struct {
	Loops          []registryLoop  `json:"loops"`
	BusinessClaims []registryClaim `json:"business_claims"`
}

type registryLoop struct {
	ID                 string `json:"id"`
	Kind               string `json:"kind"`
	CandidateCreatedAt string `json:"candidate_created_at"`
	PromotionTarget    string `json:"promotion_target"`
}

type registryClaim struct {
	ID        string        `json:"id"`
	Files     []string      `json:"files"`
	Evidence  []string      `json:"evidence"`
	Verifiers []sourceCheck `json:"verifiers"`
}

type sourceCheck struct {
	Name string `json:"name"`
	File string `json:"file"`
}

type localAcceptanceSource struct {
	Scenarios []coverageScenario `json:"scenarios"`
}

type coverageScenario struct {
	ID string `json:"id"`
}

type qaSystemSource struct {
	ExecutionInventory []qaExecution `json:"execution_inventory"`
}

type qaExecution struct {
	ID   string `json:"id"`
	Mode string `json:"mode"`
}

type qaScheduleSource struct {
	ID              string   `json:"id"`
	Cadence         string   `json:"cadence"`
	Entrypoint      string   `json:"entrypoint"`
	FreshnessWindow string   `json:"freshness_window"`
	Evidence        []string `json:"evidence"`
	ClosedLoop      []string `json:"closed_loop"`
	RejectIf        []string `json:"reject_if"`
}
