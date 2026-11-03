package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestAuthStatusTerminalViews(t *testing.T) {
	term := cliui.Terminal{Interactive: true, Width: 100}
	var buf bytes.Buffer
	renderUserAuthStatus(&buf, term, standalonecli.Status{APIOrigin: "https://monolith.api.riido.io", AuthOrigin: "browser"})
	out := buf.String()
	for _, part := range []string{"monolith.api.riido.io\n", "✗ 로그인하지 않았어요", "로그인 방식: 브라우저", "riido auth login"} {
		if !strings.Contains(out, part) {
			t.Fatalf("logged-out status missing %q:\n%s", part, out)
		}
	}
	buf.Reset()
	renderUserAuthStatus(&buf, term, standalonecli.Status{APIOrigin: "https://monolith.riido.io", AuthOrigin: "desktop", CredentialPresent: true})
	if !strings.Contains(buf.String(), "✓ 로그인 정보가 저장되어 있어요") || !strings.Contains(buf.String(), "Riido 앱") || strings.Contains(buf.String(), "auth login") {
		t.Fatalf("stored status:\n%s", buf.String())
	}
}

func TestTerminalOrJSONKeepsJSONForPipes(t *testing.T) {
	var buf bytes.Buffer
	terminalOrJSON(func(w io.Writer, _ cliui.Terminal) { t.Fatal("view used for a buffer") })(&buf, map[string]any{"a": 1})
	if strings.TrimSpace(buf.String()) != "{\n  \"a\": 1\n}" {
		t.Fatalf("pipe output = %q", buf.String())
	}
}

func TestConfigTerminalViewAlignsAndEscapes(t *testing.T) {
	var buf bytes.Buffer
	values := map[string]string{"profile": "staging", "workspace": "", "team": "evil\x1b[2J"}
	sources := map[string]string{"profile": "config", "workspace": "default", "team": "profile"}
	renderConfigTerminal(&buf, cliui.Terminal{Interactive: true, Width: 100}, "staging", configKeys, values, sources)
	out := buf.String()
	for _, part := range []string{"설정 프로필 staging", "profile    staging", "설정 파일", "(비어 있음)", "기본값", `evil\x1b[2J`, "프로필 설정"} {
		if !strings.Contains(out, part) {
			t.Fatalf("config view missing %q:\n%s", part, out)
		}
	}
	if strings.Contains(out, "\x1b[2J") {
		t.Fatal("control sequence not escaped")
	}
}
