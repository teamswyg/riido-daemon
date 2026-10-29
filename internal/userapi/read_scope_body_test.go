package userapi

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"strings"
	"testing"
)

func TestReadScopeBodyFollowsByteOffsetsToCompleteUnicodeBody(t *testing.T) {
	version := "sha256:" + strings.Repeat("a", 64)
	calls := 0
	reader := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(_ context.Context, req *http.Request) (*http.Response, error) {
		calls++
		var input struct {
			Query     string
			Variables ReadBodyRequest
		}
		if json.NewDecoder(req.Body).Decode(&input) != nil || !strings.Contains(input.Query, "getComponentReadScopeBody") || input.Variables.Version != version || input.Variables.ComponentID != "work-one" || input.Variables.TeamID != "team-one" {
			t.Fatal("lost fixed query, original version, or source identity")
		}
		q := input.Variables
		complete := q.Offset == 6
		page := ReadBodyPage{ResourceID: q.ComponentID, Part: q.Part, Source: "provider_json", Format: "json", Version: version, Offset: &q.Offset, Complete: &complete}
		if complete {
			page.Content, page.Status = "다", "available"
		} else {
			next := 6
			page.Content, page.Status, page.NextOffset = "가나", "partial", &next
		}
		raw, _ := json.Marshal(map[string]any{"data": map[string]any{"getComponentReadScopeBody": page}})
		return response(200, string(raw)), nil
	})}
	q := ReadBodyRequest{ComponentID: "work-one", TeamID: "team-one", Part: "document", Version: version, Limit: 6}
	first, err := reader.Body(t.Context(), q)
	if err != nil || first.NextOffset == nil || *first.NextOffset != 6 || *first.Complete {
		t.Fatalf("first page: %+v, %v", first, err)
	}
	q.Offset = *first.NextOffset
	last, err := reader.Body(t.Context(), q)
	if err != nil || !*last.Complete || last.NextOffset != nil || first.Content+last.Content != "가나다" || calls != 2 {
		t.Fatalf("last page: %+v, %v", last, err)
	}
}

func TestReadScopeBodyRejectsWrongIdentityVersionAndNonProgressingPages(t *testing.T) {
	version := "sha256:" + strings.Repeat("a", 64)
	q := ReadBodyRequest{ComponentID: "work-one", TeamID: "team-one", Part: "document", Version: version, Limit: 6}
	base := map[string]any{"resourceId": "work-one", "part": "document", "source": "provider_json", "format": "json", "version": version, "offset": 0, "content": "가나", "nextOffset": 6, "complete": false, "status": "partial"}
	for _, change := range []map[string]any{
		{"resourceId": "other"},
		{"part": "description"},
		{"version": "sha256:" + strings.Repeat("b", 64)},
		{"offset": nil},
		{"offset": 1},
		{"complete": nil},
		{"nextOffset": 0},
		{"nextOffset": 2},
		{"nextOffset": nil},
		{"complete": true},
		{"content": "가나다", "nextOffset": 9},
		{"source": ""},
	} {
		data := make(map[string]any, len(base))
		maps.Copy(data, base)
		maps.Copy(data, change)
		raw, _ := json.Marshal(map[string]any{"data": map[string]any{"getComponentReadScopeBody": data}})
		reader := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(context.Context, *http.Request) (*http.Response, error) { return response(200, string(raw)), nil })}
		if got, err := reader.Body(t.Context(), q); !errors.Is(err, ErrInvalidResponse) || got.ResourceID != "" {
			t.Fatalf("accepted malformed page %v: %+v, %v", change, got, err)
		}
	}
}
