package main

import "testing"

func TestBrowserMeaningScenarioProvesSearchSurface(t *testing.T) {
	got := browserMeaningScenario()
	if got.Status != statusPassed {
		t.Fatalf("browser meaning proof failed: %+v", got)
	}
	observed := got.Observed
	if observed["replaces_inferred_id"] != "browser-meaning-qa" || observed["human_browser_needed"] != false {
		t.Fatalf("unexpected observed proof: %+v", observed)
	}
}
