package mwsdbridge

func fakeSnapshotResponses() map[string]string {
	return map[string]string{
		"status":        fakeStatusResponse(),
		"graph":         fakeGraphResponse(),
		"domain":        fakeDomainResponse(),
		"harness":       fakeHarnessResponse(),
		"orchestration": fakeOrchestrationResponse(),
		"projects":      fakeProjectsResponse(),
	}
}
