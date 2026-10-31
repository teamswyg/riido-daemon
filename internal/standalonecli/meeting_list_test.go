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

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type meetingListSwitchStore struct {
	*testStore
	loads, switchAt int
}

func (s *meetingListSwitchStore) Load(ctx context.Context) (userauth.Credential, error) {
	s.loads++
	if s.loads == s.switchAt {
		s.credential.SessionID = "other-session"
		s.credential.AccessToken = "other-access"
	}
	return s.testStore.Load(ctx)
}

func TestMeetingsDiscardResultWhenSessionChanges(t *testing.T) {
	for _, switchAt := range []int{3, 4} {
		t.Run(fmt.Sprint(switchAt), func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer fixture-access" {
					t.Error("another account reached the API")
				}
				if calls.Add(1) == 1 {
					fmt.Fprint(w, `{"data":{"getComponentReadScope":{"schemaVersion":"riido-read-scope.v1","rootId":"work","teamId":"team","status":"available","selection":{"relatedMeetingLimit":5},"meetings":[{"id":"meeting","teamId":"team","reasons":[{"kind":"meeting"}]}]}}}`)
				} else {
					fmt.Fprint(w, `{"data":{"getMeetingReadScope":{"id":"meeting","teamId":"team","status":"available","summary":{"id":"meeting","kind":"meeting","title":"private title"}}}}`)
				}
			}))
			defer server.Close()
			store := &meetingListSwitchStore{testStore: &testStore{present: true, credential: userauth.Credential{SessionID: "fixture-session", AccessToken: "fixture-access", RefreshToken: "fixture-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}, switchAt: switchAt}
			service, err := New(Config{APIOrigin: server.URL, Store: store})
			if err != nil {
				t.Fatal(err)
			}
			result, err := service.Meetings(context.Background(), "work", "team")
			if result != nil || !errors.Is(err, userauth.ErrLoggedOut) || calls.Load() != int32(switchAt-2) {
				t.Fatal("mixed-session result escaped or call continued", result, err, calls.Load())
			}
		})
	}
}
