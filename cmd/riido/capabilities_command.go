package main

import (
	"fmt"
	"io"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func runCapabilities(c contextIO, raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		fmt.Fprintln(c.stderr, "  riido capabilities [--json]")
		return nil
	}
	data := contextcli.Capabilities(c.cliVersion())
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, data), renderCapabilities)
}

func renderCapabilities(w io.Writer, data any) {
	caps, ok := data.(contextcli.CapabilitiesData)
	if !ok {
		return
	}
	fmt.Fprintf(w, "%s · envelope v%s · %s v%s\n", caps.CLIVersion, caps.EnvelopeSchemaVersion, caps.ContextIPCMethod, caps.ContextIPCSchema)
	for _, cmd := range caps.Commands {
		line := fmt.Sprintf("  %-18s %-22s %v", cmd.Name, cmd.Status, cmd.Modes)
		if cmd.Note != "" {
			line += "  " + cmd.Note
		}
		fmt.Fprintln(w, line)
	}
}
