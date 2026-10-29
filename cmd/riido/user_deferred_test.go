package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestUserDeferredCommentsExplainsHowToReadTheActualPage(t *testing.T) {
	version := "sha256:" + strings.Repeat("a", 64)
	page := map[string]any{"resourceId": "work-one", "resourceKind": "component", "parentCommentId": nil, "limit": 20, "page": 1, "version": version, "status": "deferred", "complete": false, "nextPage": nil, "items": []any{}, "contentPage": map[string]any{"resourceId": "work-one", "part": "comments_page", "source": "riido_comments_page", "format": "json", "version": version, "status": "partial", "offset": 0, "content": "{", "nextOffset": 1, "complete": false, "selector": map[string]any{"page": 1, "limit": 20}}}
	service := commentService(t, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getReadScopeComments": page}})
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runWork(h.io, []string{"comments", "work-one", "--kind", "component", "--team-id", "team-one", "--json"}); err != nil {
		t.Fatal(err)
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	if len(env.Warnings) != 1 || env.Warnings[0].Code != "CONTENT_DEFERRED" || !strings.Contains(env.Warnings[0].Message, "work body") {
		t.Fatal("deferred comments were presented as an empty page without continuation guidance")
	}
	got := env.Data.(map[string]any)["comment_page"].(map[string]any)
	if got["complete"] != false || got["contentPage"] == nil {
		t.Fatal("lost the deferred page contract")
	}
}

func TestUserDeferredContextExplainsMetadataContinuation(t *testing.T) {
	var scope map[string]any
	if err := json.Unmarshal([]byte(userReadScopeFixture), &scope); err != nil {
		t.Fatal(err)
	}
	w := scope["works"].([]any)[0].(map[string]any)
	w["details"] = nil
	w["summary"] = map[string]any{"id": "work-one", "kind": "component", "title": "큰 상세"}
	w["contentPage"] = map[string]any{"resourceId": "work-one", "part": "metadata", "source": "riido_metadata", "format": "json", "version": "sha256:" + strings.Repeat("a", 64), "status": "partial", "offset": 0, "content": "{", "nextOffset": 1, "complete": false, "selector": map[string]any{}}
	service := commentService(t, func(out http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(out).Encode(map[string]any{"data": map[string]any{"getComponentReadScope": scope}})
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runContext(h.io, []string{"get", "--work", "work-one", "--team-id", "team-one", "--json"}); err != nil {
		t.Fatal(err)
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	found := false
	for _, w := range env.Warnings {
		if w.Code == "CONTENT_DEFERRED" {
			found = true
		}
	}
	if !found {
		t.Fatal("deferred metadata has no usable continuation guidance")
	}
}
