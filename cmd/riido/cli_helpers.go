package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func printJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func printUsage() { printCommandUsage(os.Stderr, true) }

func printRootUsage() { printCommandUsage(os.Stderr, false) }

func printCommandUsage(w io.Writer, includeInternal bool) {
	fmt.Fprintln(w, "usage:")
	printContextUsage(w)
	if includeInternal {
		fmt.Fprintln(w, "  # daemon and local runtime (used by Riido Desktop)")
	} else {
		fmt.Fprintln(w, "  # local runtime commands (used by Riido Desktop)")
	}
	fmt.Fprintln(w, "  riido serve [--socket PATH] [--transport unix-socket|windows-named-pipe] [--task-db PATH]")
	fmt.Fprintln(w, "  riido api <status|tasks> [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(w, "  riido api review-demo --channel CHANNEL --review-demo-consent-granted true|false [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(w, "  riido api transition <task-id> --to STATE --event EVENT --approval-id ID [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--actor ACTOR] [--source SOURCE] [--reason TEXT] [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(w, "  riido api evidence <task-id> --command COMMAND --approval-id ID [--exit-code N] [--result RESULT] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--provider-run-id ID] [--provider-run-result RESULT] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(w, "  riido api validate <task-id> --command COMMAND --approval-id ID [--workdir PATH] [--timeout-seconds N] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	if includeInternal {
		fmt.Fprintln(w, "  riido mwsd <snapshot|projection|sync|orchestration|projects|status> [--socket PATH] [--state PATH] [--task-db PATH]")
	}
	fmt.Fprintln(w, "  riido task list [--task-db PATH]")
	fmt.Fprintln(w, "  riido task transition <task-id> --to STATE --event EVENT --approval-id ID [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--actor ACTOR] [--source SOURCE] [--reason TEXT] [--task-db PATH]")
	fmt.Fprintln(w, "  riido task evidence <task-id> --command COMMAND --approval-id ID [--exit-code N] [--result RESULT] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--provider-run-id ID] [--provider-run-result RESULT] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--task-db PATH]")
	fmt.Fprintln(w, "  riido task validate <task-id> --command COMMAND --approval-id ID [--workdir PATH] [--timeout-seconds N] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--task-db PATH]")
	if includeInternal {
		fmt.Fprintln(w, "  riido bridge <providers|detect>")
	}
	fmt.Fprintln(w, "  riido version")
	if includeInternal {
		fmt.Fprintln(w, "  riido daemon start [--foreground] [--socket PATH] [--pid-file PATH] [--log-file PATH] [--lock-file PATH]")

		fmt.Fprintln(w, "  riido daemon status [--socket PATH]")

		fmt.Fprintln(w, "  riido daemon health [--socket PATH]")

		fmt.Fprintln(w, "  riido daemon ready [--socket PATH]")

		fmt.Fprintln(w, "  riido daemon metrics [--socket PATH]")

		fmt.Fprintln(w, "  riido daemon stop [--socket PATH] [--pid-file PATH] [--timeout-seconds N] [--force]")

		fmt.Fprintln(w, "  riido daemon logs --log-file PATH [--lines N]")
	}
}
