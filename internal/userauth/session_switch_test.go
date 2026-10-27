package userauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestReadRetryDoesNotCrossLoginSession(t *testing.T) {
	for _, shared := range []bool{false, true} {
		for _, sameTokens := range []bool{false, true} {
			t.Run(fmt.Sprintf("shared=%t/same-tokens=%t", shared, sameTokens), func(t *testing.T) {
				entered, release := make(chan struct{}), make(chan struct{})
				var reads, refreshes atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/graphql":
						if reads.Add(1) == 1 {
							close(entered)
							<-release
							w.WriteHeader(http.StatusUnauthorized)
							return
						}
						w.WriteHeader(http.StatusOK)
					case "/oauth/token":
						refreshes.Add(1)
						fmt.Fprint(w, `{"access_token":"rotated","refresh_token":"rotated-refresh","token_type":"Bearer","expires_in":3600}`)
					case "/oauth/revoke":
						w.WriteHeader(http.StatusOK)
					}
				}))
				defer server.Close()
				client, _ := NewClient(configFor(server))
				store := newFakeStore(Credential{})
				reader := NewManager(client, store)
				writer := reader
				if shared {
					writer = NewManager(client, store)
				}
				first := validCredential()
				if err := reader.Login(context.Background(), first); err != nil {
					t.Fatal(err)
				}
				request, _ := http.NewRequest(http.MethodPost, server.URL+"/graphql", strings.NewReader(`{"query":"query { me { id } }"}`))
				done := make(chan error, 1)
				go func() {
					response, err := reader.DoRead(context.Background(), request)
					if response != nil {
						response.Body.Close()
					}
					done <- err
				}()
				<-entered
				next := first
				if !sameTokens {
					next.AccessToken = "second-access"
					next.RefreshToken = "second-refresh"
				}
				if err := writer.Login(context.Background(), next); err != nil {
					t.Fatal(err)
				}
				close(release)
				if err := <-done; !errors.Is(err, ErrLoggedOut) {
					t.Fatalf("request crossed login identity: %v", err)
				}
				if reads.Load() != 1 || refreshes.Load() != 0 {
					t.Fatal("old request replayed or refreshed under new login")
				}
			})
		}
	}
}
