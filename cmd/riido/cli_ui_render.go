package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func renderDoctorTerminal(w io.Writer, term cliui.Terminal, report doctorData) {
	fmt.Fprintln(w, term.Bold("Riido CLI 진단"))
	width := 0
	for _, check := range report.Checks {
		width = max(width, cliui.DisplayWidth(doctorLabel(check.Name)))
	}
	for _, check := range report.Checks {
		status := string(check.Status)
		symbol := term.Symbol(map[string]string{"ok": "ok", "warn": "warn", "fail": "fail", "unsupported": "fail"}[status])
		detail := cliui.Safe(check.Detail)
		if status == "skip" {
			detail = term.Dim(detail)
		}
		fmt.Fprintf(w, "  %s %s  %s\n", symbol, cliui.PadRight(doctorLabel(check.Name), width), detail)
	}
	warnings := 0
	for _, check := range report.Checks {
		if string(check.Status) == "warn" {
			warnings++
		}
	}
	fmt.Fprintln(w)
	switch {
	case !report.Healthy:
		fmt.Fprintf(w, "%s %s\n", term.Symbol("fail"), "사용하기 전에 해결이 필요한 항목이 있어요")
	case warnings > 0:
		fmt.Fprintf(w, "%s %s\n", term.Symbol("warn"), fmt.Sprintf("경고 %d개: 일부 기능이 제한될 수 있어요", warnings))
	default:
		fmt.Fprintf(w, "%s %s\n", term.Symbol("ok"), "사용할 준비가 됐어요")
	}
	for _, check := range report.Checks {
		if check.Name == "api" && check.Status == doctorSkip {
			fmt.Fprintln(w, term.Dim("  로그인·API까지 확인하려면 riido doctor --check-api"))
		}
	}
}

// doctorLabels names checks for people; --json keeps the stable check names.
var doctorLabels = map[string]string{
	"binary":           "실행 파일",
	"path":             "PATH",
	"mode":             "실행 모드",
	"binary_target":    "실제 경로",
	"package_version":  "패키지 버전",
	"helper_bundle":    "인증 도우미",
	"helper_signature": "도우미 서명",
	"daemon_ipc":       "Riido 앱 연결",
	"context_read":     "맥락 조회",
	"api_origin":       "서버",
	"user_auth":        "로그인",
	"api":              "API 확인",
}

func doctorLabel(name string) string {
	if label, ok := doctorLabels[name]; ok {
		return label
	}
	return cliui.Safe(name)
}

func renderWorkspacesTerminal(w io.Writer, term cliui.Terminal, page *userapi.WorkspaceList) {
	rows := make([][]string, 0, len(page.Items))
	for _, item := range page.Items {
		name := cliui.Safe(item.Name)
		if item.IsDemo {
			name += " " + term.Dim("(demo)")
		}
		rows = append(rows, []string{cliui.Safe(item.ID), name})
	}
	if len(rows) == 0 {
		fmt.Fprintln(w, term.Dim("참여 중인 워크스페이스가 없어요."))
		return
	}
	_ = term.Table(w, []string{"ID", "이름"}, rows)
	if page.NextCursor != nil {
		fmt.Fprintln(w, term.Dim(fmt.Sprintf("다음 페이지: riido workspace list --limit %d --cursor %s", page.Limit, cliui.Safe(*page.NextCursor))))
	}
	fmt.Fprintln(w, term.Dim("기본으로 쓰려면: riido workspace use <ID>"))
}

func renderUserAgentsTerminal(w io.Writer, term cliui.Terminal, result any) {
	var agents []userapi.PublicAgent
	var next string
	switch value := result.(type) {
	case *userapi.AgentPage:
		agents, next = value.Agents, value.NextCursor
	case *userapi.AgentDetail:
		agents = []userapi.PublicAgent{value.Agent}
	}
	rows := make([][]string, 0, len(agents))
	for _, agent := range agents {
		connection := cliui.Safe(agent.Runtime.Connection)
		switch connection {
		case "online":
			connection = term.Success("● online")
		case "offline":
			connection = term.Dim("○ offline")
		}
		description := cliui.Safe(strings.Join(strings.Fields(agent.Description), " "))
		rows = append(rows, []string{cliui.Safe(agent.Name), cliui.Safe(agent.AgentID), connection, cliui.Safe(agent.WorkStatus), description})
	}
	if len(rows) == 0 {
		fmt.Fprintln(w, term.Dim("볼 수 있는 Agent가 없어요."))
		return
	}
	_ = term.Table(w, []string{"이름", "ID", "연결", "작업", "설명"}, rows)
	if next != "" {
		fmt.Fprintln(w, term.Dim("다음 페이지: 같은 옵션에 --cursor "+cliui.Safe(next)))
	}
	fmt.Fprintln(w, term.Dim("연결 상태는 멘션·실행 허가를 뜻하지 않아요."))
}
