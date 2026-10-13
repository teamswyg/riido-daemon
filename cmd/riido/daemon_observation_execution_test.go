package main

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func TestDaemonObservationRetainsSupervisorOwnedExecutions(t *testing.T) {
	obs := daemonObservation{}
	obs.observeRuntime(runtimeactor.Status{RuntimeID: "rt", RunningSessions: 1, RunningTasks: []runtimeactor.TaskStatus{{TaskID: "native"}}})
	obs.retainOwnedExecutions(map[string][]string{"rt": {"native", "verification", "pending-report"}})
	if obs.metrics.RunningTasks != 3 {
		t.Fatalf("host update would see false idle/double count: %+v", obs.metrics)
	}
}
