package main

type policyTableRow struct {
	Surface string            `json:"surface"`
	Cells   []policyTableCell `json:"cells"`
}

type policyTableCell struct {
	Channel       string   `json:"channel"`
	Decision      string   `json:"decision"`
	Code          string   `json:"code"`
	RequiredFacts []string `json:"required_facts,omitempty"`
}

type policySurfaceSpec struct {
	ID    string
	Label string
}

type policyFactScenario struct {
	Facts       []string
	Consent     bool
	OSGrant     bool
	StoreReview bool
}
