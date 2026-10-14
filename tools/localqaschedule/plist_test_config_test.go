package main

func testConfig() config {
	repo, s3 := ".", "s3://bucket/daily"
	product, coverage := "/tmp/product.json", "/tmp/coverage.json"
	label, plist, evidence := "io.test", "", "/tmp/schedule.json"
	clientRoot, baseURL, workspace := "/tmp/client", "http://localhost:3000", "W1"
	agentHost, riidoHost, teamID := "https://staging.ai-api.riido.io", "https://staging.api.riido.io", "team-a"
	storage := "/tmp/state.json"
	taskID, first, second, comment := "task-a", "agent-a", "agent-b", "hi"
	hour, minute := 9, 5
	install, inspect, runAtLoad := false, false, false
	runProduct, startClient, mutations := true, true, true
	return config{
		repo:             &repo,
		s3Prefix:         &s3,
		evidenceOut:      &evidence,
		productEvidence:  &product,
		coverageEvidence: &coverage,
		clientRoot:       &clientRoot,
		productBaseURL:   &baseURL,
		productAgentHost: &agentHost,
		productRiidoHost: &riidoHost,
		productWorkspace: &workspace,
		productTeamID:    &teamID,
		productStorage:   &storage,
		productTaskID:    &taskID,
		productAgentID1:  &first,
		productAgentID2:  &second,
		productComment:   &comment,
		taskMutations:    &mutations,
		taskFixture:      &mutations,
		startClient:      &startClient,
		runProduct:       &runProduct,
		label:            &label,
		plistPath:        &plist,
		hour:             &hour,
		minute:           &minute,
		install:          &install,
		inspect:          &inspect,
		runAtLoad:        &runAtLoad,
	}
}
