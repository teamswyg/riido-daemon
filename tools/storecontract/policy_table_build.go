package main

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func buildPolicyTable(channels []channel) ([]policyTableRow, []string) {
	var rows []policyTableRow
	var problems []string
	for _, surface := range policyTableSurfaces {
		row := policyTableRow{Surface: surface.Label}
		for _, item := range channels {
			channelID := hostintegration.DistributionChannel(item.ID)
			if !channelID.Valid() {
				problems = append(problems, fmt.Sprintf("policy table channel %q is unknown", item.ID))
				continue
			}
			row.Cells = append(row.Cells, evaluatePolicyTableCell(channelID, surface))
		}
		rows = append(rows, row)
	}
	return rows, problems
}

func policyInput(
	channel hostintegration.DistributionChannel,
	surface policySurfaceSpec,
	scenario policyFactScenario,
) policy.StoreChannelPolicyInput {
	return policy.StoreChannelPolicyInput{
		Channel:                channel,
		Surface:                policy.StoreSurface(surface.ID),
		ExplicitConsentGranted: scenario.Consent,
		OSGrantPresent:         scenario.OSGrant,
		StoreReviewApproved:    scenario.StoreReview,
	}
}
