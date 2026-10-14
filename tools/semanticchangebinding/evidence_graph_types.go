package main

type evidenceGraphEntry struct {
	BindingID   string   `json:"binding_id"`
	Observation []string `json:"observation"`
	Hypothesis  string   `json:"hypothesis"`
	Change      []string `json:"change"`
	Verifier    []string `json:"verifier"`
	Evidence    []string `json:"evidence"`
	Decision    string   `json:"decision"`
	NextLoop    string   `json:"next_loop,omitempty"`
}
