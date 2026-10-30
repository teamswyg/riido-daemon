package main

import (
	"errors"
	"net/http"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func userLinkError(err error) *contextcli.CLIError {
	var remote *userauth.DeviceLinkError
	var local *contextcli.CLIInstallationLinkError
	code := ""
	if errors.As(err, &remote) {
		code = remote.Code
		if remote.StatusCode == http.StatusConflict || remote.StatusCode == http.StatusGone {
			code = "LINK_REPAIR_REQUIRED"
		}
	}
	if errors.As(err, &local) {
		code = local.Code
	}
	switch code {
	case "LINK_APPROVAL_REQUIRED":
		return contextcli.Fail(contextcli.CodeAuthRequired, "기기 연결에는 브라우저 승인과 OS 사용자 확인이 필요합니다", "기존 요청은 유지됩니다. --no-input 없이 riido auth link 를 실행하세요")
	case "LINK_BINDING_MISMATCH", "LINK_REQUEST_CHANGED":
		return contextcli.Fail(contextcli.CodeConflict, "기존 요청의 계정·설치 키·기기가 현재 상태와 다릅니다", "기존 요청은 유지됩니다. 원래 계정과 기기를 확인하세요")
	case "LINK_PROOF_EXPIRED", "LINK_REPAIR_REQUIRED":
		return contextcli.Fail(contextcli.CodeConflict, "기기 연결 요청을 다시 승인할 준비가 필요합니다", "기존 요청은 유지됩니다. 자동 재승인·초기화는 지원하지 않습니다")
	case "LINK_DAEMON_UNAVAILABLE":
		return contextcli.Fail(contextcli.CodeDaemonUnreachable, "기기 연결을 확인할 daemon에 연결할 수 없습니다", "기존 요청은 유지됩니다. 원래 Riido Desktop 기기에서 다시 실행하세요")
	case "LINK_RATE_LIMITED":
		return contextcli.Fail(contextcli.CodeRateLimited, "기기 연결 요청 한도를 초과했습니다", "기존 요청은 유지됩니다. 잠시 뒤 riido auth link 를 다시 실행하세요")
	}
	if code != "" {
		return contextcli.Fail(contextcli.CodeUnavailable, "서버에서 기기 연결 완료를 확인하지 못했습니다", "기존 요청은 유지됩니다. 잠시 뒤 riido auth link 를 다시 실행하세요")
	}
	if errors.Is(err, securecredential.ErrNotFound) {
		return contextcli.Fail(contextcli.CodeConflict, "기존 설치 키를 찾을 수 없습니다", "기존 요청은 유지됩니다. 원래 OS 계정과 인증 보호 도우미를 확인하세요")
	}
	return userAuthError(err)
}
