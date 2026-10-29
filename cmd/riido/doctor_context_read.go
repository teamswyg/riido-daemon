package main

import "github.com/teamswyg/riido-daemon/internal/contextcli"

func (c contextIO) doctorContextRead(env contextcli.Environment, requestID string, reachable bool) doctorCheck {
	if !reachable {
		return doctorCheck{Name: "context_read", Status: doctorSkip, Detail: "daemon에 연결되지 않아 확인하지 않았습니다"}
	}
	if env.Mode != contextcli.ModeAgent {
		return doctorCheck{Name: "context_read", Status: doctorSkip, Detail: "사용자 모드에는 실행 capability가 없어 확인하지 않았습니다"}
	}
	socket, err := c.daemonSocket(env)
	if err != nil {
		return doctorCheck{Name: "context_read", Status: doctorSkip, Detail: "daemon 소켓 경로를 찾지 못했습니다"}
	}
	body, cliErr := c.daemonRoundTrip(socket, doctorProbeTimeout, contextcli.NewContextReadRequest(env, requestID, "capabilities", nil))
	if cliErr != nil {
		return doctorProbeError(cliErr)
	}
	reply, replyErr := contextcli.DecodeContextReadReply(body)
	if replyErr != nil {
		return doctorProbeError(replyErr)
	}
	if replyErr = contextcli.ReplyError(reply); replyErr != nil {
		return doctorProbeError(replyErr)
	}
	data, ok := reply.Data.(map[string]any)
	if !ok || data["broker"] != "daemon" {
		return doctorCheck{Name: "context_read", Status: doctorFail, Detail: "daemon이 유효한 capabilities 응답을 보내지 않았습니다"}
	}
	return doctorCheck{Name: "context_read", Status: doctorOK, Detail: "daemon이 context_read를 지원합니다 (원격 API는 확인하지 않음)"}
}

func doctorProbeError(err *contextcli.CLIError) doctorCheck {
	status := doctorFail
	if err.ExitCode() == contextcli.ExitUnsupported {
		status = doctorUnsupported
	}
	return doctorCheck{Name: "context_read", Status: status, Detail: err.Message}
}
