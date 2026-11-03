package main

import (
	"slices"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/clicompletion"
)

var completionShells = []string{"bash", "zsh", "fish", "powershell"}

// This catalog describes supported human-terminal grammar, not runtime access.
// Keep opaque values nil: completion must never enumerate files or remote data.
func completionCatalog() []clicompletion.Command {
	var result []clicompletion.Command
	user := func(path, values string, bools ...string) {
		cmd := clicompletion.Command{Path: path, Flags: append([]string{"--help", "-h", "--json", "--no-input", "--no-color"}, bools...), Values: map[string][]string{"--profile": {"local", "staging", "production"}, "--auth-origin": {"browser", "desktop"}, "--format": {"text", "json"}, "--timeout": nil}}
		for name := range strings.FieldsSeq(values) {
			cmd.Values[name] = nil
		}
		result = append(result, cmd)
	}
	for _, sub := range []string{"list", "get", "set"} {
		user(string(mainCommandConfig)+" "+sub, "")
		if sub != "list" {
			result[len(result)-1].Positionals = func(previous []string) []string {
				if len(previous) == 0 {
					return configKeys
				}
				if sub == "set" && len(previous) == 1 {
					switch previous[0] {
					case "profile":
						return []string{"local", "staging", "production"}
					case "auth_origin":
						return []string{"browser", "desktop"}
					case "format":
						return []string{"json", "text"}
					}
				}
				return nil
			}
		}
	}
	user(string(mainCommandWorkspace)+" list", "--limit --cursor")
	user(string(mainCommandWorkspace)+" use", "")
	for _, sub := range []workCommand{workCommandGet, workCommandTree, workCommandRelations} {
		user(string(mainCommandWork)+" "+string(sub), "--team-id --team --workspace")
	}
	user(string(mainCommandWork)+" "+string(workCommandComments), "--team-id --team --workspace --page --limit --kind --version")
	user(string(mainCommandWork)+" "+string(workCommandReplies), "--team-id --team --workspace --comment-id --page --limit --kind --version")
	user(string(mainCommandWork)+" "+string(workCommandActivity), "--team-id --team --workspace --page --limit")
	user(string(mainCommandWork)+" "+string(workCommandSearch), "--workspace --query --limit --cursor")
	body := "--team-id --team --part --version --offset --limit --comment-id --page --comments-limit"
	user(string(mainCommandWork)+" "+string(workCommandBody), body+" --kind --workspace")
	user(string(mainCommandMeeting)+" body", body+" --board-cursor --board-limit")
	user(string(mainCommandMeeting)+" get", "--team-id --team --cursor --limit")
	user(string(mainCommandMeeting)+" list", "--work --team-id --team --workspace")
	user(string(mainCommandMeeting)+" comments", "--team-id --team --comment-id --page --limit --version")
	user(string(mainCommandContext)+" get", "--work --team-id --team --workspace")
	for _, root := range []mainCommand{mainCommandResource, mainCommandRepo} {
		user(string(root)+" list", "--work --team-id --team --workspace")
	}
	for _, sub := range []string{"list", "show"} {
		user(string(mainCommandThread)+" "+sub, "--work --workspace --team-id --team")
	}
	user(string(mainCommandAgents)+" list", "--work --workspace --team-id --team --cursor")
	user(string(mainCommandAgents)+" show", "--work --workspace --team-id --team")
	user(string(mainCommandAuth)+" login", "", "--browser")
	delete(result[len(result)-1].Values, "--auth-origin")
	user(string(mainCommandAuth)+" repair", "", "--restart")
	for _, sub := range []string{"link", "status", "logout"} {
		user(string(mainCommandAuth)+" "+sub, "")
	}
	user(string(mainCommandWhoami), "")
	user(string(mainCommandCapabilities), "")
	user(string(mainCommandDoctor), "", "--check-api")
	// Generator options are deliberately independent of user settings.
	result = append(result, clicompletion.Command{Path: string(mainCommandCompletion), Flags: []string{"--help", "-h", "--json", "--no-input", "--no-color"}, Values: map[string][]string{"--format": {"text", "json"}}, Positionals: func(previous []string) []string {
		if len(previous) == 0 {
			return completionShells
		}
		return nil
	}})
	result = append(result, clicompletion.Command{Path: string(mainCommandVersion)})
	legacy := func(path, values string) {
		cmd := clicompletion.Command{Path: path, Flags: []string{"--help", "-h"}, Values: map[string][]string{}}
		for name := range strings.FieldsSeq(values) {
			cmd.Values[name] = nil
		}
		result = append(result, cmd)
	}
	legacy(string(mainCommandServe), "--socket --transport --task-db")
	for _, sub := range []apiCommand{apiCommandStatus, apiCommandTasks} {
		legacy(string(mainCommandAPI)+" "+string(sub), "--socket --transport")
	}
	legacy(string(mainCommandAPI)+" "+string(apiCommandReviewDemo), "--channel --review-demo-consent-granted --socket --transport")
	legacy(string(mainCommandTask)+" "+string(taskCommandList), "--task-db")
	for _, root := range []mainCommand{mainCommandAPI, mainCommandTask} {
		transport := "--task-db"
		if root == mainCommandAPI {
			transport = "--socket --transport"
		}
		base := transport + " --approval-id --provider --decision-llm --command-id --actor --source"
		legacy(string(root)+" transition", base+" --to --event --reason")
		legacy(string(root)+" evidence", base+" --command --exit-code --result --summary --validation-gate --provider-run-id --provider-run-result")
		legacy(string(root)+" validate", base+" --command --workdir --timeout-seconds --validation-gate")
	}
	slices.SortFunc(result, func(a, b clicompletion.Command) int { return strings.Compare(a.Path, b.Path) })
	return result
}
