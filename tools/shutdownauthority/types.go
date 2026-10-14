package main

type Manifest struct {
	SchemaVersion        string                `json:"schema_version"`
	ID                   string                `json:"id"`
	Title                string                `json:"title"`
	GeneratedDoc         string                `json:"generated_doc"`
	Workflow             string                `json:"workflow"`
	EvidenceArtifact     string                `json:"evidence_artifact"`
	Sources              Sources               `json:"sources"`
	ExternalSignals      []ExternalSignal      `json:"external_signals"`
	Levels               []Level               `json:"levels"`
	Timeouts             []Timeout             `json:"timeouts"`
	ConsumerRequirements []ConsumerRequirement `json:"consumer_requirements"`
	Assertions           []string              `json:"assertions"`
}

type Sources struct {
	Levels   string `json:"levels"`
	Timeouts string `json:"timeouts"`
}

type ExternalSignal struct {
	Signal string `json:"signal"`
	Action string `json:"action"`
}

type Level struct {
	Name  string `json:"name"`
	Const string `json:"const"`
	Order int    `json:"order"`
}

type Timeout struct {
	Const    string `json:"const"`
	Duration string `json:"duration"`
}

type ConsumerRequirement struct {
	File     string `json:"file"`
	Contains string `json:"contains"`
	Reason   string `json:"reason"`
}
