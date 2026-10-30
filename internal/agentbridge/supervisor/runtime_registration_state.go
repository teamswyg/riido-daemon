package supervisor

import (
	"fmt"
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
		out.WriteString(capability.VersionObservedAt.String())
		out.WriteByte(';')
	}
	for _, model := range status.Models {
		fmt.Fprintf(&out, "%q:%q:%t;", model.ModelID, model.Label, model.IsDefault)
	}
	return out.String()
}
