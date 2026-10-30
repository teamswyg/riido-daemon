package main

import (
	"errors"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func userRepairError(err error) *contextcli.CLIError {
	var remote *userauth.InstallationRepairError
	if errors.As(err, &remote) {
		switch remote.Code {
		case "REPAIR_APPROVAL_REQUIRED":
			return contextcli.Fail(contextcli.CodeAuthRequired, "연결 복구에는 브라우저 승인과 OS 사용자 확인이 필요합니다", "기존 요청은 유지됩니다. --no-input 없이 riido auth repair 를 실행하세요")
		case "REPAIR_APPROVAL_EXPIRED", "REPAIR_EXPIRED":
			return contextcli.Fail(contextcli.CodeConflict, "기존 복구 요청에 새 승인이 필요합니다", "새 요청을 승인하려면 riido auth repair --restart 를 실행하세요")
		case "REPAIR_CONSENT_REJECTED":
			return contextcli.Fail(contextcli.CodeConflict, "연결 복구 승인이 거절됐습니다", "기존 요청은 유지됩니다. 다시 승인하려면 riido auth repair --restart 를 실행하세요")
		case "REPAIR_RECOVERY_EXPIRED":
			return contextcli.Fail(contextcli.CodeConflict, "기존 복구 요청의 유효기간이 끝났습니다", "이미 연결을 완료했다면 riido auth login 으로 로그인하세요. 연결 복구가 계속 필요하면 riido auth repair --restart 로 새 승인을 요청하세요")
		case "REPAIR_IN_PROGRESS":
			return contextcli.Fail(contextcli.CodeConflict, "기존 복구 요청이 진행 중입니다", "기존 요청을 이어가려면 --restart 없이 riido auth repair 를 실행하세요")
		case "REPAIR_BINDING_MISMATCH", "REPAIR_REQUEST_CHANGED":
			return contextcli.Fail(contextcli.CodeConflict, "기존 복구 요청의 계정·키·기기가 현재 상태와 다릅니다", "기존 요청은 유지됩니다. 원래 OS 계정과 Riido 기기를 확인하세요")
		case "REPAIR_DAEMON_UNAVAILABLE":
			return contextcli.Fail(contextcli.CodeDaemonUnreachable, "복구할 원래 daemon에 연결할 수 없습니다", "기존 요청은 유지됩니다. 원래 Riido 기기에서 다시 실행하세요")
		case "REPAIR_RATE_LIMITED":
			return contextcli.Fail(contextcli.CodeRateLimited, "복구 요청 한도를 초과했습니다", "잠시 뒤 riido auth repair 를 다시 실행하세요")
		}
		return contextcli.Fail(contextcli.CodeUnavailable, "서버에서 연결 복구 완료를 확인하지 못했습니다", "기존 요청은 유지됩니다. 잠시 뒤 riido auth repair 를 다시 실행하세요")
	}
	if errors.Is(err, userauth.ErrLoggedOut) {
		return contextcli.Fail(contextcli.CodeConflict, "복구를 시작한 뒤 로컬 로그인 상태가 변경됐습니다", "기존 요청은 새 계정으로 진행하지 않습니다. 원래 계정과 요청을 확인하세요")
	}
	if errors.Is(err, securecredential.ErrNotFound) {
		return contextcli.Fail(contextcli.CodeConflict, "복구에 필요한 기존 설치 키를 찾을 수 없습니다", "원래 OS 계정과 인증 보호 도우미를 확인하세요. 새 키로 자동 대체하지 않습니다")
	}
	return userAuthError(err)
}
