package main

import "github.com/teamswyg/riido-daemon/internal/scheduling"

func requireMissingReason(name string, got scheduling.Eligibility) []problem {
	if got.Eligible || firstReasonCode(got) != "MISSING_REQUIRED_SURFACE" {
		return []problem{{Message: name + " did not produce missing required surface"}}
	}
	return nil
}

func requireEligible(name string, got scheduling.Eligibility) []problem {
	if !got.Eligible {
		return []problem{{Message: name + " was not eligible when candidate flag was true"}}
	}
	return nil
}

func validateUnknownSurfaceFailsClosed() []problem {
	got := scheduling.EvaluateCapability(require("future_surface"), candidateBase())
	if got.Eligible || firstReasonCode(got) != "UNKNOWN_REQUIRED_SURFACE" {
		return []problem{{Message: "unknown required surface did not fail closed"}}
	}
	return nil
}

func firstReasonCode(got scheduling.Eligibility) string {
	if len(got.Reasons) == 0 {
		return ""
	}
	return got.Reasons[0].Code
}
