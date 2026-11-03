package standalonecli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

// accessNativeFixture adds the RIID-6014 access item to the protected fixture.
type accessNativeFixture struct {
	*protectedNativeFixture
	access    []byte
	deleteErr error
	events    []string
}

func (n *accessNativeFixture) LoadAccess(context.Context) ([]byte, error) {
	if len(n.access) == 0 {
		return nil, securecredential.ErrNotFound
	}
	return bytes.Clone(n.access), nil
}

func (n *accessNativeFixture) SaveAccess(_ context.Context, raw []byte) error {
	n.events = append(n.events, "access-save")
	n.access = bytes.Clone(raw)
	return nil
}

func (n *accessNativeFixture) DeleteAccess(context.Context) error {
	n.events = append(n.events, "access-delete")
	if n.deleteErr != nil {
		return n.deleteErr
	}
	n.access = nil
	return nil
}

func (n *accessNativeFixture) Delete(ctx context.Context) error {
	n.events = append(n.events, "delete")
	return n.protectedNativeFixture.Delete(ctx)
}

func newAccessFixture(t *testing.T) (*ProtectedStore, *accessNativeFixture, userauth.Credential) {
	t.Helper()
	_, protected, credential, _ := newProtectedFixture(t)
	native := &accessNativeFixture{protectedNativeFixture: protected}
	return &ProtectedStore{Native: native, Resource: protectedTestResource}, native, credential
}

func TestProtectedSaveKeepsAccessCacheInStep(t *testing.T) {
	store, native, credential := newAccessFixture(t)
	if err := saveProtectedCredential(t, store, credential); err != nil {
		t.Fatal(err)
	}
	grant, err := store.LoadAccess(context.Background())
	if err != nil || grant.SessionID != credential.SessionID || grant.AccessToken != credential.AccessToken {
		t.Fatalf("cache not written: %v", err)
	}
	if bytes.Contains(native.access, []byte(credential.RefreshToken)) {
		t.Fatal("refresh token reached the access cache")
	}
	pending := credential
	pending.RefreshRecoverySecret = "synthetic-recovery"
	if err = saveProtectedCredential(t, store, pending); err != nil {
		t.Fatal(err)
	}
	if _, err = store.LoadAccess(context.Background()); !errors.Is(err, userauth.ErrNotAuthenticated) {
		t.Fatalf("pending rotation must clear the cache: %v", err)
	}
}

func TestProtectedDeleteClearsAccessCacheFirst(t *testing.T) {
	store, native, credential := newAccessFixture(t)
	if err := saveProtectedCredential(t, store, credential); err != nil {
		t.Fatal(err)
	}
	native.events = nil
	if err := store.Delete(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(native.events) != 2 || native.events[0] != "access-delete" || native.events[1] != "delete" || native.access != nil {
		t.Fatalf("delete order %v", native.events)
	}
	for name, deleteErr := range map[string]error{"unsupported helper": securecredential.ErrUnsupported, "not found": securecredential.ErrNotFound} {
		store, native, _ := newAccessFixture(t)
		native.deleteErr = deleteErr
		if err := store.Delete(context.Background()); err != nil || native.raw != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	store, native, _ = newAccessFixture(t)
	native.deleteErr = securecredential.ErrUnavailable
	if err := store.Delete(context.Background()); err == nil || native.raw == nil {
		t.Fatal("a cache that cannot be cleared must stop the delete")
	}
}

func TestProtectedAccessCacheRejectsForeignOrMalformedRecords(t *testing.T) {
	store, native, credential := newAccessFixture(t)
	if err := saveProtectedCredential(t, store, credential); err != nil {
		t.Fatal(err)
	}
	native.access = []byte(`{"schema_version":"riido-cli-access.v1","grant":{"session_id":"s","access_token":"a"},"extra":1}`)
	if _, err := store.LoadAccess(context.Background()); err == nil {
		t.Fatal("unknown field accepted")
	}
	desktop := &ProtectedStore{Native: native, Resource: protectedTestResource, AuthOrigin: userauth.AuthOriginDesktop}
	if err := store.SaveAccess(context.Background(), credential.AccessGrant()); err != nil {
		t.Fatal(err)
	}
	if _, err := desktop.LoadAccess(context.Background()); err == nil {
		t.Fatal("browser-origin grant accepted by the desktop store")
	}
	legacy := &ProtectedStore{Native: native.protectedNativeFixture, Resource: protectedTestResource}
	if _, err := legacy.LoadAccess(context.Background()); err == nil {
		t.Fatal("native without the access item must miss")
	}
}
