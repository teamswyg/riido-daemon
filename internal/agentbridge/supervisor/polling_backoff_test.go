package supervisor

import "testing"

func TestSupervisorDefaultMailboxMatchesProviderRuntimeBackpressureSSOT(t *testing.T) {
	actor := newDefaultMailboxSupervisor(t)

	assertDefaultMailboxSize(t, actor)
}

func TestSupervisorBacksOffPollingWhenIdle(t *testing.T) {
	run := startIdlePollingSupervisor(t)

	expectIdlePoll(t, run.source)
	assertNoIdlePollBeforeBackoff(t, run.source)
	expectIdlePollResumes(t, run.source)
}
