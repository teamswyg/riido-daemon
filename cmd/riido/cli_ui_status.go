package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

// terminalOrJSON renders view on a person's terminal and keeps the previous
// indented JSON text for pipes, files and agent runs.
func terminalOrJSON(view func(io.Writer, cliui.Terminal)) textRenderer {
	return func(w io.Writer, data any) {
		term := humanTerminal(w)
		if !term.Interactive {
			_ = writeIndentedJSON(w, data)
			return
		}
		view(w, term)
	}
}

func apiHost(origin string) string {
	return strings.TrimPrefix(strings.TrimPrefix(origin, "https://"), "http://")
}

func authOriginLabel(origin string) string {
	switch origin {
	case "desktop":
		return "Riido 앱"
	case "browser", "":
		return "브라우저"
	}
	return cliui.Safe(origin)
}

// renderUserAuthStatus follows `gh auth status`: the host, then one marked line.
func renderUserAuthStatus(w io.Writer, term cliui.Terminal, status standalonecli.Status) {
	fmt.Fprintln(w, term.Bold(cliui.Safe(apiHost(status.APIOrigin))))
	if status.CredentialPresent {
		fmt.Fprintf(w, "  %s 로그인 정보가 저장되어 있어요\n", term.Symbol("ok"))
	} else {
		fmt.Fprintf(w, "  %s 로그인하지 않았어요\n", term.Symbol("fail"))
	}
	fmt.Fprintf(w, "  - 로그인 방식: %s\n", authOriginLabel(status.AuthOrigin))
	if status.CredentialPresent {
		fmt.Fprintln(w, term.Dim("  저장 여부만 확인했어요. 계정까지 확인하려면 riido whoami"))
		return
	}
	fmt.Fprintf(w, "\n%s %s\n", term.Accent("→"), "riido auth login 으로 로그인하세요")
}

func renderLogout(w io.Writer, term cliui.Terminal, origin string) {
	fmt.Fprintf(w, "%s 로그아웃했어요 %s\n", term.Symbol("ok"), term.Dim("("+authOriginLabel(origin)+" 로그인)"))
}

var configSourceLabels = map[string]string{
	"config":      "설정 파일",
	"profile":     "프로필 설정",
	"environment": "환경 변수",
	"flag":        "옵션",
	"default":     "기본값",
}

// renderConfigTerminal shows settings as an aligned key/value/source list.
func renderConfigTerminal(w io.Writer, term cliui.Terminal, profile string, keys []string, values, sources map[string]string) {
	fmt.Fprintf(w, "%s %s\n", term.Bold("설정"), term.Dim("프로필 "+cliui.Safe(profile)))
	keyWidth, valueWidth := 0, 0
	shown := make([]string, 0, len(keys))
	for _, key := range keys {
		if _, ok := values[key]; !ok {
			continue
		}
		shown = append(shown, key)
		keyWidth = max(keyWidth, cliui.DisplayWidth(key))
		valueWidth = max(valueWidth, cliui.DisplayWidth(configValueText(values[key])))
	}
	for _, key := range shown {
		value := configValueText(values[key])
		styled := cliui.PadRight(value, valueWidth)
		if values[key] == "" {
			styled = term.Dim(styled)
		}
		source := configSourceLabels[sources[key]]
		if source == "" {
			source = cliui.Safe(sources[key])
		}
		fmt.Fprintf(w, "  %s  %s  %s\n", cliui.PadRight(key, keyWidth), styled, term.Dim(source))
	}
}

func configValueText(value string) string {
	if value == "" {
		return "(비어 있음)"
	}
	return cliui.Safe(value)
}
