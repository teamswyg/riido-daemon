package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func printJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  riido serve [--socket PATH] [--transport unix-socket|windows-named-pipe] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido api <status|tasks> [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido api review-demo --channel CHANNEL --review-demo-consent-granted true|false [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido api transition <task-id> --to STATE --event EVENT --approval-id ID [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--actor ACTOR] [--source SOURCE] [--reason TEXT] [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido api evidence <task-id> --command COMMAND --approval-id ID [--exit-code N] [--result RESULT] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--provider-run-id ID] [--provider-run-result RESULT] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido api validate <task-id> --command COMMAND --approval-id ID [--workdir PATH] [--timeout-seconds N] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido mwsd <snapshot|projection|sync|orchestration|projects|status> [--socket PATH] [--state PATH] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido task list [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido task transition <task-id> --to STATE --event EVENT --approval-id ID [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--actor ACTOR] [--source SOURCE] [--reason TEXT] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido task evidence <task-id> --command COMMAND --approval-id ID [--exit-code N] [--result RESULT] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--provider-run-id ID] [--provider-run-result RESULT] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido task validate <task-id> --command COMMAND --approval-id ID [--workdir PATH] [--timeout-seconds N] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido bridge <providers|detect>")
	fmt.Fprintln(os.Stderr, "  riido daemon start [--foreground] [--socket PATH] [--pid-file PATH] [--log-file PATH] [--lock-file PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon status [--socket PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon health [--socket PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon ready [--socket PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon metrics [--socket PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon stop [--socket PATH] [--pid-file PATH] [--timeout-seconds N] [--force]")
	fmt.Fprintln(os.Stderr, "  riido daemon logs --log-file PATH [--lines N]")
}
