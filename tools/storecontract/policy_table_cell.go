package main

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func evaluatePolicyTableCell(
	channel hostintegration.DistributionChannel,
	surface policySurfaceSpec,
) policyTableCell {
	for _, scenario := range policyFactScenarios {
		decision := policy.EvaluateStoreChannelPolicy(policyInput(channel, surface, scenario))
		if decision.Allowed {
			return allowedPolicyCell(channel, decision, scenario)
		}
	}
	decision := policy.EvaluateStoreChannelPolicy(policyInput(channel, surface, policyFactScenario{}))
	return policyTableCell{
		Channel:  string(channel),
		Decision: deniedPolicyDecision(decision.Code),
		Code:     decision.Code,
	}
}

func allowedPolicyCell(
	channel hostintegration.DistributionChannel,
	decision policy.Decision,
	scenario policyFactScenario,
) policyTableCell {
	if len(scenario.Facts) == 0 {
		return policyTableCell{Channel: string(channel), Decision: "allowed", Code: decision.Code}
	}
	return policyTableCell{
		Channel:       string(channel),
		Decision:      "requires " + strings.Join(scenario.Facts, " + "),
		Code:          decision.Code,
		RequiredFacts: scenario.Facts,
	}
}

func deniedPolicyDecision(code string) string {
	switch code {
	case "STORE_CHANNEL_SURFACE_FORBIDDEN":
		return "forbidden"
	case "STORE_CHANNEL_SURFACE_NOT_APPLICABLE":
		return "not applicable"
	case "STORE_CHANNEL_SURFACE_DISCOURAGED":
		return "discouraged"
	default:
		return "denied: " + code
	}
}
