package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// homeServer names the API the next user command would use, or "" when the
// configuration cannot be resolved (doctor reports the reason).
func homeServer(c contextIO) string {
	args, settings, cliErr := c.parseSettings(nil, contextcli.ArgSpec{})
	if cliErr != nil || settings == nil {
		return ""
	}
	origin, _, err := c.userAPIOrigin(args)
	if err != nil {
		return ""
	}
	return origin + " (" + settings.resolved.Profile + ")"
}

type helpEntry struct{ name, summary string }

type helpGroup struct {
	title   string
	entries []helpEntry
}

var friendlyHelpGroups = []helpGroup{
	{"시작하기", []helpEntry{
		{"auth", "로그인·로그아웃·상태 확인 (login, status, logout, link, repair)"},
		{"whoami", "지금 로그인한 계정"},
		{"doctor", "설치·연결·로그인 진단"},
	}},
	{"조회", []helpEntry{
		{"work", "작업 (get, tree, relations, comments, replies, activity, search, body)"},
		{"context", "작업 맥락 묶음 (get, part, changes)"},
		{"meeting", "미팅 (list, get, comments, body)"},
		{"agents", "Agent 목록·상세 (list, show)"},
		{"thread", "Agent 대화 (list, show)"},
		{"resource", "작업에 연결된 링크·첨부·Figma"},
		{"repo", "작업에 연결된 저장소·PR"},
		{"workspace", "워크스페이스 (list, use)"},
	}},
	{"설정", []helpEntry{
		{"config", "기본 워크스페이스·팀·프로필 (list, get, set)"},
		{"completion", "셸 자동완성 스크립트 (bash, zsh, fish, powershell)"},
		{"capabilities", "이 CLI가 지원하는 명령"},
		{"version", "버전"},
	}},
}

var quickStart = []helpEntry{
	{"riido auth login", "브라우저로 로그인"},
	{"riido workspace list", "내 워크스페이스 보기"},
	{"riido workspace use <ID>", "기본 워크스페이스로 저장"},
	{"riido work get RIID-1234", "작업 보기"},
	{"riido doctor", "설치·연결 진단"},
}

var rootExamples = []string{
	"riido auth login",
	"riido work get RIID-1234",
	"riido work search --query \"로그인\"",
	"riido meeting list --work RIID-1234",
}

func versionNumber() string {
	return strings.TrimPrefix(versionLabel(), "riido-daemon ")
}

// homeIntro is the name block shown beside the symbol on the home screen.
func homeIntro(term cliui.Terminal, server string) []string {
	lines := []string{
		term.Brand("Riido CLI") + " " + term.Dim(versionNumber()),
		term.Dim("작업·미팅·Agent 맥락을 터미널에서 조회합니다."),
	}
	if server != "" {
		lines = append(lines, "", term.Dim("서버")+" "+server)
	}
	return lines
}

// runHome is what a person sees after typing just `riido` in a terminal: the
// round symbol with the name as real text beside it (like the Claude Code
// welcome), falling back to symbol-above-text and then text only.
func runHome(w io.Writer, term cliui.Terminal, server string) error {
	intro := homeIntro(term, server)
	introWidth := 0
	for _, line := range intro {
		introWidth = max(introWidth, cliui.DisplayWidth(line))
	}
	symbol := term.SymbolLines()
	// Breathing room between the prompt and the logo, like other CLI welcomes.
	fmt.Fprintln(w)
	switch {
	case symbol != nil && term.Width >= cliui.SymbolWidth()+3+introWidth:
		intro = append([]string{""}, intro...)
		for _, line := range cliui.SideBySide(symbol, cliui.SymbolWidth(), 3, intro) {
			fmt.Fprintln(w, line)
		}
	case symbol != nil:
		for _, line := range symbol {
			fmt.Fprintln(w, line)
		}
		fmt.Fprintln(w)
		writeLines(w, intro)
	default:
		writeLines(w, intro)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, term.Bold("시작하기"))
	printEntries(w, term, quickStart, true, 0)
	fmt.Fprintln(w)
	fmt.Fprintln(w, term.Dim("도움말 riido --help · riido <명령> --help"))
	fmt.Fprintln(w)
	return nil
}

func writeLines(w io.Writer, lines []string) {
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
}

// printFriendlyHelp is the grouped root help for people. Pipes, agents and
// `riido help --all` keep the complete legacy usage.
func printFriendlyHelp(w io.Writer, term cliui.Terminal) {
	fmt.Fprintf(w, "%s  %s\n", term.Mark(), term.Dim(versionNumber()))
	fmt.Fprintln(w, "작업·미팅·Agent 맥락을 터미널에서 조회합니다.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, term.Bold("사용법"))
	fmt.Fprintln(w, "  riido <명령> [하위 명령] [옵션]")
	width := 0
	for _, group := range friendlyHelpGroups {
		for _, entry := range group.entries {
			width = max(width, cliui.DisplayWidth(entry.name))
		}
	}
	for _, group := range friendlyHelpGroups {
		fmt.Fprintln(w)
		fmt.Fprintln(w, term.Bold(group.title))
		printEntries(w, term, group.entries, false, width)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, term.Bold("공통 옵션"))
	printEntries(w, term, commonHelpFlags, false, width)
	printExamples(w, term, rootExamples)
	fmt.Fprintln(w)
	fmt.Fprintln(w, term.Bold("더 알아보기"))
	fmt.Fprintln(w, "  명령별 사용법   riido <명령> --help")
	fmt.Fprintln(w, "  전체 명령       riido help --all  "+term.Dim("(Agent·고급 명령 포함)"))
}

var commonHelpFlags = []helpEntry{
	{"--json", "스크립트용 JSON 출력"},
	{"--profile", "local | staging | production"},
	{"--no-color", "색 없이 출력 (NO_COLOR도 지원)"},
}

func printExamples(w io.Writer, term cliui.Terminal, examples []string) {
	if len(examples) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, term.Bold("예시"))
	for _, example := range examples {
		fmt.Fprintf(w, "  %s%s\n", term.Dim("$ "), example)
	}
}

func printEntries(w io.Writer, term cliui.Terminal, entries []helpEntry, commands bool, width int) {
	summaryWidth := 0
	for _, entry := range entries {
		width = max(width, cliui.DisplayWidth(entry.name))
		summaryWidth = max(summaryWidth, cliui.DisplayWidth(entry.summary))
	}
	// A narrow terminal would wrap the two columns, so put each summary under
	// its command instead.
	stacked := term.Interactive && term.Width < 2+width+2+summaryWidth
	for _, entry := range entries {
		name := entry.name
		if !stacked {
			name = cliui.PadRight(name, width)
		}
		if commands {
			name = term.Accent(name)
		} else {
			name = term.Bold(name)
		}
		if stacked {
			fmt.Fprintf(w, "  %s\n    %s\n", name, term.Dim(entry.summary))
			continue
		}
		fmt.Fprintf(w, "  %s  %s\n", name, entry.summary)
	}
}

// runHelp handles `riido help [--all | <command> [<subcommand>]]`.
func runHelp(w io.Writer, args []string) error {
	term := humanTerminal(w)
	if len(args) == 1 && args[0] == "--all" {
		printCommandUsage(w, false)
		return nil
	}
	if len(args) > 0 && term.Interactive {
		if topic, ok := commandHelpCatalog[args[0]]; ok && len(args) <= 2 {
			if len(args) == 2 {
				if sub := topic.sub(args[1]); sub != nil {
					printSubcommandHelp(w, term, topic, *sub)
					return nil
				}
			} else {
				printCommandHelp(w, term, topic)
				return nil
			}
		}
	}
	if len(args) > 0 {
		return fmt.Errorf("usage: riido help [--all | <command> [<subcommand>]]")
	}
	if !term.Interactive {
		printCommandUsage(w, false)
		return nil
	}
	printFriendlyHelp(w, term)
	return nil
}
