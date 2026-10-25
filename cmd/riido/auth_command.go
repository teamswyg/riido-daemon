package main

import (
	"fmt"
	"io"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

type authCommand string

const (
	authCommandLogin  authCommand = "login"
	authCommandLogout authCommand = "logout"
	authCommandStatus authCommand = "status"
)

// authStatusData never carries a credential value, only its presence.
type authStatusData struct {
	Mode                 contextcli.Mode `json:"mode"`
	RunCapabilityPresent bool            `json:"run_capability_present"`
	UserCredential       string          `json:"user_credential"`
	DaemonSocket         string          `json:"daemon_socket,omitempty"`
	Note                 string          `json:"note"`
}

const userCredentialUnsupported = "unsupported"

func runAuth(c contextIO, args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		printAuthUsage(c.stderr)
		if len(args) == 0 {
			return c.fail(false, contextcli.NewRequestID(), contextcli.Fail(contextcli.CodeUsage, "auth 하위 명령이 필요합니다", "riido auth --help"))
		}
		return nil
	}
	switch authCommand(args[0]) {
	case authCommandLogin:
		return c.runAuthLogin(args[1:])
	case authCommandLogout:
		return c.runAuthLogout(args[1:])
	case authCommandStatus:
		return c.runAuthStatus(args[1:])
	default:
		return c.fail(wantsJSON(args), contextcli.NewRequestID(),
			contextcli.Failf(contextcli.CodeUsage, "알 수 없는 auth 하위 명령: %s", args[0]))
	}
}

// runAuthLogin validates the login form but never reads stdin until the
// keychain store with user-presence protection exists.
func (c contextIO) runAuthLogin(raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{Bools: []string{"--desktop", "--browser", "--with-token"}})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printAuthUsage(c.stderr)
		return nil
	}
	chosen := 0
	for _, flag := range []string{"--desktop", "--browser", "--with-token"} {
		if args.Bool(flag) {
			chosen++
		}
	}
	if chosen != 1 {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUsage,
			"--desktop, --browser, --with-token 중 하나만 고르세요", "riido auth login --browser"))
	}
	if c.environment().Mode == contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodePermissionDenied,
			"에이전트 실행 안에서는 사용자 로그인을 할 수 없습니다", "에이전트는 실행 capability로만 조회합니다"))
	}
	return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeUserModeUnavailable,
		"사용자 모드 로그인은 아직 지원하지 않습니다 (서버 인증 계약과 키체인 보호 확정 전)",
		"Riido에서 Agent를 실행하면 에이전트 모드로 조회할 수 있습니다"))
}

func (c contextIO) runAuthLogout(raw []string) error {
	return c.runAuthLocal(raw, "이 버전은 사용자 자격을 저장하지 않으므로 지울 자격이 없습니다")
}

func (c contextIO) runAuthStatus(raw []string) error {
	return c.runAuthLocal(raw, "사용자 모드 자격 저장은 아직 지원하지 않습니다")
}

func (c contextIO) runAuthLocal(raw []string, note string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, contextcli.ArgSpec{})
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		printAuthUsage(c.stderr)
		return nil
	}
	env := c.environment()
	data := authStatusData{
		Mode:                 env.Mode,
		RunCapabilityPresent: env.HasRunCapability(),
		UserCredential:       userCredentialUnsupported,
		DaemonSocket:         env.DaemonSocket,
		Note:                 note,
	}
	return c.succeed(args.JSON, contextcli.NewEnvelope(requestID, data), renderAuthStatus)
}

func renderAuthStatus(w io.Writer, data any) {
	status, ok := data.(authStatusData)
	if !ok {
		return
	}
	fmt.Fprintf(w, "모드: %s\n", status.Mode)
	fmt.Fprintf(w, "실행 capability: %t\n", status.RunCapabilityPresent)
	fmt.Fprintf(w, "사용자 자격: %s\n", status.UserCredential)
	fmt.Fprintf(w, "참고: %s\n", status.Note)
}

func runWhoami(c contextIO, args []string) error {
	return c.runContextRead(contextReadCall{
		command: "whoami",
		spec:    contextcli.ArgSpec{},
		usage: func(w io.Writer) {
			fmt.Fprintln(w, "  riido whoami [--json]")
		},
		payload: func(contextcli.Args) (map[string]string, *contextcli.CLIError) {
			return map[string]string{}, nil
		},
	}, args)
}
