package userapi

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func repositoriesFixture() map[string]any {
	f := resourceScopeFixture()
	root := f["works"].([]any)[0].(map[string]any)
	root["repositories"] = []any{map[string]any{"id": "repo-one", "fullName": "team/first", "isPrivate": false, "repositoryUrl": "https://github.example/team/first", "source": "connected"}, map[string]any{"id": "repo-two", "fullName": "team/private", "isPrivate": true, "repositoryUrl": "https://github.example/team/private", "source": "linked"}}
	root["github"] = []any{map[string]any{"id": "pr-one", "title": "", "number": 0, "state": "open", "htmlUrl": "https://github.example/team/first/pull/1", "updatedAt": nil, "commentsCount": 0, "aiShortSummary": ""}}
	root["partsStatus"] = []any{map[string]any{"name": "repositories", "status": "available", "complete": true}, map[string]any{"name": "github", "status": "available", "complete": true}}
	return f
}

func TestRepositoriesPreserveStoredMetadataWithoutExternalAuthentication(t *testing.T) {
	got, err := relationReader(resourceFixtureJSON(repositoriesFixture())).Repositories(t.Context(), "work-one", "team-one")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Complete || len(got.Repositories) != 2 || *got.Repositories[0].IsPrivate || !*got.Repositories[1].IsPrivate || *got.Repositories[0].Source != "connected" || len(got.PullRequests) != 1 || *got.PullRequests[0].Title != "" || got.PullRequests[0].UpdatedAt != nil || *got.PullRequests[0].Number != 0 {
		t.Fatalf("metadata changed: %+v", got)
	}
	if got.WorkID != "work-one" || got.TeamID != "team-one" || got.MetadataVersion != nil {
		t.Fatal("wrong projection identity")
	}
}

func TestRepositoriesRejectMalformedRowsAndAvailability(t *testing.T) {
	for name, change := range map[string]func(map[string]any){
		"missing-repos":          func(r map[string]any) { delete(r, "repositories") },
		"null-prs":               func(r map[string]any) { r["github"] = nil },
		"missing-private":        func(r map[string]any) { delete(r["repositories"].([]any)[0].(map[string]any), "isPrivate") },
		"null-source":            func(r map[string]any) { r["repositories"].([]any)[0].(map[string]any)["source"] = nil },
		"duplicate-repo":         func(r map[string]any) { r["repositories"].([]any)[1] = r["repositories"].([]any)[0] },
		"null-title":             func(r map[string]any) { r["github"].([]any)[0].(map[string]any)["title"] = nil },
		"missing-number":         func(r map[string]any) { delete(r["github"].([]any)[0].(map[string]any), "number") },
		"missing-comments-count": func(r map[string]any) { delete(r["github"].([]any)[0].(map[string]any), "commentsCount") },
		"duplicate-pr":           func(r map[string]any) { r["github"] = append(r["github"].([]any), r["github"].([]any)[0]) },
		"missing-status":         func(r map[string]any) { r["partsStatus"] = r["partsStatus"].([]any)[:1] },
		"contradictory-status": func(r map[string]any) {
			s := r["partsStatus"].([]any)[0].(map[string]any)
			s["status"] = "unavailable"
			s["complete"] = false
		},
	} {
		t.Run(name, func(t *testing.T) {
			f := repositoriesFixture()
			change(f["works"].([]any)[0].(map[string]any))
			got, err := relationReader(resourceFixtureJSON(f)).Repositories(t.Context(), "work-one", "team-one")
			if got != nil || !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("accepted %+v %v", got, err)
			}
		})
	}
	f := repositoriesFixture()
	root := f["works"].([]any)[0].(map[string]any)
	root["repositories"] = []any{}
	root["github"] = []any{}
	s := root["partsStatus"].([]any)[0].(map[string]any)
	s["status"] = "unavailable"
	s["complete"] = false
	got, err := relationReader(resourceFixtureJSON(f)).Repositories(t.Context(), "work-one", "team-one")
	if err != nil || got.Complete || got.Repositories == nil || got.PullRequests == nil {
		t.Fatalf("unknown availability lost: %+v %v", got, err)
	}
}

func TestRepositoriesFixedSelectionAndDeferredMetadata(t *testing.T) {
	f := repositoriesFixture()
	root := f["works"].([]any)[0].(map[string]any)
	root["details"] = map[string]any{"id": "work-one"}
	root["ignoredBody"] = strings.Repeat("x", 140000)
	raw, _ := json.Marshal(root)
	canonical, _ := json.Marshal([]string{"riido-read-part.v1", "component", "work-one", "team-one", "metadata", "{}", "riido_metadata", "json", string(raw)})
	version := fmt.Sprintf("sha256:%x", sha256.Sum256(canonical))
	page := func(offset, limit int) map[string]any {
		end := min(offset+limit, len(raw))
		status := "partial"
		var next any = end
		if end == len(raw) {
			status = "available"
			next = nil
		}
		return map[string]any{"selector": map[string]any{}, "resourceId": "work-one", "part": "metadata", "status": status, "source": "riido_metadata", "format": "json", "version": version, "offset": offset, "content": string(raw[offset:end]), "nextOffset": next, "complete": end == len(raw)}
	}
	root["contentPage"] = page(0, 1024)
	root["repositories"] = []any{}
	root["github"] = []any{}
	for _, p := range root["partsStatus"].([]any) {
		p.(map[string]any)["status"] = "deferred"
		p.(map[string]any)["complete"] = false
	}
	calls := 0
	r := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(_ context.Context, req *http.Request) (*http.Response, error) {
		calls++
		var body struct {
			Query     string
			Variables map[string]json.RawMessage
		}
		if req.Method != http.MethodPost || json.NewDecoder(req.Body).Decode(&body) != nil {
			t.Fatal("bad request")
		}
		if string(body.Variables["componentId"]) != `"work-one"` || string(body.Variables["teamId"]) != `"team-one"` {
			t.Fatal("wrong owner scope")
		}
		if strings.Contains(body.Query, "getComponentReadScopeBody") {
			var offset, limit int
			_ = json.Unmarshal(body.Variables["offset"], &offset)
			_ = json.Unmarshal(body.Variables["limit"], &limit)
			if string(body.Variables["version"]) != fmt.Sprintf("%q", version) || string(body.Variables["selector"]) != `{}` {
				t.Fatal("lost version/selector")
			}
			raw, _ := json.Marshal(map[string]any{"data": map[string]any{"getComponentReadScopeBody": page(offset, limit)}})
			return response(200, string(raw)), nil
		}
		if !strings.Contains(body.Query, "query RiidoCLIRepositories") {
			t.Fatal("not a fixed repository query")
		}
		for _, forbidden := range []string{"mutation", "body ", "comments{", "aiSummary", "links{", "attachments{", "figma{"} {
			if strings.Contains(body.Query, forbidden) {
				t.Fatalf("overfetch %s", forbidden)
			}
		}
		return response(200, resourceFixtureJSON(f)), nil
	})}
	got, err := r.Repositories(t.Context(), "work-one", "team-one")
	if err != nil || !got.Complete || len(got.Repositories) != 2 || len(got.PullRequests) != 1 || got.MetadataVersion == nil || *got.MetadataVersion != version || calls != 4 {
		t.Fatalf("deferred %+v err=%v calls=%d", got, err, calls)
	}
}

func TestRepositoriesOnlyRequestedRootAndIncompleteEmpty(t *testing.T) {
	f := repositoriesFixture()
	works := f["works"].([]any)
	root := works[0].(map[string]any)
	adjacent := map[string]any{"id": "other-work", "kind": "component", "teamId": "other-team", "reasons": []any{map[string]any{"kind": "related", "relation": "blocks"}}, "repositories": []any{}, "github": []any{}, "contentPage": map[string]any{"version": "unread-related"}}
	f["works"] = append(works, adjacent)
	r := relationReader(resourceFixtureJSON(f))
	got, err := r.Repositories(t.Context(), "work-one", "team-one")
	if err != nil || len(got.Repositories) != 2 || len(got.PullRequests) != 1 {
		t.Fatalf("root projection failed: %+v %v", got, err)
	}
	root["repositories"] = []any{}
	root["github"] = []any{}
	for _, part := range root["partsStatus"].([]any) {
		p := part.(map[string]any)
		p["status"] = "denied"
		p["complete"] = false
	}
	got, err = relationReader(resourceFixtureJSON(f)).Repositories(t.Context(), "work-one", "team-one")
	if err != nil || got.Complete || got.Repositories == nil || got.PullRequests == nil {
		t.Fatalf("incomplete empty lost: %+v %v", got, err)
	}
}
