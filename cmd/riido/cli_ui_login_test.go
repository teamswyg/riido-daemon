package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func TestLoginResultTerminalView(t *testing.T) {
	var buf bytes.Buffer
	result := standalonecli.LoginResult{
		APIOrigin: "https://monolith.api.riido.io", AuthOrigin: "browser", Profile: "staging",
		Identity: userapi.Identity{ID: "u1", Email: "me@example.com", Username: "yoo\x1b[2J"},
	}
	renderLoginResult(&buf, cliui.Terminal{Interactive: true, Width: 100}, result)
	out := buf.String()
	for _, part := range []string{"✓ 로그인했어요: yoo\\x1b[2J (me@example.com)", "서버: monolith.api.riido.io (staging)", "로그인 방식: 브라우저", "riido workspace list"} {
		if !strings.Contains(out, part) {
			t.Fatalf("login view missing %q:\n%s", part, out)
		}
	}
	if strings.Contains(out, "\x1b[2J") {
		t.Fatal("control sequence not escaped")
	}
}

// Off a terminal (and with --json) login prints no progress lines and the
// browser opener is passed through untouched.
func TestLoginNoticesStayOffPipes(t *testing.T) {
	var stderr bytes.Buffer
	c := contextIO{stderr: &stderr}
	c.loginNotice(false, "x")
	opened := false
	open := c.announcingBrowser(false, func(context.Context, string) error { opened = true; return nil })
	if err := open(context.Background(), "https://example.com"); err != nil || !opened || stderr.Len() != 0 {
		t.Fatalf("pipe notices: opened=%v stderr=%q", opened, stderr.String())
	}
	if c.loginNoticeTerminal(true).Interactive {
		t.Fatal("--json must stay plain")
	}
}

// Warnings from a successful login stay visible on the terminal view.
func TestLoginResultShowsWarnings(t *testing.T) {
	var buf bytes.Buffer
	renderLoginResult(&buf, cliui.Terminal{Interactive: true, Width: 100}, standalonecli.LoginResult{
		Identity: userapi.Identity{ID: "u1"}, Warnings: []string{"previous_grant_revocation_failed", "other"},
	})
	out := buf.String()
	if !strings.Contains(out, "previous_grant_revocation_failed") || !strings.Contains(out, "이전 로그인") || !strings.Contains(out, "! other") {
		t.Fatalf("warnings missing:\n%s", out)
	}
}
