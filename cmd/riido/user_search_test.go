package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestUserWorkSearchActualHTTPPages(t *testing.T) {
	calls := 0
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body struct {
			Query     string
			Variables map[string]any
		}
		if r.URL.Path != "/graphql" || r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer fixture-access" || json.NewDecoder(r.Body).Decode(&body) != nil {
			t.Fatal("expected user API only")
		}
		if !strings.Contains(body.Query, "query RiidoCLIWorkSearch") || strings.Contains(body.Query, "mainSearch") || body.Variables["workspaceId"] != "workspace-one" || body.Variables["query"] != "hello" || body.Variables["limit"] != float64(1) {
			t.Fatal("changed fixed query")
		}
		complete := body.Variables["cursor"] == "page-two"
		var next any = "page-two"
		id := "work-one"
		if complete {
			next = nil
			id = "work-two"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"searchWorkspaceWork": map[string]any{"schemaVersion": "riido-work-search.v1", "workspaceId": "workspace-one", "query": "hello", "limit": 1, "complete": complete, "nextCursor": next, "items": []any{map[string]any{"kind": "component", "id": id, "teamId": "team-one", "title": "Hello\x1b[31m\nfake", "componentType": "task", "keyNumber": "1", "key": "RD-1", "updatedAt": "2026-10-31T01:02:03.123456Z", "matchedFields": []string{"title"}}}}}})
	})
	for _, cursor := range []string{"", "page-two"} {
		h := newContextHarness(t, nil)
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
		args := []string{"search", "--workspace", "workspace-one", "--query", " Hello ", "--limit", "1", "--json"}
		if cursor != "" {
			args = append(args, "--cursor", cursor)
		}
		if err := runWork(h.io, args); err != nil {
			t.Fatal(err)
		}
		env := decodeEnvelope(t, h.stdout.Bytes())
		raw, _ := json.Marshal(env.Data)
		var result struct {
			Mode   string
			Live   bool
			Search struct {
				Complete   bool
				NextCursor *string
				Items      []struct{ ID string }
			}
		}
		if json.Unmarshal(raw, &result) != nil || !result.Live || result.Mode != "user" || len(result.Search.Items) != 1 || result.Search.Complete != (cursor != "") || len(h.daemon.requests) != 0 {
			t.Fatalf("bad result %s", raw)
		}
	}
	if calls != 2 {
		t.Fatal("unexpected calls", calls)
	}
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runWork(h.io, []string{"search", "--workspace", "workspace-one", "--query", "hello", "--limit", "1"}); err != nil {
		t.Fatal(err)
	}
	if strings.ContainsRune(h.stdout.String(), '\x1b') || strings.Contains(h.stdout.String(), "\nfake") {
		t.Fatal("unsafe text")
	}
}

func TestUserWorkSearchRejectsBeforeCredentials(t *testing.T) {
	for _, tc := range []struct {
		name string
		vars map[string]string
		args []string
		exit int
	}{
		{"agent", agentVars(), []string{"search", "--query", "hello"}, 8},
		{"missing workspace", nil, []string{"search", "--query", "hello"}, 2},
		{"empty query", nil, []string{"search", "--workspace", "workspace-one", "--query", " "}, 2},
		{"limit", nil, []string{"search", "--workspace", "workspace-one", "--query", "hello", "--limit", "101"}, 2},
		{"page", nil, []string{"search", "--workspace", "workspace-one", "--query", "hello", "--page", "2"}, 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newContextHarness(t, tc.vars)
			opened := false
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { opened = true; return nil, nil }
			if got := exitCodeOf(t, runWork(h.io, append(tc.args, "--json"))); got != tc.exit || opened || len(h.daemon.requests) != 0 {
				t.Fatalf("exit=%d opened=%v", got, opened)
			}
		})
	}
}

func TestUserWorkSearchPermissionAndRateNoPartialData(t *testing.T) {
	for _, tc := range []struct{ status, exit int }{{403, 4}, {429, 7}} {
		t.Run(fmt.Sprint(tc.status), func(t *testing.T) {
			service := commentService(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				fmt.Fprint(w, `{"errors":[{"message":"private-message"}]}`)
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if got := exitCodeOf(t, runWork(h.io, []string{"search", "--workspace", "workspace-one", "--query", "hello", "--json"})); got != tc.exit {
				t.Fatal(got)
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			if env.Data != nil || env.Error == nil || strings.Contains(h.stdout.String(), "private-message") {
				t.Fatal("leaked error/partial data")
			}
		})
	}
}

func TestUserWorkSearchContinuationDoesNotComposeShellFromQuery(t *testing.T) {
	query := "$(touch /tmp/example) `echo unsafe` 'quoted'"
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"searchWorkspaceWork": map[string]any{"schemaVersion": "riido-work-search.v1", "workspaceId": "workspace-one", "query": query, "limit": 1, "complete": false, "nextCursor": "cGFnZS10d28", "items": []any{map[string]any{"kind": "component", "id": "work-one", "teamId": "team-one", "title": "Title", "componentType": "task", "keyNumber": "1", "key": "RD-1", "updatedAt": "2026-10-31T01:02:03.123456Z", "matchedFields": []string{"title"}}}}}})
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runWork(h.io, []string{"search", "--workspace", "workspace-one", "--query", query, "--limit", "1"}); err != nil {
		t.Fatal(err)
	}
	_, hint, ok := strings.Cut(h.stdout.String(), "다음 페이지:")
	if !ok || strings.Contains(hint, "riido work search") || strings.Contains(hint, "$(") || strings.Contains(hint, "`echo") || !strings.Contains(hint, "cGFnZS10d28") {
		t.Fatal("continuation must not compose an executable command from query", hint)
	}
}
