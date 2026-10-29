package userapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

const (
	activityLogFixture  = `{"id":"log-one","createdAt":"","category":"","action":"","property":null,"oldValue":"","newValue":"{\"한글\":\"🙂\"}","author":{"id":"member-one","status":"active","workspaceProfile":{"nickname":"","status":"active","avatar":{"fileLocation":null}}}}`
	activityMetaFixture = `{"total":1,"currentPage":1,"nextPage":null,"prevPage":null,"lastPage":1}`
)

func activityResponse(rows, meta string) string {
	return `{"data":{"getComponentActivityLogs":{"activityLogs":` + rows + `,"meta":` + meta + `}}}`
}

func TestActivityPreservesEveryFieldAndNullVersusEmpty(t *testing.T) {
	for _, row := range []string{
		activityLogFixture,
		strings.Replace(activityLogFixture, `"avatar":{"fileLocation":null}`, `"avatar":null`, 1),
		strings.Replace(activityLogFixture, `"fileLocation":null`, `"fileLocation":"https://static.example/avatar"`, 1),
		`{"id":"deleted-author","createdAt":"2026-10-29T01:00:00Z","category":"task","action":"update","property":"","oldValue":null,"newValue":null,"author":null}`,
	} {
		got, err := relationReader(activityResponse("["+row+"]", activityMetaFixture)).Activity(t.Context(), "work-one", "team-one", ActivityPagination{"1", "20"})
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(got.ActivityLogs[0])
		var actual, want any
		_ = json.Unmarshal(raw, &actual)
		_ = json.Unmarshal([]byte(row), &want)
		if !reflect.DeepEqual(actual, want) {
			t.Fatalf("activity fields changed: %s", raw)
		}
	}
}

func TestActivityRejectsMalformedPagesWithoutPartialOutput(t *testing.T) {
	cases := map[string]string{
		"missing-root": `{"data":{}}`, "null-root": `{"data":{"getComponentActivityLogs":null}}`,
		"null-rows": activityResponse("null", activityMetaFixture), "null-entry": activityResponse("[null]", activityMetaFixture),
		"null-meta": activityResponse("[]", "null"),
	}
	for _, field := range []string{"createdAt", "category", "action"} {
		cases["missing-"+field] = activityResponse("["+strings.Replace(activityLogFixture, `"`+field+`":"",`, "", 1)+"]", activityMetaFixture)
		cases["null-"+field] = activityResponse("["+strings.Replace(activityLogFixture, `"`+field+`":""`, `"`+field+`":null`, 1)+"]", activityMetaFixture)
	}
	for name, row := range map[string]string{
		"empty-id":           strings.Replace(activityLogFixture, `"id":"log-one"`, `"id":""`, 1),
		"profile-null":       strings.Replace(activityLogFixture, `"workspaceProfile":{"nickname":"","status":"active","avatar":{"fileLocation":null}}`, `"workspaceProfile":null`, 1),
		"nickname-missing":   strings.Replace(activityLogFixture, `"nickname":"",`, "", 1),
		"author-status-null": strings.Replace(activityLogFixture, `"status":"active"`, `"status":null`, 1),
		"value-object":       strings.Replace(activityLogFixture, `"oldValue":""`, `"oldValue":{}`, 1),
	} {
		cases[name] = activityResponse("["+row+"]", activityMetaFixture)
	}
	for name, meta := range map[string]string{
		"missing-total":  strings.Replace(activityMetaFixture, `"total":1,`, "", 1),
		"null-total":     strings.Replace(activityMetaFixture, `"total":1`, `"total":null`, 1),
		"negative-total": strings.Replace(activityMetaFixture, `"total":1`, `"total":-1`, 1),
		"overflow-total": strings.Replace(activityMetaFixture, `"total":1`, `"total":2147483648`, 1),
		"wrong-page":     strings.Replace(activityMetaFixture, `"currentPage":1`, `"currentPage":2`, 1),
		"wrong-last":     strings.Replace(activityMetaFixture, `"lastPage":1`, `"lastPage":0`, 1),
		"missing-next":   strings.Replace(activityMetaFixture, `"nextPage":null,`, "", 1),
		"wrong-next":     strings.Replace(activityMetaFixture, `"nextPage":null`, `"nextPage":2`, 1),
		"string-prev":    strings.Replace(activityMetaFixture, `"prevPage":null`, `"prevPage":"1"`, 1),
	} {
		cases[name] = activityResponse("["+activityLogFixture+"]", meta)
	}
	cases["count"] = activityResponse("[]", activityMetaFixture)
	cases["partial-errors"] = strings.TrimSuffix(activityResponse("["+activityLogFixture+"]", activityMetaFixture), "}") + `,"errors":[{"message":"private-detail"}]}`
	cases["oversized"] = activityResponse("["+strings.Replace(activityLogFixture, `"oldValue":""`, `"oldValue":"`+strings.Repeat("x", maxBodyBytes)+`"`, 1)+"]", activityMetaFixture)
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := relationReader(body).Activity(t.Context(), "work-one", "team-one", ActivityPagination{"1", "20"})
			if err == nil || got.ActivityLogs != nil || got.Meta != nil || strings.Contains(err.Error(), "private-detail") {
				t.Fatal("malformed activity escaped", err)
			}
		})
	}
}

func TestActivityPaginationRejectsBeforeAnyRead(t *testing.T) {
	for _, p := range []ActivityPagination{{"", "20"}, {"0", "20"}, {"+1", "20"}, {"01", "20"}, {" 1", "20"}, {"1.0", "20"}, {"2147483648", "1"}, {"2147483647", "2"}, {"1", ""}, {"1", "0"}, {"1", "101"}, {"1", "+1"}, {"1", "01"}, {"1", "1.5"}} {
		called := false
		r := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(context.Context, *http.Request) (*http.Response, error) { called = true; return nil, nil })}
		if _, err := r.Activity(t.Context(), "work-one", "team-one", p); !errors.Is(err, ErrInvalidResponse) || called {
			t.Fatalf("invalid page reached transport: %+v", p)
		}
	}
	if !(ActivityPagination{"2147483647", "1"}).Valid() || !(ActivityPagination{"1", "100"}).Valid() {
		t.Fatal("valid boundary rejected")
	}
}

func TestActivityEmptyAndOutOfRangePagesRetainOffsetMeta(t *testing.T) {
	for _, tc := range []struct{ page, meta string }{
		{"1", `{"total":0,"currentPage":1,"nextPage":null,"prevPage":null,"lastPage":0}`},
		{"4", `{"total":45,"currentPage":4,"nextPage":null,"prevPage":3,"lastPage":3}`},
		{"2147483647", `{"total":0,"currentPage":2147483647,"nextPage":null,"prevPage":2147483646,"lastPage":0}`},
	} {
		limit := "20"
		if tc.page == "2147483647" {
			limit = "1"
		}
		got, err := relationReader(activityResponse("[]", tc.meta)).Activity(t.Context(), "work-one", "team-one", ActivityPagination{tc.page, limit})
		if err != nil || got.ActivityLogs == nil || len(got.ActivityLogs) != 0 {
			t.Fatal("empty page changed", err)
		}
	}
}
