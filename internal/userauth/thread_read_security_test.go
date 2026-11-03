package userauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestThreadReadRejectsInvalidTicketsBeforeCP(t *testing.T) {
	tests := map[string]func(map[string]any) []byte{
		"wrong origin": func(v map[string]any) []byte {
			v["resource"] = "https://attacker.invalid"
			b, _ := json.Marshal(v)
			return b
		},
		"expired": func(v map[string]any) []byte {
			v["expires_at"] = time.Now().Add(-time.Second).Format(time.RFC3339Nano)
			b, _ := json.Marshal(v)
			return b
		},
		"long lived": func(v map[string]any) []byte {
			v["expires_at"] = time.Now().Add(time.Hour).Format(time.RFC3339Nano)
			b, _ := json.Marshal(v)
			return b
		},
		"OAuth instead": func(v map[string]any) []byte { v["access_token"] = "old-access"; b, _ := json.Marshal(v); return b },
		"noncanonical ticket": func(v map[string]any) []byte {
			v["access_token"] = "riido_cli_thread_v1_" + strings.Repeat("A", 42) + "B"
			b, _ := json.Marshal(v)
			return b
		},
		"wrong schema":  func(v map[string]any) []byte { v["schema_version"] = "future"; b, _ := json.Marshal(v); return b },
		"wrong purpose": func(v map[string]any) []byte { v["purpose"] = "admin"; b, _ := json.Marshal(v); return b },
		"wrong type":    func(v map[string]any) []byte { v["token_type"] = "Basic"; b, _ := json.Marshal(v); return b },
		"missing":       func(v map[string]any) []byte { delete(v, "purpose"); b, _ := json.Marshal(v); return b },
		"extra":         func(v map[string]any) []byte { v["url"] = "secret"; b, _ := json.Marshal(v); return b },
		"null":          func(v map[string]any) []byte { v["purpose"] = nil; b, _ := json.Marshal(v); return b },
		"case alias": func(v map[string]any) []byte {
			v["Purpose"] = v["purpose"]
			delete(v, "purpose")
			b, _ := json.Marshal(v)
			return b
		},
		"duplicate": func(v map[string]any) []byte {
			b, _ := json.Marshal(v)
			return append([]byte(`{"purpose":"task_threads_read",`), b[1:]...)
		},
		"trailing": func(v map[string]any) []byte { b, _ := json.Marshal(v); return append(b, []byte(` {}`)...) },
		"oversize": func(map[string]any) []byte { return []byte(strings.Repeat(" ", 4097)) },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			var calls atomic.Int32
			m, _ := threadReadTestManager(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(mutate(threadReadTicketFixture()))
			}, func(w http.ResponseWriter, r *http.Request) { calls.Add(1) })
			raw, _, err := m.ReadTaskThreads(context.Background(), "workspace-one", "task-one")
			var ae *Error
			if !errors.As(err, &ae) || ae.Code != "INVALID_RESPONSE" || len(raw) != 0 || calls.Load() != 0 {
				t.Fatal("invalid ticket was not rejected before CP", err, calls.Load())
			}
		})
	}
}

func TestThreadReadRejectsUnsupportedOrUnsafeBeforeStorage(t *testing.T) {
	for _, value := range []string{"https://custom.invalid", "https://monolith.api.riido.io/", "https://MONOLITH.api.riido.io", "http://127.0.0.1:3000"} {
		t.Run(value, func(t *testing.T) {
			m, s := threadReadTestManager(t, func(http.ResponseWriter, *http.Request) { t.Error("unexpected network") }, func(http.ResponseWriter, *http.Request) { t.Error("unexpected network") })
			m.client.config.Resource = value
			_, _, err := m.ReadTaskThreads(context.Background(), "workspace", "task")
			if !errors.Is(err, ErrThreadReadUnsupported) || s.loads != 0 {
				t.Fatal("unsupported origin reached credentials")
			}
		})
	}
	for _, value := range []string{"", "..", "a/b", "a%2Fb", "a?query", "a\n", "a b", "한글"} {
		t.Run("ID "+value, func(t *testing.T) {
			m, s := threadReadTestManager(t, func(http.ResponseWriter, *http.Request) { t.Error("unexpected network") }, func(http.ResponseWriter, *http.Request) { t.Error("unexpected network") })
			_, _, err := m.ReadTaskThreads(context.Background(), "workspace", value)
			if err == nil || s.loads != 0 {
				t.Fatal("unsafe ID reached credentials")
			}
		})
	}
}

func TestThreadReadDoesNotRetryFollowRedirectOrUseCookies(t *testing.T) {
	for _, hop := range []string{"exchange", "CP"} {
		for _, status := range []int{301, 302, 303, 307, 308, 401, 403, 404, 429, 503} {
			t.Run(hop+http.StatusText(status), func(t *testing.T) {
				var exchanges, reads atomic.Int32
				m, _ := threadReadTestManager(t, func(w http.ResponseWriter, r *http.Request) {
					exchanges.Add(1)
					if r.Header.Get("Cookie") != "" {
						t.Error("cookie sent")
					}
					w.Header().Set("Set-Cookie", "session=secret; Secure")
					if hop == "exchange" {
						w.Header().Set("Location", "https://monolith.api.riido.io/leak")
						w.WriteHeader(status)
						_, _ = io.WriteString(w, "secret provider body")
						return
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(threadReadTicketFixture())
				}, func(w http.ResponseWriter, r *http.Request) {
					reads.Add(1)
					if r.Header.Get("Cookie") != "" {
						t.Error("cookie sent")
					}
					w.Header().Set("Location", "https://staging.ai-api.riido.io/leak")
					w.WriteHeader(status)
					_, _ = io.WriteString(w, "secret provider body")
				})
				jar, _ := cookiejar.New(nil)
				for _, origin := range []string{"https://monolith.api.riido.io", "https://staging.ai-api.riido.io"} {
					u, _ := url.Parse(origin)
					jar.SetCookies(u, []*http.Cookie{{Name: "session", Value: "private"}})
				}
				m.client.http.Jar = jar
				m.client.http.CheckRedirect = func(*http.Request, []*http.Request) error { return nil }
				raw, got, err := m.ReadTaskThreads(context.Background(), "workspace", "task")
				wantReads := int32(1)
				if hop == "exchange" {
					wantReads = 0
				}
				if err != nil || got != status || len(raw) != 0 || exchanges.Load() != 1 || reads.Load() != wantReads {
					t.Fatal("rejected response retried, followed or exposed", got, err)
				}
			})
		}
	}
}

func TestThreadReadRejectsSessionSwitchAtEitherHop(t *testing.T) {
	for _, hop := range []string{"exchange", "CP"} {
		t.Run(hop, func(t *testing.T) {
			var m *Manager
			var calls atomic.Int32
			m, _ = threadReadTestManager(t, func(w http.ResponseWriter, r *http.Request) {
				if hop == "exchange" {
					m.generation.Add(1)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(threadReadTicketFixture())
			}, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				m.generation.Add(1)
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"private":"history"}`)
			})
			raw, _, err := m.ReadTaskThreads(context.Background(), "workspace", "task")
			if !errors.Is(err, ErrLoggedOut) || len(raw) != 0 || (hop == "exchange" && calls.Load() != 0) {
				t.Fatal("session switch returned history", err)
			}
		})
	}
}

func TestThreadReadLockCoversCPBody(t *testing.T) {
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	m, s := threadReadTestManager(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(threadReadTicketFixture())
	}, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"history":`)
		w.(http.Flusher).Flush()
		close(entered)
		<-release
		_, _ = io.WriteString(w, `[]}`)
	})
	go func() { _, _, err := m.ReadTaskThreads(context.Background(), "workspace", "task"); done <- err }()
	<-entered
	other := NewManager(m.client, s)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	_, err := other.AccessToken(ctx)
	cancel()
	close(release)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Error("other manager acquired lock before body completed", err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if _, err := other.AccessToken(context.Background()); err != nil {
		t.Fatal("lock not released", err)
	}
}

func TestThreadReadRejectsOversizeBody(t *testing.T) {
	m, _ := threadReadTestManager(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(threadReadTicketFixture())
	}, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, strings.Repeat(" ", threadReadBodyLimit+1))
	})
	raw, _, err := m.ReadTaskThreads(context.Background(), "workspace", "task")
	if !errors.Is(err, ErrThreadReadIncomplete) || len(raw) != 0 {
		t.Fatal("oversized response was not failed", err)
	}
}
