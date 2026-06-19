package main

type toolLimitation struct {
	SourceID                   string   `json:"source_id"`
	SourceOwner                string   `json:"source_owner"`
	SourceStabilizedBy         []string `json:"source_stabilized_by"`
	LocalRiidoTask             string   `json:"local_riido_task"`
	DaemonScope                string   `json:"daemon_scope"`
	RequiredAuthoritativePages []string `json:"required_authoritative_pages"`
	MustPreserveNonUINodes     []string `json:"must_preserve_non_ui_nodes"`
}

type figmaRef struct {
	FileKey     string `json:"file_key"`
	FileName    string `json:"file_name"`
	PageID      string `json:"page_id"`
	PageName    string `json:"page_name"`
	InspectedAt string `json:"inspected_at"`
}

type boundaryPolicy struct {
	Summary  string `json:"summary"`
	TopDown  string `json:"top_down"`
	BottomUp string `json:"bottom_up"`
}

type boundaryEntry struct {
	NodeID              string   `json:"node_id"`
	Name                string   `json:"name"`
	UpstreamOwner       []string `json:"upstream_owner"`
	DaemonScope         string   `json:"daemon_scope"`
	DaemonConsumedFacts []string `json:"daemon_consumed_facts"`
	ClientOwnedFacts    []string `json:"client_owned_facts"`
}
