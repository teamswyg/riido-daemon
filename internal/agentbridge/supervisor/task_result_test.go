package supervisor

import "testing"

func TestSupervisorClaimsTaskAndReportsResult(t *testing.T) {
	run := startTaskResultSupervisor(t)

	expectTaskResultStarted(t, run.reporter, "t-1")
	expectTaskResultRunningEvent(t, run.reporter)
	completeTaskResultProcess(run.running)
	expectTaskResultTextDelta(t, run.reporter, "done")
	expectTaskResultCompletedRun(t, run)
	assertTaskResultRuntimeRegistration(t, run.source)
}
