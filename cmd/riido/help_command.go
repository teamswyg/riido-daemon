package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// runFriendlyCommandHelp answers help requests for user commands with the
// curated help when stdout is a person's terminal (never in agent runs or
// pipes). handled is false when the command should run as before.
func runFriendlyCommandHelp(args []string) (handled bool, err error) {
	topic, ok := commandHelpCatalog[args[0]]
	if !ok {
		return false, nil
	}
	term := humanTerminal(os.Stdout)
	if !term.Interactive {
		return false, nil
	}
	rest := args[1:]
	wantsHelp := slices.ContainsFunc(rest, isHelpArg)
	var sub *subHelp
	first := ""
	if len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
		first = rest[0]
		sub = topic.sub(first)
	}
	switch {
	case wantsHelp && sub != nil:
		printSubcommandHelp(os.Stdout, term, topic, *sub)
	case wantsHelp && first == "":
		printCommandHelp(os.Stdout, term, topic)
	case len(topic.subs) > 0 && len(rest) == 0:
		printCommandHelp(os.Stdout, term, topic)
	case len(topic.subs) > 0 && first != "" && sub == nil:
		names := make([]string, 0, len(topic.subs))
		for _, s := range topic.subs {
			names = append(names, s.name)
		}
		suggestion := suggestCommand(first, names)
		if suggestion == "" && topic.sub("get") != nil && looksLikeWorkKey(first) {
			// `riido work RIID-1234` most likely meant `riido work get RIID-1234`.
			suggestion = "get " + cliui.Safe(first)
		}
		return true, reportUnknownCommand(first, suggestion, "riido "+topic.name+" ", unknownSubcommandExit())
	default:
		return false, nil
	}
	return true, nil
}

func unknownSubcommandExit() contextcli.ExitCode {
	return contextcli.Fail(contextcli.CodeUsage, "", "").ExitCode()
}

// friendlyUnknownCommand replaces the full legacy usage dump with a short
// error and a "did you mean" suggestion on a person's terminal.
func friendlyUnknownCommand(name string) (bool, error) {
	if !humanTerminal(os.Stderr).Interactive {
		return false, nil
	}
	names := []string{"help"}
	for _, group := range friendlyHelpGroups {
		for _, entry := range group.entries {
			names = append(names, entry.name)
		}
	}
	return true, reportUnknownCommand(name, suggestCommand(name, names), "riido ", 1)
}

// looksLikeWorkKey reports a display key such as RIID-1234.
func looksLikeWorkKey(s string) bool {
	prefix, number, ok := strings.Cut(s, "-")
	return ok && prefix != "" && number != "" && strings.Trim(number, "0123456789") == ""
}

func reportUnknownCommand(name, suggestion, prefix string, exit contextcli.ExitCode) error {
	term := humanTerminal(os.Stderr)
	fmt.Fprintf(os.Stderr, "%s 알 수 없는 명령이에요: %s\n", term.Symbol("fail"), prefix+cliui.Safe(name))
	if suggestion != "" {
		fmt.Fprintf(os.Stderr, "  %s 혹시 %s 인가요?\n", term.Accent("→"), term.Bold(prefix+suggestion))
	}
	fmt.Fprintf(os.Stderr, "  %s\n", term.Dim("명령 목록: "+strings.TrimSpace(prefix)+" --help"))
	return reportedExit{code: exit}
}

// suggestCommand returns the closest candidate within two edits, or a
// candidate the input is a prefix of (like `riido wor`).
func suggestCommand(input string, candidates []string) string {
	best, bestDistance := "", 3
	for _, candidate := range candidates {
		if len(input) >= 2 && strings.HasPrefix(candidate, input) {
			return candidate
		}
		if d := editDistance(input, candidate); d < bestDistance {
			best, bestDistance = candidate, d
		}
	}
	return best
}

func editDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur := make([]int, len(rb)+1)
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(rb)]
}

// printCommandHelp follows the GitHub CLI help layout: summary, usage,
// commands, flags, examples, learn more.
func printCommandHelp(w io.Writer, term cliui.Terminal, topic commandHelp) {
	fmt.Fprintln(w, topic.summary)
	printNotes(w, term, topic.notes)
	printSection(w, term, "사용법", "  "+topic.usage)
	if len(topic.subs) > 0 {
		entries := make([]helpEntry, 0, len(topic.subs))
		for _, sub := range topic.subs {
			entries = append(entries, helpEntry{sub.name, sub.summary})
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, term.Bold("명령"))
		printEntries(w, term, entries, true, 0)
	}
	printFlags(w, term, topic.flags)
	printExamples(w, term, topic.examples)
	more := "riido " + topic.name + " <명령> --help · 전체 옵션 riido help --all"
	if len(topic.subs) == 0 {
		more = "전체 옵션 riido help --all"
	}
	printSection(w, term, "더 알아보기", "  "+more)
}

func printSubcommandHelp(w io.Writer, term cliui.Terminal, topic commandHelp, sub subHelp) {
	fmt.Fprintln(w, sub.summary)
	// The work-key note only matters where the usage takes a work key.
	notes := slices.DeleteFunc(slices.Clone(topic.notes), func(note string) bool {
		return note == keyNote && !strings.Contains(sub.usage, "RIID-1234")
	})
	printNotes(w, term, notes)
	printSection(w, term, "사용법", "  "+sub.usage+" [옵션]")
	printFlags(w, term, append(slices.Clone(sub.flags), topic.flags...))
	if sub.example != "" {
		printExamples(w, term, []string{sub.example})
	}
	printSection(w, term, "더 알아보기", "  riido "+topic.name+" --help · 전체 옵션 riido help --all")
}

func printNotes(w io.Writer, term cliui.Terminal, notes []string) {
	for _, note := range notes {
		fmt.Fprintln(w, term.Dim(note))
	}
}

func printSection(w io.Writer, term cliui.Terminal, title, body string) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, term.Bold(title))
	fmt.Fprintln(w, body)
}

func printFlags(w io.Writer, term cliui.Terminal, flags []helpEntry) {
	if len(flags) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, term.Bold("옵션"))
	printEntries(w, term, flags, false, 0)
}
