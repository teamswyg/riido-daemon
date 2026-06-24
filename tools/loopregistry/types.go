package main

type options struct {
	Manifest          string
	EvidenceOut       string
	ChangedFiles      string
	WriteDoc          bool
	CheckDoc          bool
	GitHubAnnotations bool
}

type registry struct {
	SchemaVersion    string          `json:"schema_version"`
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	GeneratedDoc     string          `json:"generated_doc"`
	Workflow         string          `json:"workflow"`
	EvidenceArtifact string          `json:"evidence_artifact"`
	PrecommitHook    string          `json:"precommit_hook"`
	Command          string          `json:"command"`
	Loop             evidenceLoop    `json:"loop"`
	Loops            []loopEntry     `json:"loops"`
	BusinessClaims   []businessClaim `json:"business_claims"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

type loopEntry struct {
	ID           string        `json:"id"`
	Owner        string        `json:"owner"`
	Kind         string        `json:"kind"`
	Observes     []string      `json:"observes"`
	Verifies     []string      `json:"verifies"`
	Evidence     []string      `json:"evidence"`
	ExpiresAfter string        `json:"expires_after"`
	FailsWhen    []string      `json:"fails_when"`
	Graph        evidenceGraph `json:"evidence_graph"`
}

type evidenceGraph struct {
	Observation string `json:"observation"`
	Hypothesis  string `json:"hypothesis"`
	Change      string `json:"change"`
	Verifier    string `json:"verifier"`
	Evidence    string `json:"evidence"`
	Decision    string `json:"decision"`
	NextLoop    string `json:"next_loop"`
}

type businessClaim struct {
	ID        string        `json:"id"`
	Text      string        `json:"text"`
	Files     []string      `json:"files"`
	Docs      []string      `json:"docs"`
	Evidence  []string      `json:"evidence"`
	Verifiers []sourceCheck `json:"verifiers"`
	Contracts []sourceCheck `json:"contracts,omitempty"`
}

type sourceCheck struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Contains []string `json:"contains"`
}
