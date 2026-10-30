package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func resourceHTTPFixture(t *testing.T, deferred bool, continuationStatus int) (*standalonecli.Service, *[]string) {
	t.Helper()
	service, calls, _ := resourceHTTPFixtureWithPadding(t, deferred, continuationStatus, strings.Repeat("한", 45000))
	return service, calls
}

func resourceHTTPFixtureWithPadding(t *testing.T, deferred bool, continuationStatus int, padding string) (*standalonecli.Service, *[]string, string) {
	t.Helper()
	metadata := map[string]any{"details": map[string]any{"id": "work-one"}, "links": []any{map[string]any{"id": "link-one", "title": "제목\x1b[31m\n가짜 줄", "url": "https://external.example/private"}}, "attachments": []any{}, "figma": nil, "unselected": padding}
	return metadataHTTPFixture(t, "resources", []string{"links", "attachments", "figma"}, metadata, deferred, continuationStatus)
}

func metadataHTTPFixture(t *testing.T, operation string, names []string, metadata map[string]any, deferred bool, continuationStatus int) (*standalonecli.Service, *[]string, string) {
	t.Helper()
	raw, _ := json.Marshal(metadata)
	canonical, _ := json.Marshal([]string{"riido-read-part.v1", "component", "work-one", "team-one", "metadata", "{}", "riido_metadata", "json", string(raw)})
	version := fmt.Sprintf("sha256:%x", sha256.Sum256(canonical))
	page := func(offset, limit int) map[string]any {
		end := min(offset+limit, len(raw))
		for end < len(raw) && !utf8.RuneStart(raw[end]) {
			end--
		}
		status := "partial"
		var next any = end
		if end == len(raw) {
			status = "available"
			next = nil
		}
		return map[string]any{"selector": map[string]any{}, "resourceId": "work-one", "part": "metadata", "status": status, "source": "riido_metadata", "format": "json", "version": version, "offset": offset, "content": string(raw[offset:end]), "nextOffset": next, "complete": end == len(raw)}
	}
	calls := []string{}
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query     string
			Variables map[string]json.RawMessage
		}
		if r.URL.Path != "/graphql" || r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer fixture-access" || json.NewDecoder(r.Body).Decode(&body) != nil {
			t.Error("expected authenticated GraphQL only")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if string(body.Variables["componentId"]) != `"work-one"` || string(body.Variables["teamId"]) != `"team-one"` {
			t.Error("wrong root scope")
		}
		if strings.Contains(body.Query, "getComponentReadScopeBody") {
			calls = append(calls, "metadata")
			if continuationStatus != 200 {
				w.WriteHeader(continuationStatus)
				fmt.Fprint(w, `{"errors":[{"message":"private-error"}]}`)
				return
			}
			if string(body.Variables["part"]) != `"metadata"` || string(body.Variables["version"]) != fmt.Sprintf("%q", version) || string(body.Variables["selector"]) != `{}` {
				t.Error("version/selector changed")
			}
			var offset, limit int
			_ = json.Unmarshal(body.Variables["offset"], &offset)
			_ = json.Unmarshal(body.Variables["limit"], &limit)
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponentReadScopeBody": page(offset, limit)}})
			return
		}
		calls = append(calls, operation)
		queryName := "query RiidoCLIResources"
		if operation == "repositories" {
			queryName = "query RiidoCLIRepositories"
		}
		if !strings.Contains(body.Query, queryName) || strings.Contains(body.Query, "meetings") || strings.Contains(body.Query, "mutation") {
			t.Error("unexpected selection")
		}
		if operation == "repositories" && (strings.Contains(body.Query, "links{") || strings.Contains(body.Query, "attachments{") || strings.Contains(body.Query, "figma{") || strings.Contains(body.Query, "comments{") || strings.Contains(body.Query, "aiSummary")) {
			t.Error("unexpected repository content selection")
		}
		if operation == "resources" && strings.Contains(body.Query, "github{") {
			t.Error("unexpected resource content selection")
		}
		work := map[string]any{"id": "work-one", "kind": "component", "teamId": "team-one", "reasons": []any{map[string]any{"kind": "root"}}}
		for _, name := range names {
			work[name] = metadata[name]
		}
		parts := []any{}
		for _, name := range names {
			status := "available"
			if deferred {
				status = "deferred"
			}
			parts = append(parts, map[string]any{"name": name, "status": status, "complete": !deferred})
		}
		work["partsStatus"] = parts
		if deferred {
			work["contentPage"] = page(0, 1024)
			for _, name := range names {
				work[name] = []any{}
			}
			if operation == "resources" {
				work["figma"] = nil
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponentReadScope": map[string]any{"schemaVersion": "riido-read-scope.v1", "rootId": "work-one", "teamId": "team-one", "status": "available", "works": []any{work}}}})
	})
	return service, &calls, version
}

func TestUserResourcesActualHTTPInlineAndDeferred(t *testing.T) {
	for _, deferred := range []bool{false, true} {
		t.Run(fmt.Sprint(deferred), func(t *testing.T) {
			service, calls := resourceHTTPFixture(t, deferred, 200)
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if err := runResource(h.io, []string{"list", "--work", "work-one", "--team-id", "team-one", "--json"}); err != nil {
				t.Fatal(err)
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			raw, _ := json.Marshal(env.Data)
			var result struct {
				Mode      string
				Live      bool
				Resources struct {
					Complete        bool
					Links           []struct{ Title string }
					MetadataVersion *string
				}
			}
			if json.Unmarshal(raw, &result) != nil || env.Error != nil || result.Mode != "user" || !result.Live || !result.Resources.Complete || len(result.Resources.Links) != 1 || result.Resources.Links[0].Title != "제목\x1b[31m\n가짜 줄" {
				t.Fatalf("bad output: %s", raw)
			}
			want := []string{"resources"}
			if deferred {
				want = append(want, "metadata", "metadata", "metadata")
			}
			if !reflect.DeepEqual(*calls, want) || len(h.daemon.requests) != 0 {
				t.Fatalf("calls=%v IPC=%d", *calls, len(h.daemon.requests))
			}
			if (result.Resources.MetadataVersion != nil) != deferred || strings.Contains(h.stdout.String(), "fixture-access") {
				t.Fatal("version or secret output")
			}
		})
	}
}

func TestUserResourcesTextEscapesRemoteValues(t *testing.T) {
	service, _ := resourceHTTPFixture(t, false, 200)
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runResource(h.io, []string{"list", "--work", "work-one", "--team-id", "team-one"}); err != nil {
		t.Fatal(err)
	}
	if strings.ContainsRune(h.stdout.String(), '\x1b') || strings.Contains(h.stdout.String(), "\n가짜 줄") || !strings.Contains(h.stdout.String(), "external.example") {
		t.Fatal("unsafe/lost text")
	}
}

func TestUserResourcesContinuationPermissionAndRateErrors(t *testing.T) {
	for _, tc := range []struct{ status, exit int }{{403, 4}, {429, 7}} {
		t.Run(fmt.Sprint(tc.status), func(t *testing.T) {
			service, calls := resourceHTTPFixture(t, true, tc.status)
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if got := exitCodeOf(t, runResource(h.io, []string{"list", "--work", "work-one", "--team-id", "team-one", "--json"})); got != tc.exit {
				t.Fatalf("exit=%d", got)
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			if env.Data != nil || env.Error == nil || len(*calls) != 2 || strings.Contains(h.stdout.String(), "private-error") {
				t.Fatal("partial output or retry escaped")
			}
		})
	}
}

func TestUserResourcesModeAndArgumentsBeforeCredentials(t *testing.T) {
	for _, tc := range []struct {
		name string
		vars map[string]string
		args []string
		exit int
	}{
		{"agent", agentVars(), []string{"list", "--work", "RIID-1"}, 8},
		{"missing", nil, []string{"list"}, 2},
		{"key", nil, []string{"list", "--work", "RIID-1", "--team-id", "team-one"}, 8},
		{"cursor", nil, []string{"list", "--work", "work-one", "--team-id", "team-one", "--cursor", "x"}, 8},
		{"unknown", nil, []string{"get"}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newContextHarness(t, tc.vars)
			opened := false
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { opened = true; return nil, nil }
			if got := exitCodeOf(t, runResource(h.io, append(tc.args, "--json"))); got != tc.exit || opened || len(h.daemon.requests) != 0 {
				t.Fatalf("exit=%d opened=%v", got, opened)
			}
		})
	}
	h := newContextHarness(t, nil)
	if err := runResource(h.io, []string{"list", "--help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.stderr.String(), "riido resource list --work <ID> --team-id <TEAM-ID>") {
		t.Fatal("missing help")
	}
	found := false
	for _, c := range contextcli.Capabilities("fixture").Commands {
		if c.Name == "resource list" {
			found = true
			if c.Status != contextcli.CommandAvailable || !reflect.DeepEqual(c.Modes, []contextcli.Mode{contextcli.ModeUser}) {
				t.Fatal("wrong modes")
			}
		}
	}
	if !found {
		t.Fatal("missing capability")
	}
}

func TestUserResourcesLimitPreservesExecutableOriginalVersion(t *testing.T) {
	service, calls, version := resourceHTTPFixtureWithPadding(t, true, http.StatusOK, strings.Repeat("x", 16<<20))
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	err := runResource(h.io, []string{"list", "--work", "work-one", "--team-id", "team-one", "--json"})
	if got := exitCodeOf(t, err); got != 9 {
		t.Fatalf("exit=%d", got)
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	want := fmt.Sprintf("riido work body %q --team-id %q --part metadata --version %q --offset 0 --limit 65536 --json", "work-one", "team-one", version)
	if env.Error == nil || env.Error.NextAction != want || env.Data != nil {
		t.Fatalf("manual paging lost original version: %+v", env.Error)
	}
	// Execute the advertised continuation using the same authenticated HTTP API.
	continuation := newContextHarness(t, nil)
	continuation.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	before := len(*calls)
	if err := runWork(continuation.io, []string{"body", "work-one", "--team-id", "team-one", "--part", "metadata", "--version", version, "--offset", "0", "--limit", "65536", "--json"}); err != nil {
		t.Fatal(err)
	}
	pageEnvelope := decodeEnvelope(t, continuation.stdout.Bytes())
	raw, _ := json.Marshal(pageEnvelope.Data)
	var data struct {
		Body struct {
			Version    string
			Offset     int
			NextOffset *int
			Complete   bool
		}
	}
	if json.Unmarshal(raw, &data) != nil || data.Body.Version != version || data.Body.Offset != 0 || data.Body.NextOffset == nil || data.Body.Complete || len(*calls) != before+1 {
		t.Fatal("advertised work body continuation failed")
	}
	for i, call := range *calls {
		if (i == 0 && call != "resources") || (i > 0 && call != "metadata") {
			t.Fatal("unrelated read blocks fallback")
		}
	}
}
