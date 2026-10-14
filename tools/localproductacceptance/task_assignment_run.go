package main

import "strconv"

type taskAssignmentRun struct {
	Scenarios []scenario
	First     scenario
	Second    scenario
	Pair      taskAgentPair
	OK        bool
}

func createAssignmentRun(client apiClient, base string, plan taskMutationPlan) taskAssignmentRun {
	var run taskAssignmentRun
	for idx, agent := range plan.Candidates {
		sc := createAssignment(client, candidateAssignmentID(idx), base, plan.TaskID, agent.AgentID)
		annotateAssignmentCandidate(&sc, agent)
		if sc.Status == statusPassed {
			assignSuccessfulCandidate(&run, agent, sc)
		} else {
			run.Scenarios = append(run.Scenarios, sc)
		}
		if run.OK {
			return run
		}
	}
	return run
}

func assignSuccessfulCandidate(run *taskAssignmentRun, agent taskAgentCandidate, sc scenario) {
	if run.First.ID == "" {
		sc.ID = "contract.task.assignment.create.first"
		run.First = sc
		run.Pair.First = agent
	} else {
		sc.ID = "contract.task.assignment.create.second"
		run.Second = sc
		run.Pair.Second = agent
		run.OK = true
	}
	run.Scenarios = append(run.Scenarios, sc)
}

func candidateAssignmentID(idx int) string {
	return "contract.task.assignment.create.candidate." + strconv.Itoa(idx+1)
}

func annotateAssignmentCandidate(sc *scenario, agent taskAgentCandidate) {
	if sc.Observed == nil {
		sc.Observed = map[string]any{}
	}
	sc.Observed["candidate_agent_id"] = agent.AgentID
	sc.Observed["candidate_runtime_kind"] = agent.RuntimeKind
	sc.Observed["candidate_runtime_id"] = agent.RuntimeID
}
