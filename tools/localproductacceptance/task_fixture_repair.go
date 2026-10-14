package main

func taskFixtureRepair(summary string) *repair {
	return &repair{
		Class:   "task_fixture_create_failed",
		Owner:   "riido-api-server/local-qa",
		Mode:    "manual",
		Summary: summary,
	}
}
