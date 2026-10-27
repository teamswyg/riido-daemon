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
	if c.environment().Mode != contextcli.ModeAgent {
		return c.runUserWhoami(args)
	}
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
