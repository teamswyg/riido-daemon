package main

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

const (
	contextDialTimeout = 2 * time.Second
	contextReplyLimit  = 16 << 20
)

// contextReadCall describes one agent-mode read command.
type contextReadCall struct {
	command string
	spec    contextcli.ArgSpec
	usage   func(w io.Writer)
	payload func(args contextcli.Args) (map[string]string, *contextcli.CLIError)
	render  textRenderer
}

// runContextRead parses args, enforces the mode rules, and asks the daemon
// broker over local IPC. It never calls Riido servers directly.
func (c contextIO) runContextRead(call contextReadCall, raw []string) error {
	requestID := contextcli.NewRequestID()
	args, cliErr := contextcli.ParseArgs(raw, call.spec)
	if cliErr != nil {
		return c.fail(wantsJSON(raw), requestID, cliErr)
	}
	if args.Help {
		call.usage(c.stderr)
		return nil
	}
	payload, cliErr := call.payload(args)
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	if contextcli.ReadCommandUnavailable(call.command) {
		return c.fail(args.JSON, requestID, contextcli.Fail(contextcli.CodeControlPlaneUnsupported,
			call.command+" 명령은 아직 지원하지 않습니다", "riido context get 또는 riido work get을 사용하세요"))
	}
	env := c.environment()
	if cliErr := checkRunScope(env, args); cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	if env.Mode != contextcli.ModeAgent {
		return c.fail(args.JSON, requestID, userModeUnavailable())
	}
	addListOptions(payload, args)
	reply, cliErr := c.contextRead(env, args, contextcli.NewContextReadRequest(env, requestID, call.command, payload))
	if cliErr != nil {
		return c.fail(args.JSON, requestID, cliErr)
	}
	return c.succeed(args.JSON, reply, call.render)
}

// checkRunScope rejects flags that would leave the run's workspace. The
// broker and control-plane enforce the full grant; this is an early, local
// refusal so agents get a clear error.
func checkRunScope(env contextcli.Environment, args contextcli.Args) *contextcli.CLIError {
	if env.Mode != contextcli.ModeAgent || args.Workspace == "" || env.WorkspaceID == "" {
		return nil
	}
	if args.Workspace != env.WorkspaceID {
		return contextcli.Fail(contextcli.CodeOutOfRunScope,
			"이 실행에 허용된 워크스페이스 밖의 요청입니다", "--workspace 없이 다시 실행하세요")
	}
	return nil
}

func userModeUnavailable() *contextcli.CLIError {
	return contextcli.Fail(contextcli.CodeUserModeUnavailable,
		"사용자 모드 조회는 아직 지원하지 않습니다 (서버 인증 계약 확정 전)",
		"Riido에서 Agent를 실행하면 에이전트 모드로 조회할 수 있습니다")
}

func addListOptions(payload map[string]string, args contextcli.Args) {
	if args.Cursor != "" {
		payload["cursor"] = args.Cursor
	}
}

func (c contextIO) contextRead(env contextcli.Environment, args contextcli.Args, req contextcli.ContextReadRequest) (contextcli.Envelope, *contextcli.CLIError) {
	socket, err := c.daemonSocket(env)
	if err != nil {
		return contextcli.Envelope{}, contextcli.Fail(contextcli.CodeDaemonUnreachable, "daemon 소켓 경로를 찾지 못했습니다", "riido doctor")
	}
	body, cliErr := c.daemonRoundTrip(socket, args.Timeout, req)
	if cliErr != nil {
		return contextcli.Envelope{}, cliErr
	}
	return contextcli.DecodeContextReadReply(body)
}

func (c contextIO) daemonRoundTrip(socket string, timeout time.Duration, request any) ([]byte, *contextcli.CLIError) {
	conn, err := c.dial(socket, contextDialTimeout)
	if err != nil {
		return nil, contextcli.Fail(contextcli.CodeDaemonUnreachable,
			"Riido daemon에 연결할 수 없습니다", "Riido Desktop이 실행 중인지 확인하고 riido doctor 를 실행하세요")
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))
	if err := json.NewEncoder(conn).Encode(request); err != nil {
		return nil, contextcli.Fail(contextcli.CodeUnavailable, "daemon에 요청을 보내지 못했습니다", "잠시 뒤 다시 시도하세요")
	}
	body, err := io.ReadAll(io.LimitReader(conn, contextReplyLimit))
	if err != nil && !errors.Is(err, io.EOF) {
		if isTimeout(err) {
			return nil, contextcli.Fail(contextcli.CodeUnavailable, "daemon 응답 시간이 초과됐습니다", "--timeout 을 늘리거나 잠시 뒤 다시 시도하세요")
		}
		return nil, contextcli.Fail(contextcli.CodeUnavailable, "daemon 응답을 읽지 못했습니다", "잠시 뒤 다시 시도하세요")
	}
	return body, nil
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
