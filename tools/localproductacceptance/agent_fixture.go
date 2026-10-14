package main

func maybeCreateAgentFixtures(
	cfg config,
	client apiClient,
	base string,
	prep saasPrepareResult,
) agentFixture {
	if !*cfg.prepareDaemon || !*cfg.runMutations {
		return agentFixture{}
	}
	pair, ok := choosePreparedRuntimePair(prep.Runtimes)
	if !ok {
		return agentFixture{Scenarios: []scenario{agentFixtureSkippedScenario()}}
	}
	fixture := createAgentFixturePair(client, base, pair)
	fixture.Scenarios = append(fixture.Scenarios, agentIdentityContractScenario(fixture, pair))
	return fixture
}

func createAgentFixturePair(client apiClient, base string, pair [2]preparedRuntime) agentFixture {
	var fixture agentFixture
	for idx, runtime := range pair {
		sc, candidate := createAgentFixture(client, base, idx+1, runtime)
		fixture.Scenarios = append(fixture.Scenarios, sc)
		if sc.Status == statusPassed {
			fixture.Candidates = append(fixture.Candidates, candidate)
			fixture.CreatedIDs = append(fixture.CreatedIDs, candidate.AgentID)
		}
	}
	return fixture
}
