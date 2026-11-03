package userauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDesktopAuthorizationCancellationAndTimeoutNeverExchanges(t *testing.T) {
	for _, scenario := range []string{"cancel-after-callback", "cancel-during-sign", "timeout-without-callback", "timeout-during-sign", "proof-expired-before-wait"} {
		t.Run(scenario, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("cancelled authorization reached API") }))
			defer server.Close()
			config := configFor(server)
			if scenario == "timeout-without-callback" || scenario == "timeout-during-sign" {
				config.AuthTimeout = 200 * time.Millisecond
			}
			client, err := NewClient(config)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			signed := 0
			a, err := client.BeginDesktopAuthorization(ctx, fixtureInstallationKey(), func(signCtx context.Context, _ InstallationProof) (string, error) {
				signed++
				if scenario == "cancel-during-sign" {
					cancel()
				}
				if scenario == "cancel-during-sign" || scenario == "timeout-during-sign" {
					<-signCtx.Done()
				}
				return "MAYCAQECAQE", nil
			})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			if scenario != "timeout-without-callback" {
				if postDesktopReply(t, a, desktopReply(t, a), nil) != http.StatusOK {
					t.Fatal("callback")
				}
			}
			if scenario == "cancel-after-callback" {
				cancel()
			}
			if scenario == "proof-expired-before-wait" {
				// Advance only the accepted proof's deadline, without sleeping for
				// its real ten-minute validity or racing the callback handler.
				result := <-a.result
				result.installation.proof.ExpiresAt = time.Now().Add(-time.Second).Unix()
				a.result <- result
			}
			credential, err := a.Wait(t.Context())
			want := context.DeadlineExceeded
			wantSigned := 0
			if scenario == "cancel-after-callback" || scenario == "cancel-during-sign" {
				want = context.Canceled
			}
			if scenario == "cancel-during-sign" || scenario == "timeout-during-sign" {
				wantSigned = 1
			}
			if !errors.Is(err, want) || credential.AccessToken != "" || signed != wantSigned {
				t.Fatalf("err=%v signed=%d", err, signed)
			}
			connection, err := server.Client().Post(a.RedirectURI, "application/json", nil)
			if err == nil {
				_ = connection.Body.Close()
				t.Fatal("terminal callback listener remained open")
			}
		})
	}
}
