package main

func scheduleCandidate(schedule qaScheduleEvidence) closedLoopCandiate {
	return closedLoopCandiate{
		ID:     "product-loop-local-qa-schedule-closure",
		Class:  "product-acceptance",
		Reason: schedule.PartialReason,
		RequiredNextArtifacts: []string{
			"daily QA schedule evidence",
			"local QA run evidence",
			"dashboard handoff evidence",
		},
		Graph: scheduleCandidateGraph(),
	}
}

func scheduleCandidateGraph() candidateGraph {
	return candidateGraph{
		Observation: "local QA can expire without proving an automatic rerun path",
		Hypothesis:  "daily QA trigger evidence should bind run output and dashboard handoff",
		Change:      "productloopevidence reads local-qa-daily-trigger.dsl.json",
		Verifier:    "TestBuildQAScheduleRequiresRunAndDashboardEvidence",
		Evidence:    "product-loop-evidence.qa_schedule",
		Decision:    "missing schedule closure remains partial product-loop debt",
		NextLoop:    "local-qa-evidence-expiry-gate",
	}
}
