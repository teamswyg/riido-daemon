package standalonecli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

const protectedTestResource = "https://api.example.test"

// This fixture models the opaque native boundary, not OS protection acceptance.
type protectedNativeFixture struct {
	lock         chan struct{}
	mu           sync.Mutex
	raw          []byte
	loadErr      error
	saveErr      error
	commitOnFail bool
	saves        int
	deletes      int
	lockAttempt  chan struct{}
}

func newProtectedFixture(t *testing.T) (*ProtectedStore, *protectedNativeFixture, userauth.Credential, LinkBinding) {
	t.Helper()
	credential := userauth.Credential{SessionID: "original-login-session", AccessToken: "synthetic-access", RefreshToken: "synthetic-refresh", TokenType: "Bearer", Scope: "read", ExpiresAt: time.Now().Add(time.Hour)}
	raw, err := json.Marshal(credential)
	if err != nil {
		t.Fatal(err)
	}
	native := &protectedNativeFixture{lock: make(chan struct{}, 1), raw: raw}
	native.lock <- struct{}{}
	binding := LinkBinding{Resource: protectedTestResource, ClientID: "riido-cli", SessionID: credential.SessionID, KeyFingerprint: strings.Repeat("a", 64), ExpectedDeviceID: "device-original"}
	return &ProtectedStore{Native: native, Resource: protectedTestResource}, native, credential, binding
}

func (n *protectedNativeFixture) Lock(ctx context.Context) (func(), error) {
	if n.lockAttempt != nil {
		n.lockAttempt <- struct{}{}
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-n.lock:
		return func() { n.lock <- struct{}{} }, nil
	}
}

func (n *protectedNativeFixture) Load(context.Context) ([]byte, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.loadErr != nil {
		return nil, n.loadErr
	}
	if len(n.raw) == 0 {
		return nil, securecredential.ErrNotFound
	}
	return bytes.Clone(n.raw), nil
}

func (n *protectedNativeFixture) Save(_ context.Context, raw []byte) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.saves++
	if n.saveErr == nil || n.commitOnFail {
		n.raw = bytes.Clone(raw)
	}
	return n.saveErr
}

func (n *protectedNativeFixture) Delete(context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.deletes++
	n.raw = nil
	return nil
}

func (n *protectedNativeFixture) Status(ctx context.Context) (bool, error) {
	_, err := n.Load(ctx)
	return err == nil, err
}
func (*protectedNativeFixture) ConfirmUserPresence(context.Context) error { return nil }
func (*protectedNativeFixture) EnsureInstallationKey(context.Context, string) (securecredential.InstallationPublicKey, error) {
	return securecredential.InstallationPublicKey{}, securecredential.ErrUnavailable
}

func (*protectedNativeFixture) SignInstallationProof(context.Context, securecredential.InstallationPublicKey, securecredential.InstallationProof) (string, error) {
	return "", securecredential.ErrUnavailable
}

func saveProtectedCredential(t *testing.T, store *ProtectedStore, credential userauth.Credential) error {
	t.Helper()
	unlock, err := store.Lock(context.Background())
	if err != nil {
		return err
	}
	defer unlock()
	return store.Save(context.Background(), credential)
}

func TestProtectedPendingPersistsOriginalPairAndPreservesCredential(t *testing.T) {
	store, native, credential, binding := newProtectedFixture(t)
	original := bytes.Clone(native.raw)
	if _, err := store.Load(context.Background()); err != nil || !bytes.Equal(original, native.raw) || native.saves != 0 {
		t.Fatal("legacy load rewrote protected storage", err)
	}
	first, err := store.BeginPendingLink(context.Background(), binding)
	if err != nil || !first.valid() {
		t.Fatal("pending was not persisted", err)
	}
	restarted := &ProtectedStore{Native: native, Resource: store.Resource}
	again, err := restarted.BeginPendingLink(context.Background(), binding)
	if err != nil || again != first || native.saves != 1 {
		t.Fatal("retry replaced the original request or secret", err)
	}
	loaded, err := restarted.LoadPendingLink(context.Background(), binding.SessionID)
	if err != nil || loaded == nil || *loaded != first {
		t.Fatal("new store instance lost the pending request", err)
	}
	got, err := restarted.Load(context.Background())
	if err != nil || got.SessionID != credential.SessionID || got.AccessToken != credential.AccessToken {
		t.Fatal("pending write changed the current login", err)
	}
	if err = restarted.ClearPendingLink(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	got, err = restarted.Load(context.Background())
	loaded, pendingErr := restarted.LoadPendingLink(context.Background(), binding.SessionID)
	if err != nil || pendingErr != nil || loaded != nil || got.AccessToken != credential.AccessToken || native.deletes != 0 {
		t.Fatal("clearing pending damaged the credential")
	}
	second, err := restarted.BeginPendingLink(context.Background(), binding)
	if err != nil || second.RequestID == first.RequestID || second.RecoverySecret == first.RecoverySecret {
		t.Fatal("a subsequent request reused a completed pair", err)
	}
	if err = restarted.ClearPendingLink(context.Background(), first); err == nil {
		t.Fatal("late completion cleared a newer request")
	}
}

func TestProtectedCredentialCanInitializeEmptyOSRecord(t *testing.T) {
	store, native, credential, _ := newProtectedFixture(t)
	native.raw = nil
	if err := saveProtectedCredential(t, store, credential); err != nil {
		t.Fatal("initial login could not create the protected record", err)
	}
	record, err := decodeProtectedSession(native.raw)
	if err != nil || record.Credential.SessionID != credential.SessionID || record.PendingLink != nil || native.saves != 1 || native.deletes != 0 {
		t.Fatal("initial login did not store only the credential", err)
	}
}

func TestProtectedPendingCanRecoverAfterRestartWithoutDaemonBinding(t *testing.T) {
	store, native, credential, binding := newProtectedFixture(t)
	pending, err := store.BeginPendingLink(context.Background(), binding)
	if err != nil {
		t.Fatal(err)
	}
	restarted := &ProtectedStore{Native: native, Resource: store.Resource}
	// A restarted caller knows only its configured resource and protected login.
	// The original device and key binding must come from the saved record.
	known := credential.SessionID
	got, err := restarted.LoadPendingLink(context.Background(), known)
	if err != nil || got == nil || *got != pending {
		t.Fatal("receipt recovery requires a daemon or lost in-memory binding", err)
	}
	wrongResource := &ProtectedStore{Native: native, Resource: "https://other.example.test"}
	if _, err = wrongResource.LoadPendingLink(context.Background(), credential.SessionID); err == nil {
		t.Fatal("pending crossed the configured API resource")
	}
	if _, err = restarted.LoadPendingLink(context.Background(), ""); err == nil {
		t.Fatal("pending was read without a current session")
	}
}

func TestProtectedPendingRejectsBindingSubstitution(t *testing.T) {
	for _, field := range []string{"resource", "client", "session", "fingerprint", "device", "request", "secret"} {
		t.Run(field, func(t *testing.T) {
			store, native, _, binding := newProtectedFixture(t)
			pending, err := store.BeginPendingLink(context.Background(), binding)
			if err != nil {
				t.Fatal(err)
			}
			original := bytes.Clone(native.raw)
			changed := pending
			switch field {
			case "resource":
				changed.Resource = "https://other.example.test"
			case "client":
				changed.ClientID = "other-cli"
			case "session":
				changed.SessionID = "other-session"
			case "fingerprint":
				changed.KeyFingerprint = strings.Repeat("b", 64)
			case "device":
				changed.ExpectedDeviceID = "other-device"
			case "request":
				changed.RequestID = strings.Repeat("A", 43)
			case "secret":
				changed.RecoverySecret = strings.Repeat("A", 43)
			}
			if err = store.ClearPendingLink(context.Background(), changed); err == nil {
				t.Fatal("mismatched request cleared the original pending")
			}
			if changed.LinkBinding != binding {
				if _, err = store.BeginPendingLink(context.Background(), changed.LinkBinding); err == nil {
					t.Fatal("mismatched binding replaced pending")
				}
				if _, err = store.LoadPendingLink(context.Background(), changed.SessionID); changed.SessionID != binding.SessionID && err == nil {
					t.Fatal("mismatched binding read pending")
				}
			}
			if !bytes.Equal(original, native.raw) || native.saves != 1 || native.deletes != 0 {
				t.Fatal("rejected operation modified the protected record")
			}
		})
	}
}

func TestProtectedSavePreservesOnlySameSessionPending(t *testing.T) {
	store, native, credential, binding := newProtectedFixture(t)
	pending, err := store.BeginPendingLink(context.Background(), binding)
	if err != nil {
		t.Fatal(err)
	}
	credential.AccessToken, credential.RefreshToken = "rotated-access", "rotated-refresh"
	if err = saveProtectedCredential(t, store, credential); err != nil {
		t.Fatal(err)
	}
	got, err := store.LoadPendingLink(context.Background(), binding.SessionID)
	if err != nil || got == nil || *got != pending {
		t.Fatal("same-session rotation lost recovery material", err)
	}
	credential.SessionID = "replacement-login-session"
	if err = saveProtectedCredential(t, store, credential); err != nil {
		t.Fatal(err)
	}
	replaced := bytes.Clone(native.raw)
	if err = store.ClearPendingLink(context.Background(), pending); !errors.Is(err, userauth.ErrLoggedOut) {
		t.Fatal("stale clear crossed the login session", err)
	}
	if _, err = store.BeginPendingLink(context.Background(), binding); !errors.Is(err, userauth.ErrLoggedOut) {
		t.Fatal("old request rebound to a new login", err)
	}
	binding.SessionID = credential.SessionID
	got, err = store.LoadPendingLink(context.Background(), binding.SessionID)
	if err != nil || got != nil || !bytes.Equal(replaced, native.raw) {
		t.Fatal("new login retained the old pending request", err)
	}
}

func TestProtectedPendingSaveFailuresPreserveCustody(t *testing.T) {
	for _, committed := range []bool{false, true} {
		t.Run(fmt.Sprintf("committed=%t", committed), func(t *testing.T) {
			store, native, _, binding := newProtectedFixture(t)
			before := bytes.Clone(native.raw)
			native.saveErr, native.commitOnFail = securecredential.ErrUnavailable, committed
			result, err := store.BeginPendingLink(context.Background(), binding)
			if !errors.Is(err, securecredential.ErrUnavailable) || result != (PendingLink{}) || native.deletes != 0 {
				t.Fatal("failed save exposed or deleted recovery material", err)
			}
			var original *PendingLink
			if committed {
				record, decodeErr := decodeProtectedSession(native.raw)
				if decodeErr != nil || record.PendingLink == nil {
					t.Fatal("fixture did not commit the original record")
				}
				original = record.PendingLink
			} else if !bytes.Equal(before, native.raw) {
				t.Fatal("failed native save changed the original credential")
			}
			native.saveErr = nil
			retried, err := store.BeginPendingLink(context.Background(), binding)
			if err != nil || (original != nil && retried != *original) {
				t.Fatal("retry lost the committed original pair", err)
			}
		})
	}
}

func TestProtectedUnknownRecordAndOSFailuresNeverOverwriteOrDelete(t *testing.T) {
	for _, kind := range []string{"future", "null-version", "broken", "oversized", "locked", "cancelled", "unavailable"} {
		t.Run(kind, func(t *testing.T) {
			store, native, credential, binding := newProtectedFixture(t)
			switch kind {
			case "future":
				native.raw = []byte(`{"schema_version":"future-version","access_token":"legacy-looking","token_type":"Bearer"}`)
			case "null-version":
				native.raw = []byte(`{"schema_version":null,"access_token":"legacy-looking","token_type":"Bearer"}`)
			case "broken":
				native.raw = []byte(`{"schema_version":`)
			case "oversized":
				native.raw = bytes.Repeat([]byte(" "), protectedSessionLimit+1)
			case "locked":
				native.loadErr = securecredential.ErrInteractionRequired
			case "cancelled":
				native.loadErr = securecredential.ErrCancelled
			case "unavailable":
				native.loadErr = securecredential.ErrUnavailable
			}
			original := bytes.Clone(native.raw)
			if _, err := store.Load(context.Background()); err == nil {
				t.Fatal("unsafe record unexpectedly loaded")
			}
			if err := saveProtectedCredential(t, store, credential); err == nil {
				t.Fatal("unknown record or failed OS access was overwritten")
			}
			if _, err := store.BeginPendingLink(context.Background(), binding); err == nil {
				t.Fatal("pending was created without successful OS access")
			}
			if !bytes.Equal(original, native.raw) || native.saves != 0 || native.deletes != 0 {
				t.Fatal("error mutated protected custody")
			}
		})
	}
}

func TestProtectedRecordBoundAndDiagnosticsRedaction(t *testing.T) {
	store, native, credential, binding := newProtectedFixture(t)
	pending, err := store.BeginPendingLink(context.Background(), binding)
	if err != nil {
		t.Fatal(err)
	}
	record, err := decodeProtectedSession(native.raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []any{pending, &pending, binding, &binding, record, &record} {
		for _, format := range []string{"%v", "%+v", "%#v"} {
			printed := fmt.Sprintf(format, value)
			for _, secret := range []string{pending.RequestID, pending.RecoverySecret, credential.AccessToken, credential.RefreshToken, binding.SessionID} {
				if strings.Contains(printed, secret) {
					t.Fatal("secret appeared in ordinary diagnostic formatting")
				}
			}
		}
	}
	original := bytes.Clone(native.raw)
	credential.Scope = strings.Repeat("x", protectedSessionLimit)
	if err = saveProtectedCredential(t, store, credential); !errors.Is(err, securecredential.ErrUnavailable) || !bytes.Equal(original, native.raw) || native.saves != 1 {
		t.Fatal("oversized record reached native save", err)
	}
}

type protectedRoundTripper func(*http.Request) (*http.Response, error)

func (r protectedRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return r(request)
}

func protectedManager(t *testing.T, store *ProtectedStore, transport protectedRoundTripper) *userauth.Manager {
	t.Helper()
	client, err := userauth.NewClient(userauth.Config{ClientID: "riido-cli", Resource: protectedTestResource, Scopes: []string{"read"}, Endpoints: userauth.Endpoints{AuthorizeURL: protectedTestResource + "/oauth/authorize", TokenURL: protectedTestResource + "/oauth/token", RevokeURL: protectedTestResource + "/oauth/revoke"}, HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		t.Fatal(err)
	}
	return userauth.NewManager(client, store)
}

func protectedResponse(request *http.Request, body string) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}
}

func TestProtectedManagerRefreshAndLogoutPreserveSessionBoundary(t *testing.T) {
	store, native, credential, binding := newProtectedFixture(t)
	pending, err := store.BeginPendingLink(context.Background(), binding)
	if err != nil {
		t.Fatal(err)
	}
	credential.ExpiresAt = time.Now().Add(-time.Minute)
	if err = saveProtectedCredential(t, store, credential); err != nil {
		t.Fatal(err)
	}
	refreshes, revokes := 0, 0
	manager := protectedManager(t, store, func(request *http.Request) (*http.Response, error) {
		switch request.URL.Path {
		case "/oauth/token":
			refreshes++
			return protectedResponse(request, `{"access_token":"rotated","refresh_token":"rotated-refresh","token_type":"Bearer","expires_in":3600,"refresh_recovery":"v1"}`), nil
		case "/oauth/revoke":
			revokes++
			return protectedResponse(request, `{}`), nil
		default:
			t.Fatal("unexpected dependency on a daemon or unrelated API")
			return nil, errors.New("unexpected request")
		}
	})
	if token, tokenErr := manager.AccessToken(context.Background()); tokenErr != nil || token != "rotated" || refreshes != 1 {
		t.Fatal("ordinary standalone refresh failed", tokenErr)
	}
	got, err := store.LoadPendingLink(context.Background(), binding.SessionID)
	if err != nil || got == nil || *got != pending {
		t.Fatal("manager refresh discarded pending", err)
	}
	if err = manager.Logout(context.Background()); err != nil || revokes != 1 || len(native.raw) != 0 || native.deletes != 1 {
		t.Fatal("logout did not remove credential and pending together", err)
	}
	if err = store.ClearPendingLink(context.Background(), pending); !errors.Is(err, userauth.ErrNotAuthenticated) || len(native.raw) != 0 {
		t.Fatal("stale clear restored a logged-out credential", err)
	}
	if _, err = store.BeginPendingLink(context.Background(), binding); !errors.Is(err, userauth.ErrNotAuthenticated) || len(native.raw) != 0 {
		t.Fatal("stale create restored a logged-out credential", err)
	}
}

func TestProtectedPendingAndLogoutSerializeWithRefresh(t *testing.T) {
	for _, operation := range []string{"begin", "logout"} {
		t.Run(operation, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			store, native, credential, binding := newProtectedFixture(t)
			if operation == "logout" {
				if _, err := store.BeginPendingLink(ctx, binding); err != nil {
					t.Fatal(err)
				}
			}
			credential.ExpiresAt = time.Now().Add(-time.Minute)
			if err := saveProtectedCredential(t, store, credential); err != nil {
				t.Fatal(err)
			}
			native.lockAttempt = make(chan struct{}, 8)
			entered, release := make(chan struct{}), make(chan struct{})
			manager := protectedManager(t, store, func(request *http.Request) (*http.Response, error) {
				if request.URL.Path == "/oauth/token" {
					close(entered)
					select {
					case <-release:
					case <-ctx.Done():
						return nil, ctx.Err()
					}
					return protectedResponse(request, `{"access_token":"rotated","refresh_token":"rotated-refresh","token_type":"Bearer","expires_in":3600,"refresh_recovery":"v1"}`), nil
				}
				if request.URL.Path != "/oauth/revoke" {
					return nil, errors.New("unexpected endpoint")
				}
				return protectedResponse(request, `{}`), nil
			})
			refreshed := make(chan error, 1)
			go func() { _, err := manager.AccessToken(ctx); refreshed <- err }()
			waitProtectedSignal(t, ctx, entered)
			waitProtectedSignal(t, ctx, native.lockAttempt) // refresh owns the lock
			finished := make(chan error, 1)
			go func() {
				if operation == "logout" {
					finished <- manager.Logout(ctx)
					return
				}
				_, err := store.BeginPendingLink(ctx, binding)
				finished <- err
			}()
			waitProtectedSignal(t, ctx, native.lockAttempt) // queued behind refresh
			close(release)
			refreshErr := <-refreshed
			if err := <-finished; err != nil {
				t.Fatal("serialized operation failed", err)
			}
			if operation == "logout" {
				if !errors.Is(refreshErr, userauth.ErrLoggedOut) || len(native.raw) != 0 || native.deletes != 1 {
					t.Fatal("stale refresh restored logged-out credential or pending", refreshErr)
				}
				return
			}
			if refreshErr != nil {
				t.Fatal(refreshErr)
			}
			record, err := decodeProtectedSession(native.raw)
			if err != nil || record.Credential.AccessToken != "rotated" || record.PendingLink == nil || record.PendingLink.LinkBinding != binding {
				t.Fatal("pending operation overwrote the rotated credential", err)
			}
		})
	}
}

func waitProtectedSignal(t *testing.T, ctx context.Context, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
	case <-ctx.Done():
		t.Fatal("protected store synchronization timed out")
	}
}
