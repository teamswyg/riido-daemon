package standalonecli

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

// The hook runs on the reader's goroutine after HTTP response parsing, not in
// the server goroutine. Replacing the fixture record deliberately bypasses the
// store lock to exercise the SessionID fence as well as cooperative locking.
type compoundResponseBody struct {
	io.ReadCloser
	afterClose func()
}

func (b *compoundResponseBody) Close() error {
	err := b.ReadCloser.Close()
	b.afterClose()
	return err
}

func compoundContextFixture() userapi.ReadScope {
	complete, offset := true, 0
	return userapi.ReadScope{
		SchemaVersion: userapi.ReadScopeSchemaVersion, RootID: "work", TeamID: "team", Status: "available",
		Selection: &userapi.ReadScopeSelection{RelatedWorkLimit: 5, RelatedMeetingLimit: 5, ChildrenLimit: 10},
		Meetings:  []*userapi.ReadMeetingReference{{ID: "meeting", TeamID: "team", Reasons: []*userapi.ReadScopeReason{{Kind: "meeting"}}}},
		Works: []*userapi.ReadScopeWork{{
			ID: "work", Kind: "component", TeamID: "team", Reasons: []*userapi.ReadScopeReason{{Kind: "root"}},
			Details: json.RawMessage(`{"id":"work"}`), PartsStatus: []*userapi.ReadPartStatus{},
			Body:       &userapi.ReadBodyPage{ResourceID: "work", Part: "document", Status: "empty", Source: "provider_json", Format: "json", Version: "sha256:" + strings.Repeat("a", 64), Offset: &offset, Complete: &complete},
			Comments:   &userapi.ReadScopeCommentPage{ResourceID: "work", ResourceKind: "component", Limit: 20, Version: "sha256:" + strings.Repeat("a", 64), Status: "available", Page: 1, Complete: &complete, Items: []json.RawMessage{}},
			Priorities: []json.RawMessage{}, Checklists: []json.RawMessage{}, Attachments: []json.RawMessage{}, Sprints: []json.RawMessage{}, Links: []json.RawMessage{}, GitHub: []json.RawMessage{}, Repositories: []json.RawMessage{},
		}},
	}
}

func compoundMeetingFixture() userapi.ReadMeeting {
	complete, offset := true, 0
	body := &userapi.ReadBodyPage{ResourceID: "meeting", Part: "document", Status: "empty", Source: "provider_json", Format: "json", Version: "sha256:" + strings.Repeat("a", 64), Offset: &offset, Complete: &complete}
	description := *body
	description.Part = "description"
	return userapi.ReadMeeting{ID: "meeting", TeamID: "team", Status: "available", Details: json.RawMessage(`{"id":"meeting","meetingParticipants":[],"meetingAttendees":[]}`), Labels: []json.RawMessage{}, PartsStatus: []*userapi.ReadPartStatus{}, Body: body, DescriptionBody: &description, Board: &userapi.ReadMeetingBoard{Version: "board-version", Offset: &offset, Complete: &complete, Groups: []json.RawMessage{}, Cards: []json.RawMessage{}}}
}

func compoundFixture(t *testing.T, kind string, hook func(*Service, *testStore, int), refresh bool) (*Service, *testStore, *atomic.Int32, *atomic.Int32) {
	t.Helper()
	var calls, refreshes atomic.Int32
	store := &testStore{present: true, credential: userauth.Credential{SessionID: "fixture-session", AccessToken: "fixture-access", RefreshToken: "fixture-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
	metadata := `{"details":{"id":"work"},"links":[],"attachments":[],"figma":null,"repositories":[],"github":[]}`
	canonical, _ := json.Marshal([]string{"riido-read-part.v1", "component", "work", "team", "metadata", "{}", "riido_metadata", "json", metadata})
	version := fmt.Sprintf("sha256:%x", sha256.Sum256(canonical))
	page := func(offset, end int) map[string]any {
		var next any
		status := "available"
		if end < len(metadata) {
			next, status = end, "partial"
		}
		return map[string]any{"selector": map[string]any{}, "resourceId": "work", "part": "metadata", "status": status, "source": "riido_metadata", "format": "json", "version": version, "offset": offset, "content": metadata[offset:end], "nextOffset": next, "complete": end == len(metadata)}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			refreshes.Add(1)
			fmt.Fprint(w, `{"access_token":"refreshed-access","refresh_token":"refreshed-refresh","token_type":"Bearer","expires_in":3600,"refresh_recovery":"v1"}`)
			return
		}
		call := calls.Add(1)
		if refresh && call == 2 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		want := "Bearer fixture-access"
		if refreshes.Load() > 0 {
			want = "Bearer refreshed-access"
		}
		if r.Header.Get("Authorization") != want {
			t.Error("another account reached the API")
		}
		var request struct {
			Query     string
			Variables map[string]json.RawMessage
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		field, value := "getComponentReadScope", any(compoundContextFixture())
		switch {
		case kind == "context":
			if strings.Contains(request.Query, "getMeetingReadScope") {
				field, value = "getMeetingReadScope", compoundMeetingFixture()
			}
		case strings.Contains(request.Query, "getComponentReadScopeBody"):
			if string(request.Variables["offset"]) != "32" || string(request.Variables["version"]) != fmt.Sprintf("%q", version) {
				t.Error("metadata continuation lost original offset/version")
			}
			field, value = "getComponentReadScopeBody", page(32, len(metadata))
		default:
			names := []string{"links", "attachments", "figma"}
			if kind == "repositories" {
				names = []string{"repositories", "github"}
			}
			parts := []any{}
			for _, name := range names {
				parts = append(parts, map[string]any{"name": name, "status": "deferred", "complete": false})
			}
			value = map[string]any{"schemaVersion": userapi.ReadScopeSchemaVersion, "rootId": "work", "teamId": "team", "status": "partial", "works": []any{map[string]any{"id": "work", "kind": "component", "teamId": "team", "reasons": []any{map[string]any{"kind": "root"}}, "partsStatus": parts, "contentPage": page(0, 32)}}}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{field: value}})
	}))
	t.Cleanup(server.Close)
	var service *Service
	client := server.Client()
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	client.Transport = protectedRoundTripper(func(r *http.Request) (*http.Response, error) {
		res, err := transport.RoundTrip(r)
		if err == nil && r.URL.Path == "/graphql" && res.StatusCode == http.StatusOK && hook != nil {
			call := int(calls.Load())
			res.Body = &compoundResponseBody{ReadCloser: res.Body, afterClose: func() { hook(service, store, call) }}
		}
		return res, err
	})
	var err error
	service, err = New(Config{APIOrigin: server.URL, Store: store, HTTPClient: client})
	if err != nil {
		t.Fatal(err)
	}
	return service, store, &calls, &refreshes
}

// Normalize only the failure result for assertions: Context returns a value,
// whereas the two metadata commands return pointers.
func compoundRead(ctx context.Context, service *Service, kind string) (bool, error) {
	switch kind {
	case "context":
		result, err := service.Context(ctx, "work", "team")
		return result.RootID != "" || result.Works != nil || result.Meetings != nil, err
	case "resources":
		result, err := service.Resources(ctx, "work", "team")
		return result != nil, err
	default:
		result, err := service.Repositories(ctx, "work", "team")
		return result != nil, err
	}
}
