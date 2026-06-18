package mwsdbridge

// Snapshot is Riido's initial projection from macmini-workspace.
type Snapshot struct {
	Status        Status                `json:"status"`
	Graph         GraphExport           `json:"graph"`
	Domain        DomainExport          `json:"domain"`
	Harness       HarnessIndex          `json:"harness"`
	Orchestration OrchestrationSnapshot `json:"orchestration"`
	Projects      ProjectRegistry       `json:"projects"`
}
