package main

import (
	"fmt"
	"io"
	"slices"

	"github.com/teamswyg/riido-daemon/internal/clicompletion"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func runCompletion(c contextIO, raw []string) error {
	if len(raw) == 1 && raw[0] == "__complete" {
		words := clicompletion.ReadWords(c.stdin)
		if slices.ContainsFunc(words, contextcli.IsSecretArg) {
			return nil
		}
		for _, candidate := range clicompletion.Candidates(completionCatalog(), words) {
			if _, err := fmt.Fprintln(c.stdout, candidate); err != nil {
				return err
			}
		}
		return nil
	}
	id := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{MaxPositionals: 1})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), id, contextcli.Fail(cliErr.Code, "completion 입력이 올바르지 않습니다", "riido completion --help"))
	}
	if args.Help {
		printCompletionUsage(c.stderr)
		return nil
	}
	if len(args.Positionals) != 1 || args.HasValue("--workspace") || args.HasValue("--cursor") || args.HasValue("--timeout") {
		return c.fail(args.JSON, id, contextcli.Fail(contextcli.CodeUsage, "자동완성을 받을 셸 하나를 지정하세요", "riido completion --help"))
	}
	script, ok := clicompletion.Script(args.Positionals[0])
	if !ok {
		return c.fail(args.JSON, id, contextcli.Fail(contextcli.CodeUsage, "지원하는 셸은 bash, zsh, fish, powershell입니다", "riido completion --help"))
	}
	if args.JSON {
		return c.succeed(true, contextcli.NewEnvelope(id, map[string]any{"shell": args.Positionals[0], "script": script}), nil)
	}
	_, err := io.WriteString(c.stdout, script)
	return err
}

func printCompletionUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido completion <bash|zsh|fish|powershell> [--json]")
	fmt.Fprintln(w, "    Static user-terminal completion. Save the output and source it in the matching shell (zsh: after compinit). No config, credentials, API calls or automatic startup-file changes.")
	fmt.Fprintln(w, "    PowerShell keeps native filename fallback when there are no static matches.")
}
