package main

import (
	"context"
	"fmt"
	"io"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

// loginNoticeTerminal is where login progress goes: a person's terminal on
// stderr, never --json output, pipes or agent runs.
func (c contextIO) loginNoticeTerminal(jsonOut bool) cliui.Terminal {
	if jsonOut {
		return cliui.Plain
	}
	return humanTerminal(c.stderr)
}

// loginNotice prints one progress line so a person knows what the browser or
// OS confirmation is waiting for. The authorization URL is never printed: it
// carries the state and PKCE values.
func (c contextIO) loginNotice(jsonOut bool, message string) {
	term := c.loginNoticeTerminal(jsonOut)
	if !term.Interactive {
		return
	}
	fmt.Fprintf(c.stderr, "%s %s\n", term.Accent("›"), message)
}

// announcingBrowser reports a successfully opened approval page.
func (c contextIO) announcingBrowser(jsonOut bool, open func(context.Context, string) error) func(context.Context, string) error {
	if !c.loginNoticeTerminal(jsonOut).Interactive {
		return open
	}
	return func(ctx context.Context, address string) error {
		if err := open(ctx, address); err != nil {
			return err
		}
		c.loginNotice(jsonOut, "브라우저에서 승인 화면을 열었어요. 승인하면 자동으로 이어져요 (취소: Ctrl+C)")
		return nil
	}
}

// renderLoginResult follows `gh auth login`: one success line with the
// account, then where it applies and what to do next.
func renderLoginResult(w io.Writer, term cliui.Terminal, result standalonecli.LoginResult) {
	account := result.Identity.Email
	if result.Identity.Username != "" {
		if account != "" {
			account = result.Identity.Username + " (" + account + ")"
		} else {
			account = result.Identity.Username
		}
	}
	if account == "" {
		account = result.Identity.ID
	}
	fmt.Fprintf(w, "%s 로그인했어요: %s\n", term.Symbol("ok"), term.Bold(cliui.Safe(account)))
	server := apiHost(result.APIOrigin)
	if result.Profile != "" {
		server += " (" + result.Profile + ")"
	}
	fmt.Fprintf(w, "  - 서버: %s\n", cliui.Safe(server))
	fmt.Fprintf(w, "  - 로그인 방식: %s\n", authOriginLabel(result.AuthOrigin))
	for _, warning := range result.Warnings {
		fmt.Fprintf(w, "%s %s\n", term.Symbol("warn"), loginWarningText(warning))
	}
	if result.NextAction != "" {
		fmt.Fprintf(w, "%s %s\n", term.Symbol("warn"), cliui.Safe(result.NextAction))
		return
	}
	fmt.Fprintf(w, "\n%s 다음: riido workspace list 로 워크스페이스를 확인하세요\n", term.Accent("→"))
}

// loginWarningText explains the warnings Service.Login can return; unknown
// codes are shown as they are so none is hidden.
func loginWarningText(code string) string {
	switch code {
	case "previous_grant_revocation_failed":
		return "이전 로그인을 서버에서 철회하지 못했어요. 이전 로그인이 아직 유효할 수 있어요 (previous_grant_revocation_failed)"
	case "auth_origin_selection_not_saved":
		return "로그인 방식 선택을 설정에 저장하지 못했어요 (auth_origin_selection_not_saved)"
	}
	return cliui.Safe(code)
}
