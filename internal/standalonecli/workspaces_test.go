package standalonecli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestWorkspacesDiscardResultWhenSessionChanges(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Header.Get("Authorization") != "Bearer fixture-access" {
			t.Error("wrong login")
		}
		fmt.Fprint(w, `{"data":{"getCLIWorkspaces":{"schemaVersion":"riido-cli-workspaces.v1","userId":"user-one","limit":50,"items":[{"id":"ws-one","name":"private title","isDemo":false}],"nextCursor":null,"complete":true}}}`)
	}))
	defer server.Close()
	store := &meetingListSwitchStore{testStore: &testStore{present: true, credential: userauth.Credential{SessionID: "fixture-session", AccessToken: "fixture-access", RefreshToken: "fixture-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}, switchAt: 3}
	service, err := New(Config{APIOrigin: server.URL, Store: store})
	if err != nil {
		t.Fatal(err)
	}
	got, err := service.Workspaces(context.Background(), userapi.WorkspaceListRequest{Limit: 50})
	if got != nil || !errors.Is(err, userauth.ErrLoggedOut) || calls.Load() != 1 {
		t.Fatal("changed login result exposed", got, err, calls.Load())
	}
}

func TestWorkspaceSelectionDiscardsChangedLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":{"getCLIWorkspaces":{"schemaVersion":"riido-cli-workspaces.v1","userId":"user-one","limit":100,"items":[{"id":"ws-one","name":"Workspace","isDemo":false}],"nextCursor":null,"complete":true}}}`)
	}))
	defer server.Close()
	store := &meetingListSwitchStore{testStore: &testStore{present: true, credential: userauth.Credential{SessionID: "fixture-session", AccessToken: "fixture-access", RefreshToken: "fixture-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}, switchAt: 3}
	s, err := New(Config{APIOrigin: server.URL, Store: store})
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Workspace(context.Background(), "ws-one")
	if got != nil || !errors.Is(err, userauth.ErrLoggedOut) {
		t.Fatal("selection survived login change", got, err)
	}
}

func TestWorkspaceSelectionPaginatesOneSession(t *testing.T) {
	for _, changeUser := range []bool{false, true} {
		t.Run(fmt.Sprint(changeUser), func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := calls.Add(1)
				var body struct{ Variables userapi.WorkspaceListRequest }
				if json.NewDecoder(r.Body).Decode(&body) != nil || body.Variables.Limit != 100 || r.Header.Get("Authorization") != "Bearer fixture-access" {
					t.Error("wrong request")
				}
				page := userapi.WorkspaceList{SchemaVersion: "riido-cli-workspaces.v1", UserID: "user-one", Limit: 100, Items: []userapi.WorkspaceItem{}, Complete: n == 2}
				if n == 1 {
					if body.Variables.Cursor != nil {
						t.Error("unexpected cursor")
					}
					for i := range 100 {
						page.Items = append(page.Items, userapi.WorkspaceItem{ID: fmt.Sprintf("ws-%03d", i)})
					}
					next := "bmV4dA"
					page.NextCursor = &next
				} else {
					if body.Variables.Cursor == nil || *body.Variables.Cursor != "bmV4dA" {
						t.Error("cursor lost")
					}
					page.Items = append(page.Items, userapi.WorkspaceItem{ID: "ws-target", Name: "Selected"})
					if changeUser {
						page.UserID = "user-two"
					}
				}
				if err := json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getCLIWorkspaces": page}}); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			store := &testStore{present: true, credential: userauth.Credential{SessionID: "fixture-session", AccessToken: "fixture-access", RefreshToken: "fixture-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
			s, err := New(Config{APIOrigin: server.URL, Store: store})
			if err != nil {
				t.Fatal(err)
			}
			got, err := s.Workspace(context.Background(), "ws-target")
			if calls.Load() != 2 {
				t.Fatal(calls.Load())
			}
			if changeUser {
				if got != nil || !errors.Is(err, userapi.ErrInvalidResponse) {
					t.Fatal(got, err)
				}
			} else if err != nil || got == nil || got.ID != "ws-target" {
				t.Fatal(got, err)
			}
		})
	}
}
