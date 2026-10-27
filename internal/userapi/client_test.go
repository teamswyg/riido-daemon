package userapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type doerFunc func(context.Context, *http.Request) (*http.Response, error)

func (f doerFunc) DoRead(ctx context.Context, r *http.Request) (*http.Response, error) {
	return f(ctx, r)
}

func TestWhoamiUsesReadQueryWithoutDaemon(t *testing.T) {
	reader := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(_ context.Context, r *http.Request) (*http.Response, error) {
		var request struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		if r.Method != http.MethodPost || !strings.Contains(request.Query, "getMyAccountInfo") || strings.Contains(request.Query, "mutation") {
			t.Fatal("unexpected API operation")
		}
		return response(200, `{"data":{"getMyAccountInfo":{"id":"user-1","username":"Fixture","email":"fixture@example.test"}}}`), nil
	})}
	who, err := reader.Whoami(context.Background())
	if err != nil || who.ID != "user-1" {
		t.Fatalf("whoami failed: %v", err)
	}
}

func TestWorkUsesVariablesAndReturnsBody(t *testing.T) {
	reader := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(_ context.Context, r *http.Request) (*http.Response, error) {
		var body struct {
			Query     string            `json:"query"`
			Variables map[string]string `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Variables["componentId"] != "work-1" || body.Variables["teamId"] != "team-1" || !strings.Contains(body.Query, `getDocument:"true"`) {
			t.Fatal("wrong read scope")
		}
		return response(200, `{"data":{"getComponent":{"id":"work-1","title":"Fixture","document":{"HTMLContent":"<p>fixture body</p>"}}}}`), nil
	})}
	work, err := reader.Work(context.Background(), "work-1", "team-1")
	if err != nil || work.Document == nil || work.Document.HTMLContent != "<p>fixture body</p>" {
		t.Fatalf("read failed: %v", err)
	}
}

func TestGraphQLErrorsNeverBecomeSuccessfulReads(t *testing.T) {
	for _, status := range []int{200, 401, 403, 500} {
		reader := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(context.Context, *http.Request) (*http.Response, error) {
			return response(status, `{"data":{"getMyAccountInfo":{"id":"partial"}},"errors":[{"message":"private-server-detail"}]}`), nil
		})}
		_, err := reader.Whoami(context.Background())
		if err == nil || strings.Contains(err.Error(), "private-server-detail") {
			t.Fatalf("status=%d error=%v", status, err)
		}
	}
}

func TestNullOrOversizedBodyFails(t *testing.T) {
	for _, body := range []string{`{"data":{"getMyAccountInfo":null}}`, strings.Repeat("x", maxBodyBytes+1)} {
		reader := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(context.Context, *http.Request) (*http.Response, error) { return response(200, body), nil })}
		if _, err := reader.Whoami(context.Background()); !errors.Is(err, ErrInvalidResponse) {
			t.Fatalf("error=%v", err)
		}
	}
}

func response(code int, body string) *http.Response {
	return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}

func TestGraphQLStableCodesPreserveFailureClassification(t *testing.T) {
	for _, tc := range []struct {
		code string
		want error
	}{
		{"permission_denied", ErrPermission},
		{"not_found", ErrNotFound},
		{"authentication_required", ErrAuthentication},
		{"private-unknown-code", ErrGraphQL},
	} {
		t.Run(tc.code, func(t *testing.T) {
			body, _ := json.Marshal(map[string]any{"data": map[string]any{"getMyAccountInfo": map[string]string{"id": "partial"}}, "errors": []any{map[string]any{"message": "private-server-detail", "extensions": map[string]string{"code": tc.code}}}})
			reader := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(context.Context, *http.Request) (*http.Response, error) {
				return response(http.StatusOK, string(body)), nil
			})}
			result, err := reader.Whoami(context.Background())
			if !errors.Is(err, tc.want) || result.ID != "" {
				t.Fatalf("classification: got %v want %v", err, tc.want)
			}
			if strings.Contains(err.Error(), "private-server-detail") || strings.Contains(err.Error(), "private-unknown-code") {
				t.Fatal("private response escaped")
			}
		})
	}
}
