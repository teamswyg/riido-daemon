package userapi

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestWorkInspectorPreservesCurrentTaskDetails(t *testing.T) {
	const inspector = `{"id":"work-1","title":"Review","startDate":"2026-10-28T00:00:00.000Z","endDate":null,"estimate":0,"totalPriorityScore":0,"isDone":false,"isArchived":false,"status":{"id":"status-1","statusName":"In progress","statusType":"IN_PROGRESS"},"componentMembers":[{"id":"assignment-1","memberOrder":0,"teamMember":{"id":"member-1","status":"ACTIVE","workspaceProfile":{"nickname":"Reviewer","status":"ACTIVE"}}},{"id":"assignment-2","memberOrder":1,"teamMember":null},{"id":"assignment-3","memberOrder":2,"teamMember":{"id":"member-3","status":"INACTIVE","workspaceProfile":null}}],"componentLabels":[{"id":"assignment-label-1","labelOrder":0,"label":{"id":"label-1","labelName":"CLI"}},{"id":"assignment-label-2","labelOrder":1,"label":null}]}`
	reader := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(_ context.Context, r *http.Request) (*http.Response, error) {
		var request struct {
			Query     string            `json:"query"`
			Variables map[string]string `json:"variables"`
		}
		if json.NewDecoder(r.Body).Decode(&request) != nil || request.Variables["componentId"] != "work-1" || request.Variables["teamId"] != "team-1" {
			t.Fatal("inspector lost explicit task and team binding")
		}
		for _, field := range []string{"startDate", "endDate", "estimate", "totalPriorityScore", "isDone", "isArchived", "statusName", "componentMembers", "memberOrder", "teamMember", "workspaceProfile", "nickname", "componentLabels", "labelOrder", "labelName"} {
			if !strings.Contains(request.Query, field) {
				t.Errorf("inspector query did not request %s", field)
			}
		}
		return response(http.StatusOK, `{"data":{"getComponent":`+inspector+`}}`), nil
	})}
	work, err := reader.Work(t.Context(), "work-1", "team-1")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(work)
	if err != nil {
		t.Fatal(err)
	}
	var want, got map[string]any
	if json.Unmarshal([]byte(inspector), &want) != nil || json.Unmarshal(encoded, &got) != nil {
		t.Fatal("invalid fixture")
	}
	for key, expected := range want {
		if !reflect.DeepEqual(got[key], expected) {
			t.Errorf("inspector field %s changed or disappeared: got %#v want %#v", key, got[key], expected)
		}
	}
}

func TestWorkInspectorKeepsEmptyCollectionsAndNullValues(t *testing.T) {
	reader := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(context.Context, *http.Request) (*http.Response, error) {
		return response(http.StatusOK, `{"data":{"getComponent":{"id":"work-1","title":"Empty","status":null,"estimate":null,"totalPriorityScore":null,"startDate":null,"endDate":null,"componentMembers":[],"componentLabels":[]}}}`), nil
	})}
	work, err := reader.Work(t.Context(), "work-1", "team-1")
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(work)
	var got map[string]any
	if json.Unmarshal(encoded, &got) != nil {
		t.Fatal("invalid output")
	}
	for _, key := range []string{"status", "estimate", "totalPriorityScore", "startDate", "endDate"} {
		value, present := got[key]
		if !present || value != nil {
			t.Errorf("%s must preserve null", key)
		}
	}
	for _, key := range []string{"componentMembers", "componentLabels"} {
		if value, ok := got[key].([]any); !ok || len(value) != 0 {
			t.Errorf("%s must preserve empty array", key)
		}
	}
}
