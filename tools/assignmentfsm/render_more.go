package main

import (
	"fmt"
	"strings"
)

func renderMetadata(b *strings.Builder, fsm FSMSnapshot) {
	fmt.Fprintln(b, "## FSM Metadata")
	fmt.Fprintln(b)
	fmt.Fprintf(b, "- name: `%s`\n", fsm.Name)
	fmt.Fprintf(b, "- type union: `%s`\n", fsm.TypeUnion)
	fmt.Fprintf(b, "- states: `%s`\n", strings.Join(fsm.States, "`, `"))
	fmt.Fprintf(b, "- start states: `%s`\n", strings.Join(fsm.StartStates, "`, `"))
	fmt.Fprintf(b, "- terminal states: `%s`\n", strings.Join(fsm.TerminalStates, "`, `"))
	fmt.Fprintf(b, "- user-visible active states: `%s`\n", strings.Join(fsm.AgentActive, "`, `"))
	fmt.Fprintf(b, "- transition count: `%d`\n\n", len(fsm.Transitions))
}

func renderRelated(b *strings.Builder, manifest Manifest) {
	fmt.Fprintln(b, "## Related Sections")
	fmt.Fprintln(b)
	for _, section := range manifest.RelatedSections {
		fmt.Fprintf(b, "- [%s](%s)\n", section.Title, section.Path)
	}
}
