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

const relatedComponentFixture = `{"relationCategory":"blockedBy","issueId":"issue-one","issueType":"component","issue":{"id":"issue-one","component":{"id":"component-one","componentType":"task","title":"","isDone":false,"status":{"id":"status-one","statusType":"inProgress","statusColor":"#112233","colorPalette":{"id":"palette-one","colorHex":"","colorHexLight":"#ffffff","textColorHex":"#000000","textColorHexLight":"#000000","statusColorHex":"#112233","statusColorHexLight":"#445566"}}},"backlog":null,"intake":null}}`

func relationReader(body string) Reader {
	return Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(context.Context, *http.Request) (*http.Response, error) {
		return response(http.StatusOK, body), nil
	})}
}

func TestRelationsPreserveDirectionsTypesNullableFieldsAndOrder(t *testing.T) {
	rows := []any{}
	for _, category := range []RelationCategory{RelationRelatedTo, RelationBlocking, RelationBlockedBy, RelationHasDuplicate, RelationDuplicatedOf} {
		var row map[string]any
		if err := json.Unmarshal([]byte(relatedComponentFixture), &row); err != nil {
			t.Fatal(err)
		}
		row["relationCategory"] = category
		rows = append(rows, row)
	}
	for _, kind := range []string{"backlog", "intake"} {
		rows = append(rows, map[string]any{"relationCategory": "relatedTo", "issueId": "issue-" + kind, "issueType": kind, "issue": map[string]any{"id": "issue-" + kind, "component": nil, "backlog": nil, "intake": nil, kind: map[string]any{"id": "resource-" + kind, "title": ""}}})
	}
	for _, kind := range []string{"component", "backlog", "intake"} {
		rows = append(rows, map[string]any{"relationCategory": "relatedTo", "issueId": "nullable-" + kind, "issueType": kind, "issue": map[string]any{"id": "nullable-" + kind, "component": nil, "backlog": nil, "intake": nil}})
	}
	raw, _ := json.Marshal(map[string]any{"data": map[string]any{"getComponentRelatedIssues": rows}})
	got, err := relationReader(string(raw)).Relations(context.Background(), "work-one", "team-one")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var normalized any
	_ = json.Unmarshal(encoded, &normalized)
	wantJSON, _ := json.Marshal(rows)
	var want any
	_ = json.Unmarshal(wantJSON, &want)
	if !reflect.DeepEqual(normalized, want) {
		t.Fatalf("changed relation data: %s", encoded)
	}
	empty, err := relationReader(`{"data":{"getComponentRelatedIssues":[]}}`).Relations(context.Background(), "work-one", "team-one")
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatal("real empty list changed")
	}
}

func TestRelationsRejectMalformedContractWithoutPartialOutput(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"missing", `{"data":{}}`}, {"null", `{"data":{"getComponentRelatedIssues":null}}`},
		{"null-item", `{"data":{"getComponentRelatedIssues":[null]}}`},
		{"category", strings.ReplaceAll(relatedComponentFixture, `"blockedBy"`, `"private-unknown"`)},
		{"type", strings.ReplaceAll(relatedComponentFixture, `"issueType":"component"`, `"issueType":"unknown"`)},
		{"identity", strings.ReplaceAll(relatedComponentFixture, `"issueId":"issue-one"`, `"issueId":"different"`)},
		{"issue-missing", `{"relationCategory":"relatedTo","issueId":"issue-one","issueType":"component","issue":null}`},
		{"done-missing", strings.ReplaceAll(relatedComponentFixture, `"isDone":false,`, ``)},
		{"done-null", strings.ReplaceAll(relatedComponentFixture, `"isDone":false`, `"isDone":null`)},
		{"title-null", strings.ReplaceAll(relatedComponentFixture, `"title":""`, `"title":null`)},
		{"component-type-null", strings.ReplaceAll(relatedComponentFixture, `"componentType":"task"`, `"componentType":null`)},
		{"resource-id-empty", strings.ReplaceAll(relatedComponentFixture, `"id":"component-one"`, `"id":""`)},
		{"status-type-null", strings.ReplaceAll(relatedComponentFixture, `"statusType":"inProgress"`, `"statusType":null`)},
		{"palette-color-missing", strings.ReplaceAll(relatedComponentFixture, `"colorHex":"",`, ``)},
		{"palette-id-empty", strings.ReplaceAll(relatedComponentFixture, `"id":"palette-one"`, `"id":""`)},
		{"multiple-resource", strings.ReplaceAll(relatedComponentFixture, `"backlog":null`, `"backlog":{"id":"backlog-one","title":""}`)},
		{"wrong-resource-type", strings.ReplaceAll(relatedComponentFixture, `"issueType":"component"`, `"issueType":"backlog"`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := tc.body
			if !strings.HasPrefix(body, `{"data"`) {
				body = `{"data":{"getComponentRelatedIssues":[` + relatedComponentFixture + `,` + body + `]}}`
			}
			got, err := relationReader(body).Relations(context.Background(), "work-one", "team-one")
			if !errors.Is(err, ErrInvalidResponse) || got != nil {
				t.Fatalf("malformed response accepted: err=%v count=%d", err, len(got))
			}
		})
	}
}

func TestRelationsKeepReadBoundaries(t *testing.T) {
	for _, body := range []string{`{"data":{"getComponentRelatedIssues":[]} ,"errors":[{"message":"private-response"}]}`, strings.Repeat("x", maxBodyBytes+1)} {
		got, err := relationReader(body).Relations(context.Background(), "work-one", "team-one")
		if err == nil || got != nil || strings.Contains(err.Error(), "private-response") {
			t.Fatal("partial or oversized data escaped")
		}
	}
}
