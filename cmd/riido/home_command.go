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
	{"riido work get RIID-1234", "작업 보기"},
	{"riido doctor", "설치·연결 진단"},
}

func versionNumber() string {
	return strings.TrimPrefix(versionLabel(), "riido-daemon ")
}

// runHome is what a person sees after typing just `riido` in a terminal.
func runHome(w io.Writer, term cliui.Terminal, server string) error {
	if logo := term.Logo(); logo != "" {
		fmt.Fprint(w, logo)
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, "%s %s\n", term.Bold("Riido CLI"), term.Dim(versionNumber()))
	fmt.Fprintln(w, term.Dim("작업·미팅·Agent 맥락을 터미널에서 조회합니다."))
	if server != "" {
		fmt.Fprintf(w, "%s %s\n", term.Dim("서버"), server)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, term.Bold("시작하기"))
	printEntries(w, term, quickStart, true, 0)
	fmt.Fprintln(w)
	fmt.Fprintln(w, term.Dim("도움말 riido --help · 전체 명령 riido help --all"))
	return nil
}

// printFriendlyHelp is the grouped root help for people. Pipes, agents and
// `riido help --all` keep the complete legacy usage.
func printFriendlyHelp(w io.Writer, term cliui.Terminal) {
	fmt.Fprintf(w, "%s  %s\n\n", term.Mark(), term.Dim(versionNumber()))
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
	printEntries(w, term, []helpEntry{
		{"--json", "스크립트용 JSON 출력"},
		{"--profile", "local | staging | production"},
		{"--no-color", "색 없이 출력 (NO_COLOR도 지원)"},
	}, false, width)
	fmt.Fprintln(w)
	fmt.Fprintln(w, term.Dim("명령별 사용법: riido <명령> --help"))
	fmt.Fprintln(w, term.Dim("전체 명령(Agent·고급): riido help --all"))
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

// runHelp handles `riido help [--all]`.
func runHelp(w io.Writer, args []string) error {
	term := humanTerminal(w)
	if len(args) > 1 || len(args) == 1 && args[0] != "--all" {
		return fmt.Errorf("usage: riido help [--all]")
	}
	if len(args) == 1 || !term.Interactive {
		printCommandUsage(w, false)
		return nil
	}
	printFriendlyHelp(w, term)
	return nil
}
