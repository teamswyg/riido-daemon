package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func userAuthError(err error) *contextcli.CLIError {
	switch {
	case errors.Is(err, userauth.ErrNotAuthenticated), errors.Is(err, userauth.ErrReauthRequired), errors.Is(err, userapi.ErrAuthentication):
		return contextcli.Fail(contextcli.CodeAuthRequired, "Riido CLI 로그인이 필요하거나 만료되었습니다", "riido auth login")
	case errors.Is(err, userauth.ErrStoreLocked):
		return contextcli.Fail(contextcli.CodeAuthRequired, "OS 사용자 확인이 필요하거나 취소되었습니다", "대화형 터미널에서 --no-input 없이 다시 실행하세요")
	case errors.Is(err, securecredential.ErrUnavailable):
		return contextcli.Fail(contextcli.CodeUnsupported, "서명된 CLI 인증 보호 도우미를 사용할 수 없습니다", "Riido CLI의 인증 보호 도우미 설치와 서명을 확인하세요")
	case errors.Is(err, userapi.ErrPermission):
		return contextcli.Fail(contextcli.CodePermissionDenied, "현재 사용자에게 이 자료의 열람 권한이 없습니다", "계정과 작업 권한을 확인하세요")
	case errors.Is(err, userapi.ErrNotFound):
		return contextcli.Fail(contextcli.CodeNotFound, "요청한 자료를 찾을 수 없습니다", "작업 ID와 팀 ID를 확인하세요")
	case errors.Is(err, userauth.ErrLoggedOut):
		return contextcli.Fail(contextcli.CodeAuthRequired, "요청 중 CLI 인증이 변경되었습니다", "riido auth status 이후 다시 실행하세요")
	case errors.Is(err, standalonecli.ErrLogoutRemoteIncomplete):
		return contextcli.Fail(contextcli.CodeIncomplete, "로컬 자격은 삭제했으나 서버 인증 폐기를 확인하지 못했습니다", "서버 연결의 폐기 상태를 확인하세요")
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return contextcli.Fail(contextcli.CodeUnavailable, "CLI 인증 또는 조회가 취소되었거나 시간이 초과됐습니다", "필요하면 --timeout 값을 늘려 다시 실행하세요")
	}
	var authErr *userauth.Error
	if errors.As(err, &authErr) {
		if authErr.Code == "STORAGE_ERROR" {
			return userAuthError(securecredential.ErrUnavailable)
		}
		if authErr.Code == "INVALID_CONFIG" {
			return contextcli.Fail(contextcli.CodeUsage, "CLI API 주소 설정이 올바르지 않습니다", "RIIDO_CLI_API_URL에는 HTTPS API origin을 지정하세요")
		}
		if authErr.StatusCode == http.StatusTooManyRequests {
			return contextcli.Fail(contextcli.CodeRateLimited, "요청 한도를 초과했습니다", "잠시 뒤 다시 시도하세요")
		}
		if authErr.Code == "AUTH_DENIED" {
			return contextcli.Fail(contextcli.CodeAuthRequired, "브라우저에서 CLI 인증이 승인되지 않았습니다", "riido auth login")
		}
	}
	return contextcli.Fail(contextcli.CodeUnavailable, "CLI 인증 또는 사용자 API 요청을 완료하지 못했습니다", "계정·연결 상태를 확인한 뒤 다시 실행하세요")
}
