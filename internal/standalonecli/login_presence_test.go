package standalonecli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestLoginConfirmationRefusalPreservesPreviousSession(t *testing.T) {
	for _, present := range []bool{false, true} {
		for _, refusal := range []error{userauth.ErrStoreLocked, securecredential.ErrUnavailable, context.Canceled, context.DeadlineExceeded} {
			t.Run(fmt.Sprintf("present=%t/%v", present, refusal), func(t *testing.T) {
				previous := userauth.Credential{AccessToken: "previous-fixture", RefreshToken: "previous-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}
				store := &testStore{present: present, credential: previous, confirmErr: refusal}
				network := 0
				service, opened := testLoginService(t, store, func(w http.ResponseWriter, _ *http.Request) { network++; w.WriteHeader(http.StatusServiceUnavailable) })
				_, err := service.Login(context.Background())
				if !errors.Is(err, refusal) || *opened != 0 || network != 0 || store.saves != 0 || store.confirms != 1 {
					t.Fatalf("confirmation bypassed: err=%v browser=%d network=%d saves=%d confirms=%d", err, *opened, network, store.saves, store.confirms)
				}
				if store.present != present || !reflect.DeepEqual(store.credential, previous) {
					t.Fatal("refused confirmation changed previous credential")
				}
			})
		}
	}
}

func TestLoginCancellationAfterSuccessfulConfirmationStopsBeforeOAuth(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := &testStore{confirmHook: cancel}
	network := 0
	service, opened := testLoginService(t, store, func(w http.ResponseWriter, _ *http.Request) { network++; w.WriteHeader(http.StatusServiceUnavailable) })
	_, err := service.Login(ctx)
	if !errors.Is(err, context.Canceled) || store.confirms != 1 || *opened != 0 || network != 0 || store.saves != 0 {
		t.Fatal("late confirmation success continued after cancellation")
	}
}

func TestNewAndReplacementLoginConfirmBeforeBrowserAndAPI(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(fmt.Sprintf("present=%t", present), func(t *testing.T) {
			store := &testStore{present: present, credential: userauth.Credential{AccessToken: "old-fixture", RefreshToken: "old-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
			service, _ := testLoginService(t, store, func(w http.ResponseWriter, r *http.Request) {
				store.order = append(store.order, r.URL.Path)
				switch r.URL.Path {
				case "/oauth/token":
					fmt.Fprint(w, `{"access_token":"new-fixture","refresh_token":"new-refresh","token_type":"Bearer","expires_in":3600}`)
				case "/graphql":
					fmt.Fprint(w, `{"data":{"getMyAccountInfo":{"id":"fixture-user"}}}`)
				case "/oauth/revoke":
					w.WriteHeader(http.StatusOK)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			})
			open := service.openBrowser
			service.openBrowser = func(ctx context.Context, address string) error {
				store.order = append(store.order, "browser")
				return open(ctx, address)
			}
			result, err := service.Login(context.Background())
			want := []string{"status", "confirm", "browser", "/oauth/token", "/graphql", "save"}
			if present {
				want = append(want, "/oauth/revoke")
			}
			if err != nil || !result.Authenticated || store.confirms != 1 || !reflect.DeepEqual(store.order, want) {
				t.Fatalf("login order=%v err=%v", store.order, err)
			}
		})
	}
}
