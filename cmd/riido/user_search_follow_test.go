package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func searchFollowPage(limit int, complete bool, next string, ids ...string) *userapi.WorkSearchPage {
	page := &userapi.WorkSearchPage{Limit: limit, Complete: complete, Items: []*userapi.WorkSearchItem{}}
	if next != "" {
		page.NextCursor = &next
	}
	for _, id := range ids {
		page.Items = append(page.Items, &userapi.WorkSearchItem{Kind: "component", ID: id})
	}
	return page
}

func TestFollowWorkSearchFillsPageAcrossBudgetCuts(t *testing.T) {
	pages := map[string]*userapi.WorkSearchPage{
		"":   searchFollowPage(3, false, "c1", "a"),
		"c1": searchFollowPage(2, false, "c2"),
		"c2": searchFollowPage(2, true, "", "b"),
	}
	limits := []int{}
	search := func(_ context.Context, q userapi.WorkSearchRequest) (*userapi.WorkSearchPage, error) {
		limits = append(limits, q.Limit)
		return pages[q.Cursor], nil
	}
	got, err := followWorkSearch(t.Context(), search, userapi.WorkSearchRequest{Limit: 3})
	if err != nil || !got.Complete || got.NextCursor != nil || len(got.Items) != 2 || got.Items[1].ID != "b" || got.Limit != 3 {
		t.Fatalf("bad merge: %+v %v", got, err)
	}
	if len(limits) != 3 || limits[0] != 3 || limits[1] != 2 || limits[2] != 2 {
		t.Fatalf("continuation limits %v", limits)
	}
}

func TestFollowWorkSearchStopsWithoutLosingReadPages(t *testing.T) {
	first := func() *userapi.WorkSearchPage { return searchFollowPage(3, false, "c1", "a") }
	t.Run("later error keeps cursor", func(t *testing.T) {
		search := func(_ context.Context, q userapi.WorkSearchRequest) (*userapi.WorkSearchPage, error) {
			if q.Cursor == "" {
				return first(), nil
			}
			return nil, userapi.ErrUnavailable
		}
		got, err := followWorkSearch(t.Context(), search, userapi.WorkSearchRequest{Limit: 3})
		if err != nil || got.Complete || got.NextCursor == nil || *got.NextCursor != "c1" || len(got.Items) != 1 {
			t.Fatalf("%+v %v", got, err)
		}
	})
	t.Run("deadline reserve", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		calls := 0
		search := func(context.Context, userapi.WorkSearchRequest) (*userapi.WorkSearchPage, error) {
			calls++
			return first(), nil
		}
		got, err := followWorkSearch(ctx, search, userapi.WorkSearchRequest{Limit: 3})
		if err != nil || calls != 1 || got.Complete || *got.NextCursor != "c1" {
			t.Fatalf("%+v %v calls=%d", got, err, calls)
		}
	})
	t.Run("first error fails", func(t *testing.T) {
		search := func(context.Context, userapi.WorkSearchRequest) (*userapi.WorkSearchPage, error) {
			return nil, userapi.ErrPermission
		}
		if _, err := followWorkSearch(t.Context(), search, userapi.WorkSearchRequest{Limit: 3}); !errors.Is(err, userapi.ErrPermission) {
			t.Fatal(err)
		}
	})
	t.Run("duplicate across pages", func(t *testing.T) {
		search := func(_ context.Context, q userapi.WorkSearchRequest) (*userapi.WorkSearchPage, error) {
			if q.Cursor == "" {
				return first(), nil
			}
			return searchFollowPage(2, true, "", "a"), nil
		}
		if _, err := followWorkSearch(t.Context(), search, userapi.WorkSearchRequest{Limit: 3}); !errors.Is(err, userapi.ErrInvalidResponse) {
			t.Fatal(err)
		}
	})
}

func TestUserWorkSearchExplainsBudgetCutPage(t *testing.T) {
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Variables map[string]any }
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Variables["cursor"] != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"searchWorkspaceWork": map[string]any{"schemaVersion": "riido-work-search.v1", "workspaceId": "workspace-one", "query": "hello", "limit": 20, "complete": false, "nextCursor": "cGFnZS10d28", "items": []any{}}}})
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runWork(h.io, []string{"search", "--workspace", "workspace-one", "--query", "hello"}); err != nil {
		t.Fatal(err)
	}
	out := h.stdout.String()
	if !strings.Contains(out, "제한 시간 안에 끝까지 찾지 못했습니다") || !strings.Contains(out, "cGFnZS10d28") || strings.Contains(out, "검색 결과가 없습니다") {
		t.Fatal(out)
	}
}
