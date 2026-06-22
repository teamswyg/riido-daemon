package main

func testConfig() config {
	repo, s3 := ".", "s3://bucket/daily"
	product, label, plist, evidence := "/tmp/product.json", "io.test", "", "/tmp/schedule.json"
	clientRoot, baseURL, workspace := "/tmp/client", "http://localhost:3000", "W1"
	riidoHost, teamID := "https://development.api.riido.io", "team-a"
	storage := "/tmp/state.json"
	taskID, first, second, comment := "task-a", "agent-a", "agent-b", "hi"
	hour, minute := 9, 5
	install, runAtLoad, runProduct, startClient, mutations := false, false, true, true, true
	return config{
		repo:             &repo,
		s3Prefix:         &s3,
		evidenceOut:      &evidence,
		productEvidence:  &product,
		clientRoot:       &clientRoot,
		productBaseURL:   &baseURL,
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
		runAtLoad:        &runAtLoad,
	}
}
