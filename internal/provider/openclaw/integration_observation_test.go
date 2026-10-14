package openclaw

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

type openClawIntegrationObservation struct {
	result agentbridge.Result
	events []agentbridge.Event
}
