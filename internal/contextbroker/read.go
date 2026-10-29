package contextbroker

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

const fetchTimeout = 15 * time.Second

// read answers a Riido data command from the assignment context snapshot.
func (b *Broker) read(req contextcli.ContextReadRequest, requestID string, binding Binding) contextcli.Envelope {
	view, ok := viewForCommand(req.Command)
	if !ok {
		return contextcli.FailureEnvelope(requestID, contextcli.Fail(contextcli.CodeControlPlaneUnsupported,
			"이 명령은 아직 지원하지 않습니다 (읽기 범위 결정 전)", "riido context get 또는 riido work get 을 사용하세요"))
	}
	if b.Fetcher == nil || binding.AgentID == "" || binding.AssignmentID == "" || binding.LeaseToken == "" {
		return contextcli.FailureEnvelope(requestID, contextcli.Fail(contextcli.CodeControlPlaneUnsupported,
			"이 실행은 Riido 배정 맥락을 조회할 수 없습니다", "프롬프트로 받은 맥락을 사용하세요"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()
	result, err := b.Fetcher.FetchAssignmentContext(ctx, Query{
		AgentID: binding.AgentID, AssignmentID: binding.AssignmentID, LeaseToken: binding.LeaseToken,
		View: view, Target: queryTarget(req.Args), Cursor: strings.TrimSpace(req.Args["cursor"]), RequestID: requestID,
	})
	if err != nil {
		return contextcli.FailureEnvelope(requestID, fetchFailure(err))
	}
	budget := budgetChars(req.Args)
	preservePage := capturedReadBudget(view, result.Data, budget)
	data, budgetCut, budgetFits := result.Data, false, true
	if !preservePage {
		data, budgetCut, budgetFits = applyBudget(result.Data, budget)
	}
	truncated := result.Truncated || budgetCut
	env := contextcli.NewEnvelope(requestID, data)
	if (view == viewComments && !capturedReadResponse(view, result.Data)) || view == viewAgents || view == viewContent {
		env.Page = cursorPage(view, result.Data)
	}
	status := contextcli.SourceAvailable
	if truncated {
		status = contextcli.SourcePartial
	}
	fetchedAt := result.CapturedAt.UTC().Format(time.RFC3339)
	for _, source := range result.Sources {
		env.Sources = append(env.Sources, contextcli.Source{ID: source.ID, Kind: source.Kind, Version: source.Version, FetchedAt: fetchedAt, Status: status, Truncated: truncated})
	}
	env.Warnings = append(env.Warnings, freshnessWarning(result.Live, fetchedAt))
	if result.Truncated {
		env.Warnings = append(env.Warnings, contextcli.Warning{Code: "SNAPSHOT_TRUNCATED", Message: "일부 자료가 완전히 저장되지 않았습니다. 각 내용의 상태와 저장된 다음 페이지를 확인하세요"})
	}
	if preservePage {
		env.Warnings = append(env.Warnings, contextcli.Warning{Code: "BUDGET_PRESERVED_PAGE", Message: "저장된 페이지의 버전과 다음 위치를 보존하기 위해 --budget을 초과한 원문을 그대로 반환했습니다. 더 좁은 조회에는 응답의 부분 자료 커서를 사용하세요"})
	}
	return withBudgetWarnings(env, budgetCut, budgetFits)
}

func fetchFailure(err error) *contextcli.CLIError {
	var fetchErr *FetchError
	if !errors.As(err, &fetchErr) {
		return contextcli.Fail(contextcli.CodeUnavailable, "Riido 서버에 연결하지 못했습니다", "잠시 후 다시 시도하세요")
	}
	switch {
	case fetchErr.Status == http.StatusBadRequest:
		return contextcli.Fail(contextcli.CodeUsage, "조회 입력 또는 페이지 커서가 유효하지 않습니다", "새 맥락의 부분 자료는 riido context part <SOURCE-ID> --cursor <CURSOR>로 이어 읽으세요")
	case fetchErr.Status == http.StatusTooManyRequests:
		return &contextcli.CLIError{
			Code: contextcli.CodeRateLimited, NoRetry: true,
			Message:    "이 실행에서 조회할 수 있는 횟수를 모두 사용했습니다",
			NextAction: "같은 실행에서 재시도하지 말고, 사용자에게 새 요청으로 새 실행을 시작하도록 안내하세요",
		}
	case fetchErr.Status == http.StatusServiceUnavailable && strings.Contains(fetchErr.Message, "not configured"),
		fetchErr.Status == http.StatusNotFound && fetchErr.Message == "not found":
		return contextcli.Fail(contextcli.CodeControlPlaneUnsupported, "Riido 서버가 배정 맥락 조회를 아직 지원하지 않습니다", "프롬프트로 받은 맥락을 사용하세요")
	case fetchErr.Status == http.StatusForbidden && strings.Contains(fetchErr.Message, "outside"):
		return contextcli.Fail(contextcli.CodeOutOfRunScope, "이 실행의 배정 맥락에 없는 작업입니다", "프롬프트에 나온 작업만 조회할 수 있습니다")
	case fetchErr.Status == http.StatusUnauthorized, fetchErr.Status == http.StatusForbidden:
		return contextcli.Fail(contextcli.CodePermissionDenied, "Riido 서버가 이 실행의 조회를 거부했습니다", "")
	case fetchErr.Status == http.StatusConflict:
		return contextcli.Fail(contextcli.CodeLeaseLost, "실행이 끝났거나 배정이 바뀌었습니다", "")
	case fetchErr.Status == http.StatusNotFound:
		return contextcli.Fail(contextcli.CodeNotFound, "이 배정의 저장된 맥락이 없습니다", "프롬프트로 받은 맥락을 사용하세요")
	default:
		return contextcli.Fail(contextcli.CodeUnavailable, "Riido 서버가 배정 맥락을 돌려주지 못했습니다", "잠시 후 다시 시도하세요")
	}
}
