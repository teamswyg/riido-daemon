package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/teamswyg/riido-daemon/internal/cliconfig"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func userAuthError(err error) *contextcli.CLIError {
	switch {
	case errors.Is(err, standalonecli.ErrDesktopUnavailable):
		return contextcli.Fail(contextcli.CodeUnavailable, "이 환경의 Riido Desktop 인증을 열 수 없습니다", "호환되는 Riido 앱 설치·로그인을 확인하거나 riido auth login --browser 를 실행하세요")
	case errors.Is(err, cliconfig.ErrInvalid), errors.Is(err, cliconfig.ErrUnavailable), errors.Is(err, cliconfig.ErrConflict):
		return settingsError(err)
	case errors.Is(err, userauth.ErrAgentReadUnsupported):
		return contextcli.Fail(contextcli.CodeUnsupported, "이 API 환경에서는 Agent 조회를 지원하지 않습니다", "Riido production 또는 staging API origin을 사용하세요")
	case errors.Is(err, userauth.ErrAgentReadIncomplete):
		return contextcli.Fail(contextcli.CodeIncomplete, "Agent 응답이 16 MiB 한도를 초과해 조회를 완료하지 못했습니다", "일부 결과는 반환하지 않았습니다. Riido 화면에서 확인하세요")
	case errors.Is(err, userauth.ErrThreadReadUnsupported):
		return contextcli.Fail(contextcli.CodeUnsupported, "이 API 환경에서는 대화 조회를 지원하지 않습니다", "Riido production 또는 staging API origin을 사용하세요")
	case errors.Is(err, userauth.ErrThreadReadIncomplete):
		return contextcli.Fail(contextcli.CodeIncomplete, "대화 응답이 16 MiB 한도를 초과해 조회를 완료하지 못했습니다", "현재 API는 대화 페이지 조회를 지원하지 않습니다. Riido 화면에서 확인하세요")
	case errors.Is(err, userapi.ErrThreadNotRetained):
		return contextcli.Fail(contextcli.CodeNotFound, "현재 권한으로 볼 수 있는 서버 보관분에 해당 대화가 없습니다", "작업·대화 ID를 확인하세요. 과거에 대화가 없었다는 의미는 아닙니다")
	case errors.Is(err, userauth.ErrNotAuthenticated), errors.Is(err, userauth.ErrReauthRequired), errors.Is(err, userapi.ErrAuthentication):
		return contextcli.Fail(contextcli.CodeAuthRequired, "Riido CLI 로그인이 필요하거나 만료되었습니다", "riido auth login")
	case errors.Is(err, userauth.ErrStoreLocked):
		return contextcli.Fail(contextcli.CodeAuthRequired, "OS 사용자 확인(Touch ID 또는 암호)이 취소됐거나 끝나지 않았습니다", "확인 창에서 승인한 뒤 다시 실행하세요. --no-input 을 썼다면 빼고 대화형 터미널에서 실행하세요")
	case errors.Is(err, securecredential.ErrUnavailable):
		return contextcli.Fail(contextcli.CodeUnsupported, "서명된 CLI 인증 보호 도우미를 사용할 수 없습니다", "Riido CLI의 인증 보호 도우미 설치와 서명을 확인하세요")
	case errors.Is(err, userapi.ErrPermission):
		return contextcli.Fail(contextcli.CodePermissionDenied, "현재 사용자에게 이 자료의 열람 권한이 없습니다", "계정과 작업 권한을 확인하세요")
	case errors.Is(err, userapi.ErrNotFound):
		return contextcli.Fail(contextcli.CodeNotFound, "요청한 자료를 찾을 수 없습니다", "작업 ID와 팀 ID를 확인하세요")
	case errors.Is(err, userapi.ErrRateLimited):
		return contextcli.Fail(contextcli.CodeRateLimited, "요청 한도를 초과했습니다", "잠시 뒤 다시 시도하세요")
	case errors.Is(err, userapi.ErrAmbiguousWorkReference):
		return contextcli.Fail(contextcli.CodeConflict, "표시 키에 해당하는 작업이 여러 개입니다", "작업의 실제 ID와 --team-id를 지정하세요")
	case errors.Is(err, userapi.ErrDataIntegrity):
		return contextcli.Fail(contextcli.CodeDataIntegrity, "작업 자료가 올바르지 않아 읽을 수 없습니다", "다시 실행해도 같습니다. Riido 화면에서 작업을 확인하고, 계속되면 관리자에게 알려 주세요")
	case errors.Is(err, userapi.ErrConflict):
		return contextcli.Fail(contextcli.CodeConflict, "자료가 변경되어 기존 페이지를 이어 읽을 수 없습니다", "맥락을 다시 조회해 새 version의 첫 페이지부터 읽으세요")
	case errors.Is(err, userapi.ErrInvalidRequest):
		return contextcli.Fail(contextcli.CodeUsage, "자료 조회의 페이지 또는 입력이 올바르지 않습니다", "원 응답의 version과 다음 페이지 값을 사용하고 명령의 --help를 확인하세요")
	case errors.Is(err, userauth.ErrLoggedOut):
		return contextcli.Fail(contextcli.CodeAuthRequired, "요청 중 CLI 인증이 변경되었습니다", "riido auth status 이후 다시 실행하세요")
	case errors.Is(err, standalonecli.ErrLogoutRemoteIncomplete):
		return contextcli.Fail(contextcli.CodeIncomplete, "로컬 자격은 삭제했으나 서버 인증 폐기를 확인하지 못했습니다", "서버 연결의 폐기 상태를 확인하세요")
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return contextcli.Fail(contextcli.CodeUnavailable, "CLI 인증 또는 조회가 취소되었거나 시간이 초과됐습니다", "OS 확인 창과 서버 연결 상태를 확인한 뒤 다시 실행하세요")
	}
	var authErr *userauth.Error
	if errors.As(err, &authErr) {
		if authErr.Code == "DESKTOP_UNAVAILABLE" {
			return userAuthError(standalonecli.ErrDesktopUnavailable)
		}
		if authErr.Code == "DESKTOP_LOGIN_REQUIRED" {
			return contextcli.Fail(contextcli.CodeAuthRequired, "Riido Desktop에 로그인이 필요합니다", "앱에서 로그인하거나 riido auth login --browser 를 실행하세요")
		}
		if authErr.Code == "AUTH_ORIGIN_MISMATCH" {
			return contextcli.Fail(contextcli.CodeAuthRequired, "서버에서 Desktop 로그인 기원을 확인하지 못했습니다", "서버·앱 버전을 확인하거나 riido auth login --browser 를 실행하세요")
		}
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
			return contextcli.Fail(contextcli.CodeAuthRequired, "CLI 인증이 승인되지 않았습니다", "riido auth login")
		}
	}
	return contextcli.Fail(contextcli.CodeUnavailable, "CLI 인증 또는 사용자 API 요청을 완료하지 못했습니다", "계정·연결 상태를 확인한 뒤 다시 실행하세요")
}
