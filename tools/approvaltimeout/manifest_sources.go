package main

type SemanticActivityManifest struct {
	SemanticActivity []string `json:"semantic_activity"`
}

type ProviderDraftManifest struct {
	MappedEvents  []MappedEvent  `json:"mapped_events"`
	SkippedEvents []SkippedEvent `json:"skipped_events"`
}

type MappedEvent struct {
	EventKind string `json:"event_kind"`
	EventType string `json:"event_type"`
}

type SkippedEvent struct {
	EventKind string `json:"event_kind"`
	Reason    string `json:"reason"`
}
