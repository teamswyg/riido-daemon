package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestUserBodyReconstructsLargeTypedPageAcrossCLIInvocations(t *testing.T) {
	original, err := json.Marshal(map[string]any{"items": []any{map[string]any{"id": "comment-one", "body": strings.Repeat("큰 댓글🙂", 12000)}}, "complete": true, "nextPage": nil})
	if err != nil {
		t.Fatal(err)
	}
	version := "sha256:" + strings.Repeat("a", 64)
	selector := map[string]any{"page": float64(2), "limit": float64(3), "parentCommentId": "parent-one"}
	calls := 0
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		var req struct {
			Query     string
			Variables map[string]any
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil || !strings.Contains(req.Query, "getIntakeReadScopeBody") || req.Variables["version"] != version {
			t.Fatal("wrong source or version")
		}
		offset := int(req.Variables["offset"].(float64))
		end := offset + int(req.Variables["limit"].(float64))
		if end > len(original) {
			end = len(original)
		}
		for end < len(original) && !utf8.RuneStart(original[end]) {
			end--
		}
		var next any
		status := "available"
		if end < len(original) {
			next = end
			status = "partial"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getIntakeReadScopeBody": map[string]any{"resourceId": "work-one", "part": "comments_page", "source": "riido_comments_page", "format": "json", "version": version, "offset": offset, "content": string(original[offset:end]), "nextOffset": next, "complete": end == len(original), "status": status, "selector": selector}}})
	})
	var reconstructed strings.Builder
	offset := 0
	done := false
	for attempt := 0; attempt < 32; attempt++ {
		h := newContextHarness(t, nil)
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
		args := []string{"body", "work-one", "--kind", "intake", "--team-id", "team-one", "--part", "comments_page", "--version", version, "--page", "2", "--comments-limit", "3", "--comment-id", "parent-one", "--offset", strconv.Itoa(offset), "--limit", "16384", "--json"}
		if err := runWork(h.io, args); err != nil {
			t.Fatal(err)
		}
		page := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)["body"].(map[string]any)
		reconstructed.WriteString(page["content"].(string))
		if len(h.daemon.requests) != 0 {
			t.Fatal("standalone page contacted daemon")
		}
		if page["complete"] == true {
			done = true
			break
		}
		offset = int(page["nextOffset"].(float64))
	}
	if !done || calls < 5 || reconstructed.String() != string(original) {
		t.Fatal("large logical page lost UTF-8 content or continuation")
	}
	var decoded map[string]any
	if json.Unmarshal([]byte(reconstructed.String()), &decoded) != nil || decoded["complete"] != true {
		t.Fatal("reconstructed logical page is unusable")
	}
}
