package userapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func meetingListFixture(n int) map[string]any {
	refs := []any{}
	for i := range n {
		refs = append(refs, map[string]any{"id": fmt.Sprintf("meeting-%d", i), "teamId": fmt.Sprintf("team-%d", i), "reasons": []any{map[string]any{"kind": "meeting", "relation": nil, "source": nil}}})
	}
	return map[string]any{"schemaVersion": ReadScopeSchemaVersion, "rootId": "work", "teamId": "team", "status": "available", "selection": map[string]any{"relatedMeetingLimit": 5}, "meetings": refs}
}

func meetingListTestReader(t *testing.T, scope any, mutate func(int, map[string]any) int) (Reader, *int) {
	t.Helper()
	calls := 0
	r := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(ctx context.Context, r *http.Request) (*http.Response, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		calls++
		var body struct {
			Query     string
			Variables map[string]string
		}
		if r.Method != http.MethodPost || r.URL.Path != "/graphql" || json.NewDecoder(r.Body).Decode(&body) != nil {
			t.Error("invalid fixed read")
		}
		for _, field := range []string{"mutation", "body{", "board{", "details{", "works{", "contentPage", "getComponentRelatedMeetings"} {
			if strings.Contains(body.Query, field) {
				t.Error("unexpected content/legacy operation selected", field)
			}
		}
		field, value, status := "getComponentReadScope", scope, http.StatusOK
		if calls == 1 {
			if body.Query != meetingListQuery || len(body.Variables) != 2 || body.Variables["componentId"] != "work" || body.Variables["teamId"] != "team" {
				t.Error("wrong root selection")
			}
		} else {
			index := calls - 2
			id, team := fmt.Sprintf("meeting-%d", index), fmt.Sprintf("team-%d", index)
			if body.Query != meetingSummaryQuery || len(body.Variables) != 2 || body.Variables["meetingId"] != id || body.Variables["teamId"] != team {
				t.Error("meeting identity/order or cross-team authority changed")
			}
			v := map[string]any{"id": id, "teamId": team, "status": "available", "summary": map[string]any{"id": id, "kind": "meeting", "title": "현재 제목"}}
			if mutate != nil {
				status = mutate(index, v)
			}
			field, value = "getMeetingReadScope", v
		}
		raw, _ := json.Marshal(map[string]any{"data": map[string]any{field: value}})
		return response(status, string(raw)), nil
	})}
	return r, &calls
}

func TestMeetingListPreservesAuthorizedSelectionAndCurrentTitles(t *testing.T) {
	for _, n := range []int{0, 1, 5} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			scope := meetingListFixture(n)
			scope["status"] = "partial"
			reader, calls := meetingListTestReader(t, scope, func(_ int, v map[string]any) int {
				v["summary"].(map[string]any)["title"] = ""
				v["status"] = "partial"
				return http.StatusOK
			})
			result, err := reader.Meetings(context.Background(), "work", "team")
			if err != nil || result == nil || result.Meetings == nil || len(result.Meetings) != n || result.ContextStatus != "partial" || result.MaxDepth != 1 || result.SelectionLimit != 5 || *calls != n+1 {
				t.Fatalf("result=%+v calls=%d err=%v", result, *calls, err)
			}
			for i, row := range result.Meetings {
				if row.ID != fmt.Sprintf("meeting-%d", i) || row.TeamID != fmt.Sprintf("team-%d", i) || row.Title != "" || row.Status != "partial" || row.Reasons[0].Kind != "meeting" {
					t.Fatal("authorized order, empty title or partial status lost")
				}
			}
			raw, _ := json.Marshal(result)
			for _, absent := range []string{`"complete"`, `"total"`, `"nextCursor"`, `"body"`} {
				if strings.Contains(string(raw), absent) {
					t.Error("invented completeness or content", absent)
				}
			}
		})
	}
}

func TestMeetingListRejectsMalformedScopeBeforeSummaryReads(t *testing.T) {
	for _, kind := range []string{"nil", "version", "root", "team", "status", "selection", "limit", "overflow", "missing", "null item", "empty id", "empty team", "duplicate", "reason"} {
		t.Run(kind, func(t *testing.T) {
			f := meetingListFixture(1)
			ref := f["meetings"].([]any)[0].(map[string]any)
			switch kind {
			case "nil":
				f = nil
			case "version":
				f["schemaVersion"] = "v2"
			case "root":
				f["rootId"] = "other"
			case "team":
				f["teamId"] = "other"
			case "status":
				f["status"] = "denied"
			case "selection":
				f["selection"] = nil
			case "limit":
				f["selection"].(map[string]any)["relatedMeetingLimit"] = 6
			case "overflow":
				f = meetingListFixture(6)
			case "missing":
				delete(f, "meetings")
			case "null item":
				f["meetings"] = []any{nil}
			case "empty id":
				ref["id"] = ""
			case "empty team":
				ref["teamId"] = ""
			case "duplicate":
				f["meetings"] = []any{ref, ref}
			case "reason":
				ref["reasons"] = []any{map[string]any{"kind": "mention"}}
			}
			r, calls := meetingListTestReader(t, f, nil)
			result, err := r.Meetings(context.Background(), "work", "team")
			if result != nil || !errors.Is(err, ErrInvalidResponse) || *calls != 1 {
				t.Fatal("invalid selection reached summary reads", result, err, *calls)
			}
		})
	}
}

func TestMeetingListDiscardsAllRowsOnMidReadFailure(t *testing.T) {
	for _, kind := range []string{"denied", "auth", "rate", "id", "team", "status", "summary", "summary id", "summary kind", "missing title", "null title"} {
		t.Run(kind, func(t *testing.T) {
			want := ErrInvalidResponse
			r, calls := meetingListTestReader(t, meetingListFixture(3), func(index int, v map[string]any) int {
				if index != 1 {
					return http.StatusOK
				}
				summary := v["summary"].(map[string]any)
				switch kind {
				case "denied":
					want = ErrPermission
					return http.StatusForbidden
				case "auth":
					want = ErrAuthentication
					return http.StatusUnauthorized
				case "rate":
					want = ErrRateLimited
					return http.StatusTooManyRequests
				case "id", "team", "status":
					key := map[string]string{"id": "id", "team": "teamId", "status": "status"}[kind]
					v[key] = "other"
				case "summary":
					v["summary"] = nil
				case "summary id":
					summary["id"] = "other"
				case "summary kind":
					summary["kind"] = "component"
				case "missing title":
					delete(summary, "title")
				case "null title":
					summary["title"] = nil
				}
				return http.StatusOK
			})
			result, err := r.Meetings(context.Background(), "work", "team")
			if result != nil || !errors.Is(err, want) || *calls != 3 {
				t.Fatal("partial result emitted or request continued", result, err, *calls)
			}
		})
	}
}
