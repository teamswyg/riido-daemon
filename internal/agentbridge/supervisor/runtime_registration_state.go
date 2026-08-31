package supervisor

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func runtimeRegistrationState(status runtimeactor.Status) string {
	var out strings.Builder
	for _, capability := range status.Capabilities {
		health, code, _ := providerHealthObservation(capability)
		out.WriteString(capability.Provider)
		out.WriteByte('|')
		out.WriteString(string(health))
		out.WriteByte('|')
		out.WriteString(string(code))
		out.WriteByte('|')
		out.WriteString(capability.Version)
		out.WriteByte(';')
	}
	return out.String()
}
