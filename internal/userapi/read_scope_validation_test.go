package userapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func readScopeFixture(t *testing.T) *ReadScope {
	t.Helper()
	complete, offset := true, 0
	return &ReadScope{
		SchemaVersion: ReadScopeSchemaVersion, RootID: "root", TeamID: "team", Status: "available",
		Selection: &ReadScopeSelection{5, 5, 10}, Meetings: []*ReadMeetingReference{},
		Works: []*ReadScopeWork{{
			ID: "root", Kind: "component", TeamID: "team", Reasons: []*ReadScopeReason{{Kind: "root"}},
			PartsStatus: []*ReadPartStatus{{Name: "body", Status: "empty", Complete: &complete}},
			Details:     json.RawMessage(`{"id":"root"}`),
			Body:        &ReadBodyPage{ResourceID: "root", Part: "document", Status: "empty", Source: "provider_json", Format: "json", Version: "sha256:" + strings.Repeat("a", 64), Offset: &offset, Complete: &complete},
			Comments:    &ReadScopeCommentPage{ResourceID: "root", ResourceKind: "component", Limit: 20, Version: "sha256:" + strings.Repeat("a", 64), Status: "available", Page: 1, Complete: &complete, Items: []json.RawMessage{}},
			Priorities:  []json.RawMessage{}, Checklists: []json.RawMessage{}, Attachments: []json.RawMessage{},
			Sprints: []json.RawMessage{}, Links: []json.RawMessage{}, GitHub: []json.RawMessage{}, Repositories: []json.RawMessage{}, Figma: json.RawMessage(`null`),
		}},
	}
}

func TestReadScopeRejectsMalformedSelectionAndInitialPartStates(t *testing.T) {
	changes := map[string]func(*ReadScope){
		"missing required sprints": func(s *ReadScope) { s.Works[0].Sprints = nil },
		"unknown work kind":        func(s *ReadScope) { s.Works[0].Kind = "unknown" },
		"null reason":              func(s *ReadScope) { s.Works[0].Reasons = []*ReadScopeReason{nil} },
		"empty reason":             func(s *ReadScope) { s.Works[0].Reasons = []*ReadScopeReason{{}} },
		"bad selection limit":      func(s *ReadScope) { s.Selection.ChildrenLimit = -1 },
		"body missing completion":  func(s *ReadScope) { s.Works[0].Body.Complete = nil },
		"body missing version":     func(s *ReadScope) { s.Works[0].Body.Version = "" },
		"body wrong description target": func(s *ReadScope) {
			b := *s.Works[0].Body
			b.ResourceID = "other"
			b.Part = "description"
			s.Works[0].DescriptionBody = &b
		},
		"comment page missing status": func(s *ReadScope) { s.Works[0].Comments.Status = "" },
		"comment page cannot progress": func(s *ReadScope) {
			p := 1
			done := false
			s.Works[0].Comments.Complete = &done
			s.Works[0].Comments.NextPage = &p
		},
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			s := readScopeFixture(t)
			change(s)
			if validReadScope(s, "root", "team") {
				t.Fatal("accepted malformed selected scope")
			}
		})
	}
	if !validReadScope(readScopeFixture(t), "root", "team") {
		t.Fatal("valid authoritative empty source was rejected")
	}
}
