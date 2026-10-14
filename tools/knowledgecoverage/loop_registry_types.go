package main

type loopRegistryEntry struct {
	ID           string   `json:"id"`
	LoopSource   string   `json:"loop_source"`
	Observes     []string `json:"observes"`
	Verifies     []string `json:"verifies"`
	Evidence     []string `json:"evidence"`
	ExpiresAfter string   `json:"expires_after"`
	FailsWhen    []string `json:"fails_when"`
}

type loopRegistrySummary struct {
	Count       int      `json:"count"`
	IDs         []string `json:"ids"`
	Expires     []string `json:"expires_after"`
	EvidenceRef int      `json:"evidence_ref_count"`
}
