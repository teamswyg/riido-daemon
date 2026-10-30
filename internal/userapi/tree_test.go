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

func treeFixture() map[string]any {
	row := func(id, team string, reasons ...string) map[string]any {
		r := []any{}
		for _, kind := range reasons {
			r = append(r, map[string]any{"kind": kind})
		}
		return map[string]any{
			"id": id, "kind": "component", "teamId": team, "reasons": r,
			"summary": map[string]any{"id": id, "kind": "component", "title": id, "keyNumber": nil, "componentType": "task"},
		}
	}
	return map[string]any{
		"schemaVersion": ReadScopeSchemaVersion, "rootId": "root", "teamId": "team",
		"status": "available", "selection": map[string]any{"childrenLimit": 10}, "works": []any{
			row("root", "team", "root"), row("mentioned", "team", "mention"),
			row("child-z", "team", "child", "related"), row("parent", "team", "parentTask"),
			row("goal-project", "other-team", "milestone", "project"), row("child-a", "team", "child"),
			row("related", "team", "related"),
		},
	}
}

func readTreeFixture(scope any) (*WorkTree, error) {
	raw, _ := json.Marshal(map[string]any{"data": map[string]any{"getComponentReadScope": scope}})
	return relationReader(string(raw)).Tree(context.Background(), "root", "team")
}

func TestTreeProjectsOnlyDirectAuthorizedHierarchy(t *testing.T) {
	tree, err := readTreeFixture(treeFixture())
	if err != nil {
		t.Fatal(err)
	}
	if tree.Root.ID != "root" || tree.MaxDepth != 1 || tree.ChildrenLimit != 10 || tree.ContextStatus != "available" {
		t.Fatalf("bad metadata: %+v", tree)
	}
	roles, children := []string{}, []string{}
	for _, a := range tree.Ancestors {
		roles = append(roles, a.Relation+":"+a.Work.ID)
	}
	for _, child := range tree.Children {
		children = append(children, child.ID)
	}
	if !reflect.DeepEqual(roles, []string{"project:goal-project", "milestone:goal-project", "parentTask:parent"}) || !reflect.DeepEqual(children, []string{"child-z", "child-a"}) {
		t.Fatalf("roles=%v children=%v", roles, children)
	}
	if tree.Ancestors[0].Work.TeamID != "other-team" || tree.Root.KeyNumber != nil {
		t.Fatal("authorized cross-team/null metadata lost")
	}
}

func TestTreeKeepsEmptyAndPartialWithoutInventingCompleteness(t *testing.T) {
	f := treeFixture()
	f["works"] = f["works"].([]any)[:1]
	f["status"] = "partial"
	f["works"].([]any)[0].(map[string]any)["summary"].(map[string]any)["title"] = ""
	tree, err := readTreeFixture(f)
	if err != nil || tree.ContextStatus != "partial" || tree.Root.Title != "" || tree.Children == nil || tree.Ancestors == nil || len(tree.Children) != 0 {
		t.Fatalf("tree=%+v err=%v", tree, err)
	}
}

func TestTreeEnforcesSelectedChildrenLimit(t *testing.T) {
	f := treeFixture()
	rows := f["works"].([]any)[:1]
	for i := range 11 {
		child := treeFixture()["works"].([]any)[2].(map[string]any)
		id := strings.Repeat("a", i+1)
		child["id"] = id
		child["summary"].(map[string]any)["id"] = id
		rows = append(rows, child)
	}
	f["works"] = rows[:11]
	if got, err := readTreeFixture(f); err != nil || len(got.Children) != 10 {
		t.Fatalf("ten children rejected: %v", err)
	}
	f["works"] = rows
	if got, err := readTreeFixture(f); got != nil || !errors.Is(err, ErrInvalidResponse) {
		t.Fatal("overflow returned partial tree")
	}
}

func TestTreeExcludesMentionedBacklogsAndIntakes(t *testing.T) {
	f := treeFixture()
	rows := f["works"].([]any)
	for _, kind := range []string{"backlog", "intake"} {
		row := treeFixture()["works"].([]any)[1].(map[string]any)
		row["kind"], row["id"] = kind, kind
		summary := row["summary"].(map[string]any)
		summary["kind"], summary["id"], summary["componentType"] = kind, kind, nil
		rows = append(rows, row)
	}
	f["works"] = rows
	got, err := readTreeFixture(f)
	if err != nil || len(got.Children) != 2 || len(got.Ancestors) != 3 {
		t.Fatalf("leaf mention affected hierarchy: %+v %v", got, err)
	}
}

func TestTreeRejectsInvalidContractWithoutPartialOutput(t *testing.T) {
	for name, change := range map[string]func(map[string]any){
		"schema":        func(f map[string]any) { f["schemaVersion"] = "future" },
		"scope-root":    func(f map[string]any) { f["rootId"] = "other" },
		"scope-team":    func(f map[string]any) { f["teamId"] = "other" },
		"status":        func(f map[string]any) { f["status"] = "denied" },
		"selection":     func(f map[string]any) { f["selection"] = nil },
		"quota":         func(f map[string]any) { f["selection"].(map[string]any)["childrenLimit"] = 100 },
		"null-works":    func(f map[string]any) { f["works"] = nil },
		"no-root":       func(f map[string]any) { f["works"] = f["works"].([]any)[1:] },
		"null-row":      func(f map[string]any) { f["works"].([]any)[1] = nil },
		"duplicate-row": func(f map[string]any) { f["works"].([]any)[1] = f["works"].([]any)[0] },
		"ambiguous-parent": func(f map[string]any) {
			f["works"].([]any)[1].(map[string]any)["reasons"] = []any{map[string]any{"kind": "parentTask"}}
		},
		"root-as-child": func(f map[string]any) {
			f["works"].([]any)[0].(map[string]any)["reasons"] = []any{map[string]any{"kind": "root"}, map[string]any{"kind": "child"}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			f := treeFixture()
			change(f)
			tree, err := readTreeFixture(f)
			if tree != nil || !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("accepted: %+v %v", tree, err)
			}
		})
	}
	for name, change := range map[string]func(map[string]any){
		"summary-null":  func(row map[string]any) { row["summary"] = nil },
		"summary-id":    func(row map[string]any) { row["summary"].(map[string]any)["id"] = "different" },
		"summary-kind":  func(row map[string]any) { row["summary"].(map[string]any)["kind"] = "backlog" },
		"title-null":    func(row map[string]any) { row["summary"].(map[string]any)["title"] = nil },
		"team-missing":  func(row map[string]any) { delete(row, "teamId") },
		"reason-null":   func(row map[string]any) { row["reasons"] = []any{nil} },
		"reasons-empty": func(row map[string]any) { row["reasons"] = []any{} },
	} {
		t.Run(name, func(t *testing.T) {
			f := treeFixture()
			change(f["works"].([]any)[2].(map[string]any))
			tree, err := readTreeFixture(f)
			if tree != nil || !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("accepted: %+v %v", tree, err)
			}
		})
	}
}

func TestTreeUsesOneFixedSummaryQueryAndPreservesReadErrors(t *testing.T) {
	calls := 0
	raw, _ := json.Marshal(map[string]any{"data": map[string]any{"getComponentReadScope": treeFixture()}})
	r := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(_ context.Context, req *http.Request) (*http.Response, error) {
		calls++
		var body struct {
			Query     string
			Variables map[string]string
		}
		if json.NewDecoder(req.Body).Decode(&body) != nil || req.Method != http.MethodPost || body.Variables["componentId"] != "root" || body.Variables["teamId"] != "team" || !strings.Contains(body.Query, "getComponentReadScope") {
			t.Fatal("wrong request")
		}
		for _, forbidden := range []string{"mutation", "details", "body", "meetings", "comments", "contentPage"} {
			if strings.Contains(body.Query, forbidden) {
				t.Fatalf("overfetch: %s", forbidden)
			}
		}
		return response(200, string(raw)), nil
	})}
	if _, err := r.Tree(context.Background(), "root", "team"); err != nil || calls != 1 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
	for _, body := range []string{`{"data":{"getComponentReadScope":null}}`, `{"data":{},"errors":[{"message":"private-server-detail"}]}`, strings.Repeat("x", maxBodyBytes+1)} {
		got, err := relationReader(body).Tree(context.Background(), "root", "team")
		if got != nil || err == nil || strings.Contains(err.Error(), "private-server-detail") {
			t.Fatal("partial/private data escaped")
		}
	}
}
