package main

func agentIdentityContractScenario(fixture agentFixture, pair [2]preparedRuntime) scenario {
	sc := scenario{ID: "contract.task.frontend_identity_contract", Status: statusPassed}
	sc.Observed = map[string]any{
		"same_runtime_kind_pair": len(fixture.Candidates) == 2 &&
			fixture.Candidates[0].RuntimeKind == fixture.Candidates[1].RuntimeKind,
		"runtime_ids_distinct": pair[0].RuntimeID != pair[1].RuntimeID,
		"agent_ids_distinct": len(fixture.Candidates) == 2 &&
			fixture.Candidates[0].AgentID != fixture.Candidates[1].AgentID,
		"dedupe_key":       "thread_id",
		"forbidden_dedupe": "provider,runtime_kind,task_id",
	}
	if len(fixture.Candidates) != 2 {
		sc.Status = statusFailed
		sc.FailureSummary = "two created QA agents are required for identity contract"
	}
	return sc
}

func agentFixtureSkippedScenario() scenario {
	return scenario{
		ID:             "local.saas.agent_fixture.runtime_pair",
		Status:         statusSkipped,
		FailureSummary: "Need two online prepared runtimes with the same provider kind.",
		Repair:         apiRuntimeRepair(),
	}
}
