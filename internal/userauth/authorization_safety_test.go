package userauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestAuthorizationParentCancellationStopsTokenExchange(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
		fmt.Fprint(w, `{"access_token":"access","token_type":"Bearer","expires_in":3600}`)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	flow, err := client.BeginAuthorization(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer flow.Close()
	parsed, _ := url.Parse(flow.URL)
	response, err := http.Get(flow.RedirectURI + "?state=" + url.QueryEscape(parsed.Query().Get("state")) + "&code=code")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	done := make(chan error, 1)
	go func() { _, err := flow.Wait(context.Background()); done <- err }()
	<-entered
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("exchange cancellation=%v", err)
		}
		close(release)
	case <-time.After(300 * time.Millisecond):
		close(release)
		<-done
		t.Error("parent cancellation did not stop token exchange")
	}
}

func TestAuthorizationFormattingRedactsVerifier(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	flow, err := client.BeginAuthorization(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer flow.Close()
	if strings.Contains(fmt.Sprintf("%#v", flow), flow.verifier) {
		t.Fatal("authorization formatter exposed PKCE verifier")
	}
}

func TestCallbackRejectsSpoofedHostDuplicateAndReplayedState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	flow, err := client.BeginAuthorization(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer flow.Close()
	parsed, _ := url.Parse(flow.URL)
	suffix := "?state=" + url.QueryEscape(parsed.Query().Get("state")) + "&code=code"
	request, _ := http.NewRequest(http.MethodGet, flow.RedirectURI+suffix, nil)
	request.Host = "localhost"
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatal("spoofed Host accepted")
	}
	response, err = http.Get(flow.RedirectURI + suffix + "&state=duplicate")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatal("duplicate state accepted")
	}
	for _, expected := range []int{200, 409} {
		response, err = http.Get(flow.RedirectURI + suffix)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != expected {
			t.Fatalf("callback status=%d", response.StatusCode)
		}
	}
}
