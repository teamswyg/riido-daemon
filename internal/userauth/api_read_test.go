package userauth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestExplicitReadPOSTRetries401WithReplayableBody(t *testing.T) {
	var calls, refreshes atomic.Int32
	const body = `{"query":"query { me { id } }"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			refreshes.Add(1)
			fmt.Fprint(w, `{"access_token":"new-access","token_type":"Bearer","expires_in":3600}`)
			return
		}
		received, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		if string(received) != body {
			t.Error("read body changed during retry")
		}
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Authorization") != "Bearer new-access" {
			t.Error("retry did not use refreshed token")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	manager := NewManager(client, newFakeStore(validCredential()))
	request, _ := http.NewRequest(http.MethodPost, server.URL+"/graphql", strings.NewReader(body))
	response, err := manager.DoRead(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || calls.Load() != 2 || refreshes.Load() != 1 {
		t.Fatal("read POST did not refresh and replay exactly once")
	}
}

func TestOrdinaryPOSTAndForbiddenReadAreNotRetried(t *testing.T) {
	for _, read := range []bool{false, true} {
		t.Run(fmt.Sprint(read), func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if read {
					w.WriteHeader(http.StatusForbidden)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
			}))
			defer server.Close()
			client, _ := NewClient(configFor(server))
			manager := NewManager(client, newFakeStore(validCredential()))
			request, _ := http.NewRequest(http.MethodPost, server.URL+"/graphql", strings.NewReader(`{}`))
			var response *http.Response
			var err error
			if read {
				response, err = manager.DoRead(context.Background(), request)
			} else {
				response, err = manager.Do(context.Background(), request)
			}
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if calls.Load() != 1 {
				t.Fatal("non-retryable POST repeated")
			}
		})
	}
}

func TestExplicitReadPOSTRequiresReplayableBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	store := newFakeStore(validCredential())
	request, _ := http.NewRequest(http.MethodPost, server.URL+"/graphql", io.NopCloser(strings.NewReader(`{}`)))
	response, err := NewManager(client, store).DoRead(context.Background(), request)
	if response != nil {
		response.Body.Close()
	}
	if err == nil || store.loads != 0 {
		t.Fatal("non-replayable read accessed credentials")
	}
}
