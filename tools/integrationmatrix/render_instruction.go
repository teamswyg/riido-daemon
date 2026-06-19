package main

import (
	"fmt"
	"strings"
)

func renderInstructionEffectiveness(m manifest) string {
	return renderDoc("Agent Instruction Effectiveness Probe", "../integration-matrix.md", func(b *strings.Builder) {
		b.WriteString("Provider instruction effectiveness is separate from deterministic prompt placement.\n\n")
		fmt.Fprintf(b, "Public CI verifies placement with `%s`, without launching provider CLIs.\n\n", m.InstructionProbe.PublicCI)
		b.WriteString("Real provider evidence is opt-in and must use the same harness:\n\n")
		fmt.Fprintf(b, "1. Build the provider-specific probe with `%s`.\n", m.InstructionProbe.Builder)
		b.WriteString("2. Send the generated prompt/system prompt through the provider adapter's normal integration path.\n")
		fmt.Fprintf(b, "3. Validate the provider output with `%s`.\n", m.InstructionProbe.Validator)
		b.WriteString("4. Record missing executable or missing authentication as a skip only before the provider roundtrip starts.\n\n")
		b.WriteString("| Provider | Probe marker | Expected instruction surface |\n| --- | --- | --- |\n")
		for _, p := range m.InstructionProbe.Providers {
			fmt.Fprintf(b, "| %s | `%s` | `%s` |\n", displayName(m, p.Provider), p.Marker, p.Surface)
		}
	})
}

func displayName(m manifest, id string) string {
	for _, p := range m.ProviderValidation.Providers {
		if p.Provider == id {
			return p.DisplayName
		}
	}
	return id
}
