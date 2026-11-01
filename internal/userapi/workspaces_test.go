package userapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

const workspacePageFixture = `{"schemaVersion":"riido-cli-workspaces.v1","userId":"user-one","limit":50,"items":[{"id":"ws-one","name":"","isDemo":false}],"nextCursor":null,"complete":true}`

func workspaceReader(t *testing.T, body string, status int, check func(WorkspaceListRequest)) (Reader, *int) {
	t.Helper()
	calls := 0
	return Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(_ context.Context, req *http.Request) (*http.Response, error) {
		calls++
		var input struct {
			Query     string
			Variables WorkspaceListRequest
		}
		if req.Method != http.MethodPost || req.URL.Path != "/graphql" || json.NewDecoder(req.Body).Decode(&input) != nil || input.Query != workspaceListQuery {
			t.Error("not the fixed minimal workspace read")
		}
		if check != nil {
			check(input.Variables)
		}
		return response(status, body), nil
	})}, &calls
}

func TestWorkspaceListPagesAndEmptyFields(t *testing.T) {
	for _, tc := range []struct {
		name, page string
		q          WorkspaceListRequest
		size       int
		complete   bool
	}{
		{"final", workspacePageFixture, WorkspaceListRequest{Limit: 50}, 1, true},
		{"empty", strings.Replace(workspacePageFixture, `[{"id":"ws-one","name":"","isDemo":false}]`, `[]`, 1), WorkspaceListRequest{Limit: 50}, 0, true},
		{"continuation", strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(workspacePageFixture, `"limit":50`, `"limit":1`), `"nextCursor":null`, `"nextCursor":"bmV4dA"`), `"complete":true`, `"complete":false`), WorkspaceListRequest{Limit: 1}, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, calls := workspaceReader(t, `{"data":{"getCLIWorkspaces":`+tc.page+`}}`, 200, func(q WorkspaceListRequest) {
				if q.Limit != tc.q.Limit || q.Cursor != nil {
					t.Error("changed page")
				}
			})
			got, err := r.Workspaces(t.Context(), tc.q)
			if err != nil || got == nil || len(got.Items) != tc.size || got.Complete != tc.complete || *calls != 1 || got.Items == nil {
				t.Fatal("invalid page", got, err)
			}
			if tc.size > 0 && (got.Items[0].Name != "" || got.Items[0].IsDemo) {
				t.Fatal("empty/false changed")
			}
		})
	}
	cursor := "cHJldmlvdXM"
	r, calls := workspaceReader(t, `{"data":{"getCLIWorkspaces":`+workspacePageFixture+`}}`, 200, func(q WorkspaceListRequest) {
		if q.Cursor == nil || *q.Cursor != cursor {
			t.Error("cursor changed")
		}
	})
	if _, err := r.Workspaces(t.Context(), WorkspaceListRequest{Limit: 50, Cursor: &cursor}); err != nil || *calls != 1 {
		t.Fatal(err)
	}
}

func TestWorkspaceListRejectsMalformedAndPartialResults(t *testing.T) {
	for name, mutate := range map[string]func(map[string]any){
		"missing_cursor": func(p map[string]any) { delete(p, "nextCursor") },
		"schema":         func(p map[string]any) { p["schemaVersion"] = "v2" },
		"missing_user":   func(p map[string]any) { delete(p, "userId") },
		"invalid_id":     func(p map[string]any) { p["userId"] = "../user" },
		"limit":          func(p map[string]any) { p["limit"] = 49 },
		"null_items":     func(p map[string]any) { p["items"] = nil },
		"null_item":      func(p map[string]any) { p["items"] = []any{nil} },
		"missing_name":   func(p map[string]any) { delete(p["items"].([]any)[0].(map[string]any), "name") },
		"missing_demo":   func(p map[string]any) { delete(p["items"].([]any)[0].(map[string]any), "isDemo") },
		"duplicate":      func(p map[string]any) { p["items"] = append(p["items"].([]any), p["items"].([]any)[0]) },
		"unordered": func(p map[string]any) {
			p["items"] = append(p["items"].([]any), map[string]any{"id": "A", "name": "x", "isDemo": true})
		},
		"missing_complete": func(p map[string]any) { delete(p, "complete") },
		"incomplete_empty": func(p map[string]any) { p["complete"] = false; p["items"] = []any{}; p["nextCursor"] = "bmV4dA" },
		"complete_cursor":  func(p map[string]any) { p["nextCursor"] = "bmV4dA" },
	} {
		t.Run(name, func(t *testing.T) {
			var p map[string]any
			_ = json.Unmarshal([]byte(workspacePageFixture), &p)
			mutate(p)
			b, _ := json.Marshal(map[string]any{"data": map[string]any{"getCLIWorkspaces": p}})
			r, _ := workspaceReader(t, string(b), 200, nil)
			got, err := r.Workspaces(t.Context(), WorkspaceListRequest{Limit: 50})
			if got != nil || !errors.Is(err, ErrInvalidResponse) {
				t.Fatal("invalid result accepted", err)
			}
		})
	}
	r, _ := workspaceReader(t, `{"data":{"getCLIWorkspaces":`+workspacePageFixture+`},"errors":[{"message":"private canary","extensions":{"code":"permission_denied"}}]}`, 200, nil)
	if got, err := r.Workspaces(t.Context(), WorkspaceListRequest{Limit: 50}); got != nil || !errors.Is(err, ErrPermission) {
		t.Fatal("partial authorization error accepted", got, err)
	}
}

func TestWorkspaceListInputAndHTTPFailures(t *testing.T) {
	for _, limit := range []int{0, -1, 101} {
		r, calls := workspaceReader(t, "", 200, nil)
		if got, err := r.Workspaces(t.Context(), WorkspaceListRequest{Limit: limit}); got != nil || !errors.Is(err, ErrInvalidRequest) || *calls != 0 {
			t.Fatal("bad limit reached network")
		}
	}
	for _, cursor := range []string{"", "bad=", "\n", strings.Repeat("a", 1025)} {
		r, calls := workspaceReader(t, "", 200, nil)
		if got, err := r.Workspaces(t.Context(), WorkspaceListRequest{Limit: 50, Cursor: &cursor}); got != nil || !errors.Is(err, ErrInvalidRequest) || *calls != 0 {
			t.Fatal("bad cursor reached network")
		}
	}
	for status, want := range map[int]error{401: ErrAuthentication, 403: ErrPermission, 429: ErrRateLimited, 503: ErrUnavailable} {
		r, calls := workspaceReader(t, "private canary", status, nil)
		if got, err := r.Workspaces(t.Context(), WorkspaceListRequest{Limit: 50}); got != nil || !errors.Is(err, want) || *calls != 1 {
			t.Fatal("wrong HTTP failure", status, err)
		}
	}
}
