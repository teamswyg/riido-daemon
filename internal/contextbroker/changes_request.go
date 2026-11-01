package contextbroker

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func changesSince(args map[string]string) (string, *contextcli.CLIError) {
	for key := range args {
		if key != "since" && key != "budget_chars" {
			return "", contextcli.Fail(contextcli.CodeUsage, "context changes의 혼합 선택은 지원하지 않습니다", "riido context changes --since SNAPSHOT")
		}
	}
	if raw, present := args["budget_chars"]; present {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return "", contextcli.Fail(contextcli.CodeUsage, "조회 예산이 유효하지 않습니다", "riido context changes --help")
		}
	}
	if len(args) == 0 {
		return "", contextcli.Fail(contextcli.CodeUsage, "context changes에는 since가 필요합니다", "riido context changes --since SNAPSHOT")
	}
	return contextcli.SnapshotRef(args["since"])
}

func changesFetchFailure(q Query, err error) *contextcli.CLIError {
	var failure *FetchError
	if q.View == viewChanges && errors.As(err, &failure) {
		if failure.Status == http.StatusGone {
			return &contextcli.CLIError{Code: contextcli.CodeSnapshotExpired, Message: "이전 스냅샷을 비교할 수 없습니다", NextAction: "riido context get", NoRetry: true}
		}
		if failure.Status == http.StatusBadRequest && strings.Contains(failure.Message, `unknown field "since"`) {
			return contextcli.Fail(contextcli.CodeControlPlaneUnsupported, "Riido 서버가 맥락 변경 비교를 아직 지원하지 않습니다", "riido context get")
		}
	}
	return fetchFailure(err)
}
