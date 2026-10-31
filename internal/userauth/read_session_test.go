package userauth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func readSessionRequest(t *testing.T, reader SessionReadClient, endpoint string) error {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(`{"query":"query { me { id } }"}`))
	if err != nil {
		t.Fatal(err)
	}
	res, err := reader.DoRead(context.Background(), req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, err = io.Copy(io.Discard, res.Body)
	return err
}

func TestReadSessionBindsAccountAcrossReadsAndParsing(t *testing.T) {
	for _, at := range []string{"between requests", "after parsing", "generation after parsing"} {
		t.Run(at, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Header.Get("Authorization") != "Bearer old-access" {
					t.Error("another account reached the API")
				}
				fmt.Fprint(w, `{}`)
			}))
			defer server.Close()
			client, _ := NewClient(configFor(server))
			store := newFakeStore(validCredential())
			manager := NewManager(client, store)
			err := manager.WithReadSession(context.Background(), func(reader SessionReadClient) error {
				if err := readSessionRequest(t, reader, server.URL+"/graphql"); err != nil {
					return err
				}
				if at == "generation after parsing" {
					manager.generation.Add(1)
				} else {
					// Synthetic replacement under the already held lock tests the
					// SessionID fence independently of cooperative store writers.
					store.credential.SessionID = "other-session"
					store.credential.AccessToken = "other-access"
				}
				if at == "between requests" {
					return readSessionRequest(t, reader, server.URL+"/graphql")
				}
				return nil
			})
			if !errors.Is(err, ErrLoggedOut) || calls.Load() != 1 {
				t.Fatal("account switch allowed mixed result", err, calls.Load())
			}
		})
	}
}

func TestReadSessionHoldsStoreLockUntilParsingCompletes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, `{}`) }))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	store := newFakeStore(validCredential())
	manager, other := NewManager(client, store), NewManager(client, store)
	var retained SessionReadClient
	err := manager.WithReadSession(context.Background(), func(reader SessionReadClient) error {
		retained = reader
		for range 2 {
			if err := readSessionRequest(t, reader, server.URL+"/graphql"); err != nil {
				return err
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		if _, err := other.AccessToken(ctx); !errors.Is(err, context.DeadlineExceeded) {
			t.Error("second manager entered during compound response parsing", err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.AccessToken(context.Background()); err != nil {
		t.Fatal("store lock was not released", err)
	}
	if err := readSessionRequest(t, retained, server.URL+"/graphql"); !errors.Is(err, ErrLoggedOut) {
		t.Fatal("session client remained usable after callback", err)
	}
}

func TestReadSessionRefreshRetainsOriginalLoginAndExactRead(t *testing.T) {
	var reads, refreshes atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			refreshes.Add(1)
			fmt.Fprint(w, `{"access_token":"new-access","refresh_token":"new-refresh","token_type":"Bearer","expires_in":3600,"refresh_recovery":"v1"}`)
			return
		}
		reads.Add(1)
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"query":"query { me { id } }"}` {
			t.Error("replay body changed")
		}
		if reads.Load() == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Authorization") != "Bearer new-access" {
			t.Error("refreshed credential not retained")
		}
		fmt.Fprint(w, `{}`)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	store := newFakeStore(validCredential())
	err := NewManager(client, store).WithReadSession(context.Background(), func(reader SessionReadClient) error {
		for range 2 {
			if err := readSessionRequest(t, reader, server.URL+"/graphql"); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil || reads.Load() != 3 || refreshes.Load() != 1 || store.credential.SessionID != "fixture-session" {
		t.Fatal("compound refresh changed session/replay count", err, reads.Load(), refreshes.Load())
	}
}

func TestReadSessionRejectsUnsafeRequestsAndReleasesAfterError(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	manager := NewManager(client, newFakeStore(validCredential()))
	for _, mode := range []string{"other origin", "host override", "write", "no replay", "nil"} {
		t.Run(mode, func(t *testing.T) {
			err := manager.WithReadSession(context.Background(), func(reader SessionReadClient) error {
				req, _ := http.NewRequest(http.MethodPost, server.URL+"/graphql", strings.NewReader(`{}`))
				switch mode {
				case "other origin":
					req.URL.Host = "other.invalid"
				case "host override":
					req.Host = "other.invalid"
				case "write":
					req.Method = http.MethodDelete
				case "no replay":
					req.GetBody = nil
				case "nil":
					req = nil
				}
				res, err := reader.DoRead(context.Background(), req)
				if res != nil {
					res.Body.Close()
					t.Error("unsafe request got a response")
				}
				return err
			})
			if err == nil || calls.Load() != 0 {
				t.Fatal("unsafe request reached API")
			}
			if _, err := manager.AccessToken(context.Background()); err != nil {
				t.Fatal("error retained store lock", err)
			}
		})
	}
}

func TestReadSessionCancellationDuringAndAfterParsing(t *testing.T) {
	for _, during := range []bool{true, false} {
		t.Run(fmt.Sprint(during), func(t *testing.T) {
			entered := make(chan struct{})
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				fmt.Fprint(w, `{"body":`)
				if during {
					w.(http.Flusher).Flush()
					close(entered)
					<-r.Context().Done()
					return
				}
				fmt.Fprint(w, `"value"}`)
			}))
			defer server.Close()
			client, _ := NewClient(configFor(server))
			manager := NewManager(client, newFakeStore(validCredential()))
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() {
				done <- manager.WithReadSession(ctx, func(reader SessionReadClient) error {
					req, err := http.NewRequest(http.MethodPost, server.URL+"/graphql", strings.NewReader(`{}`))
					if err != nil {
						return err
					}
					res, err := reader.DoRead(ctx, req)
					if err != nil {
						return err
					}
					defer res.Body.Close()
					_, err = io.ReadAll(res.Body)
					if !during {
						cancel()
					}
					return err
				})
			}()
			if during {
				select {
				case <-entered:
					cancel()
				case <-time.After(3 * time.Second):
					cancel()
					t.Fatal("response did not start")
				}
			}
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) || calls.Load() != 1 {
					t.Fatal("cancellation was swallowed or read repeated", err, calls.Load())
				}
			case <-time.After(3 * time.Second):
				t.Fatal("canceled read retained the lock")
			}
			if _, err := manager.AccessToken(context.Background()); err != nil {
				t.Fatal("cancellation failed to release the lock", err)
			}
		})
	}
}
