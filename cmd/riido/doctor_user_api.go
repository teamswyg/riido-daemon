package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func doctorErrorCheck(name string, err *contextcli.CLIError) doctorCheck {
	return doctorCheck{Name: name, Status: doctorFail, Detail: err.Message, ErrorCode: err.Code, NextAction: err.NextAction}
}

func (c contextIO) doctorAPI(args contextcli.Args, env contextcli.Environment, configurationError *contextcli.CLIError) (doctorCheck, bool, *contextcli.CLIError) {
	if !args.Bool("--check-api") {
		return doctorCheck{Name: "api", Status: doctorSkip, Detail: "원격 API 연결·인증은 확인하지 않았습니다"}, false, nil
	}
	if env.Mode == contextcli.ModeAgent {
		return doctorCheck{Name: "api", Status: doctorUnsupported, Detail: "에이전트 모드는 로컬 capability만 진단합니다. 사용자 자격과 원격 조회를 사용하지 않았습니다"}, false, nil
	}
	if configurationError != nil {
		return doctorCheck{Name: "api", Status: doctorSkip, Detail: "API 주소 설정 오류로 요청하지 않았습니다"}, false, nil
	}
	session, err := c.openUserSession(args)
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
		defer cancel()
		_, err = session.Whoami(ctx)
	}
	if err != nil {
		cliErr := userAuthError(err)
		return doctorErrorCheck("api", cliErr), false, cliErr
	}
	return doctorCheck{Name: "api", Status: doctorOK, Detail: "선택한 API에서 보호된 사용자 인증으로 Whoami 조회를 확인했습니다"}, true, nil
}
