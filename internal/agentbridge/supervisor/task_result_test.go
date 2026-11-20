package supervisor

import "testing"

func TestSupervisorClaimsTaskAndReportsResult(t *testing.T) {
	run := startTaskResultSupervisor(t)

	expectTaskResultStarted(t, run.reporter, "t-1")
	expectTaskResultRunningEvent(t, run.reporter)
	completeTaskResultProcess(run.running)
	// The authoritative final result may supersede an unsent partial display.
	expectTaskResultCompletedRun(t, run)
	assertTaskResultRuntimeRegistration(t, run.source)
}

func TestFirstAnswerDisplayDoesNotWaitForCompletion(t *testing.T) {
	run := startTaskResultSupervisor(t)
	expectTaskResultStarted(t, run.reporter, "t-1")
	expectTaskResultRunningEvent(t, run.reporter)
	run.running.EmitStdout([]byte("event"))
	expectTaskResultTextDelta(t, run.reporter, "event")
	completeTaskResultProcess(run.running)
	expectTaskResultCompleted(t, run.reporter)
}
