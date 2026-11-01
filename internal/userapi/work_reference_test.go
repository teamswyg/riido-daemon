package userapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

const resolvedWorkFixture = `{"schemaVersion":"riido-work-reference.v1","workspaceId":"workspace-one","teamId":"team-one","id":"work-one","key":"SC-1","keyNumber":"1","componentType":"task"}`

func referenceReader(t *testing.T, body string) (Reader, *int) {
	t.Helper()
	calls := 0
	return Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(_ context.Context, req *http.Request) (*http.Response, error) {
		calls++
		var q struct {
			Query     string
			Variables WorkReferenceRequest
		}
		if req.URL.String() != "https://api.example/graphql" || req.Method != http.MethodPost || json.NewDecoder(req.Body).Decode(&q) != nil || q.Query != workReferenceQuery || !q.Variables.Valid() {
			t.Error("unfixed or invalid lookup")
		}
		return response(200, body), nil
	})}, &calls
}

func TestWorkReferenceScopeAndResponse(t *testing.T) {
	for _, q := range []WorkReferenceRequest{{Reference: "SC-1", TeamID: "team-one"}, {Reference: "SC-1", WorkspaceID: "workspace-one"}, {Reference: "SC-1", TeamID: "team-one", WorkspaceID: "workspace-one"}} {
		r, calls := referenceReader(t, `{"data":{"resolveCLIWorkReference":`+resolvedWorkFixture+`}}`)
		got, err := r.ResolveWorkReference(t.Context(), q)
		if err != nil || got.ID != "work-one" || *calls != 1 {
			t.Fatal("scoped exact lookup failed", got, err)
		}
	}
	for _, field := range []string{"schemaVersion", "workspaceId", "teamId", "id", "key", "keyNumber", "componentType"} {
		for _, value := range []any{nil, ""} {
			var p map[string]any
			_ = json.Unmarshal([]byte(resolvedWorkFixture), &p)
			p[field] = value
			b, _ := json.Marshal(map[string]any{"data": map[string]any{"resolveCLIWorkReference": p}})
			r, _ := referenceReader(t, string(b))
			if got, err := r.ResolveWorkReference(t.Context(), WorkReferenceRequest{Reference: "SC-1", TeamID: "team-one"}); !errors.Is(err, ErrInvalidResponse) || got.ID != "" {
				t.Fatalf("missing %s accepted: %v", field, err)
			}
		}
	}
	for field, value := range map[string]string{"schemaVersion": "future", "workspaceId": "another-workspace", "teamId": "other-team", "id": "bad id", "key": "SC-10", "keyNumber": "10"} {
		var p map[string]any
		_ = json.Unmarshal([]byte(resolvedWorkFixture), &p)
		p[field] = value
		b, _ := json.Marshal(map[string]any{"data": map[string]any{"resolveCLIWorkReference": p}})
		r, _ := referenceReader(t, string(b))
		if got, err := r.ResolveWorkReference(t.Context(), WorkReferenceRequest{Reference: "SC-1", TeamID: "team-one", WorkspaceID: "workspace-one"}); !errors.Is(err, ErrInvalidResponse) || got.ID != "" {
			t.Fatalf("mismatched %s accepted: %v", field, err)
		}
	}
}

func TestWorkReferenceErrorsAndInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		code string
		want error
	}{{"conflict", ErrAmbiguousWorkReference}, {"not_found", ErrNotFound}, {"permission_denied", ErrPermission}, {"authentication_required", ErrAuthentication}} {
		b, _ := json.Marshal(map[string]any{"data": map[string]any{"resolveCLIWorkReference": json.RawMessage(resolvedWorkFixture)}, "errors": []any{map[string]any{"message": "private-error-canary", "extensions": map[string]any{"code": tc.code}}}})
		r, _ := referenceReader(t, string(b))
		got, err := r.ResolveWorkReference(t.Context(), WorkReferenceRequest{Reference: "SC-1", TeamID: "team-one"})
		if got.ID != "" || !errors.Is(err, tc.want) {
			t.Fatal("partial/error accepted", got, err)
		}
	}
	for _, q := range []WorkReferenceRequest{{Reference: "SC-1"}, {Reference: "work-one", TeamID: "team-one"}, {Reference: "sc-1", TeamID: "team-one"}, {Reference: "SC-1", WorkspaceID: "bad ws"}, {Reference: "SC-1", TeamID: "bad\nteam"}} {
		r, calls := referenceReader(t, "")
		got, err := r.ResolveWorkReference(t.Context(), q)
		if got.ID != "" || !errors.Is(err, ErrInvalidRequest) || *calls != 0 {
			t.Fatal("invalid scope reached API", err)
		}
	}
}
