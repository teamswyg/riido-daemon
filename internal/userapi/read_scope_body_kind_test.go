package userapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestReadBodyUsesOnlyFixedRootForEachResourceKind(t *testing.T) {
	for _, tc := range []struct{ kind, field, id string }{
		{"component", "getComponentReadScopeBody", "componentId"},
		{"backlog", "getBacklogReadScopeBody", "backlogId"},
		{"intake", "getIntakeReadScopeBody", "intakeId"},
		{"meeting", "getMeetingReadScopeBody", "meetingId"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			q := ReadBodyRequest{Kind: tc.kind, ComponentID: "source", TeamID: "team", Part: "document_alternate", Version: "sha256:" + strings.Repeat("a", 64), Limit: 128}
			calls := 0
			r := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(_ context.Context, req *http.Request) (*http.Response, error) {
				calls++
				var input struct {
					Query     string
					Variables map[string]any
				}
				if json.NewDecoder(req.Body).Decode(&input) != nil || !strings.Contains(input.Query, tc.field+"(") || input.Variables[tc.id] != "source" || input.Variables["part"] != "document_alternate" || len(input.Variables) != 6 {
					t.Fatal("wrong typed source binding")
				}
				complete, offset := true, 0
				page := ReadBodyPage{ResourceID: "source", Part: q.Part, Status: "available", Source: "provider_html", Format: "html", Version: q.Version, Offset: &offset, Complete: &complete, Content: "<p>원본</p>"}
				raw, _ := json.Marshal(map[string]any{"data": map[string]any{tc.field: page}})
				return response(200, string(raw)), nil
			})}
			if page, err := r.Body(t.Context(), q); err != nil || page.Content != "<p>원본</p>" || calls != 1 {
				t.Fatalf("page=%+v err=%v calls=%d", page, err, calls)
			}
		})
	}
}

func TestReadBodyRejectsUnknownKindBeforeNetwork(t *testing.T) {
	r := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(context.Context, *http.Request) (*http.Response, error) {
		t.Fatal("invalid kind reached network")
		return nil, nil
	})}
	q := ReadBodyRequest{Kind: "arbitrary-query", ComponentID: "source", TeamID: "team", Part: "document", Version: "sha256:" + strings.Repeat("a", 64), Limit: 128}
	if _, err := r.Body(t.Context(), q); !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("err=%v", err)
	}
}
