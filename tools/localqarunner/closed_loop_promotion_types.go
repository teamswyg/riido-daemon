package main

type closedLoopPromotionRegistry struct {
	SchemaVersion string                `json:"schema_version"`
	ID            string                `json:"id"`
	Promotions    []closedLoopPromotion `json:"promotions"`
}

type closedLoopPromotion struct {
	CandidateID string `json:"candidate_id"`
	LoopID      string `json:"loop_id"`
	LoopSource  string `json:"loop_source"`
	Verifier    string `json:"verifier"`
	Evidence    string `json:"evidence"`
	Decision    string `json:"decision"`
}
