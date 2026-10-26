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
		View: view, Target: strings.TrimSpace(req.Args["work"]), Cursor: strings.TrimSpace(req.Args["cursor"]), RequestID: requestID,
	})
	if err != nil {
		return contextcli.FailureEnvelope(requestID, fetchFailure(err))
	}
	env := contextcli.NewEnvelope(requestID, result.Data)
	if view == viewComments {
		env.Page = commentPage(result.Data)
	}
	status := contextcli.SourceAvailable
	if result.Truncated {
		status = contextcli.SourcePartial
	}
	fetchedAt := result.CapturedAt.UTC().Format(time.RFC3339)
	for _, source := range result.Sources {
		env.Sources = append(env.Sources, contextcli.Source{ID: source.ID, Kind: source.Kind, Version: source.Version, FetchedAt: fetchedAt, Status: status, Truncated: result.Truncated})
	}
	env.Warnings = append(env.Warnings, contextcli.Warning{
		Code:    "ASSIGNMENT_SNAPSHOT",
		Message: "배정 시점(" + fetchedAt + ")에 저장된 맥락입니다. 이후 변경은 반영되지 않습니다",
	})
	if result.Truncated {
		env.Warnings = append(env.Warnings, contextcli.Warning{Code: "SNAPSHOT_TRUNCATED", Message: "본문이 길어 일부만 저장됐습니다"})
	}
	return env
}

func fetchFailure(err error) *contextcli.CLIError {
	var fetchErr *FetchError
	if !errors.As(err, &fetchErr) {
		return contextcli.Fail(contextcli.CodeUnavailable, "Riido 서버에 연결하지 못했습니다", "잠시 후 다시 시도하세요")
	}
	switch {
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
