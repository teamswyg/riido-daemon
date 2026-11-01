package standalonecli

import (
	"context"
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
