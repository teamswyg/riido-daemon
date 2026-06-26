package main

type registrySource struct {
	Loops          []registryLoop  `json:"loops"`
	BusinessClaims []registryClaim `json:"business_claims"`
}

type registryLoop struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
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
