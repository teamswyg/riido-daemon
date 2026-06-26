package main

import "testing"

func TestBuildQAScheduleRequiresRunAndDashboardEvidence(t *testing.T) {
	m := manifest{LocalQARunEvidence: ".riido-local/evidence/local-qa-run.json"}
	source := qaScheduleSource{
		ID:              "daily-evidence-sweep",
		Cadence:         "daily",
		Entrypoint:      "go run ./tools/localqarunner",
		FreshnessWindow: "24h",
		Evidence: []string{
			".riido-local/evidence/local-qa-run.json",
			".riido-local/evidence/local-qa-schedule.json",
			".riido-local/dashboard/index.html",
		},
	}
	got := buildQASchedule(m, source)
	if got.Status != statusPassed {
		t.Fatalf("qa schedule = %+v", got)
	}
	source.Evidence = source.Evidence[:1]
	got = buildQASchedule(m, source)
	if got.Status != statusPartial || got.PartialReason == "" {
		t.Fatalf("qa schedule should be partial: %+v", got)
	}
}
