package standalonecli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestProtectedOriginFactoryKeepsLegacyKeysAndSharedLocks(t *testing.T) {
	lockDir := filepath.Join(t.TempDir(), "private")
	if err := os.Mkdir(lockDir, 0o700); err != nil {
		t.Fatal(err)
	}
	base := securecredential.Store{HelperPath: "/fixture/helper", Interactive: true, LockDir: lockDir}
	browser, err := NewProtectedStore(base, protectedTestResource, "browser")
	if err != nil {
		t.Fatal(err)
	}
	desktop, err := NewProtectedStore(base, protectedTestResource, "desktop")
	if err != nil {
		t.Fatal(err)
	}
	old := sha256.Sum256([]byte("riido-cli:" + protectedTestResource))
	oldRepair := sha256.Sum256([]byte("riido-cli-repair:" + protectedTestResource))
	bn, dn := browser.Native.(*securecredential.Store), desktop.Native.(*securecredential.Store)
	br, dr := browser.RepairNative.(*securecredential.Store), desktop.RepairNative.(*securecredential.Store)
	if bn.Key != hex.EncodeToString(old[:]) || br.Key != hex.EncodeToString(oldRepair[:]) || dn.Key == bn.Key || dr.Key == br.Key || dn.Key == dr.Key {
		t.Fatal("credential/repair namespace changed or collided")
	}
	if bn.HelperPath != dn.HelperPath || !dn.Interactive || bn.LockDir != dn.LockDir {
		t.Fatal("origin separation changed OS protection")
	}
	unlock, err := browser.Lock(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 40*time.Millisecond)
	defer cancel()
	if other, err := desktop.Lock(ctx); err == nil {
		other()
		t.Error("origins do not share credential lock")
	}
	unlock()
	unlock, err = desktop.LockRepairOperation(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if other, err := browser.LockRepairOperation(t.Context()); err == nil {
		other()
		t.Error("same installation has parallel repair operations")
	}
	unlock()
	for _, origin := range []string{"Desktop", "unknown", " browser"} {
		if _, err := NewProtectedStore(base, protectedTestResource, origin); err == nil {
			t.Fatal("invalid origin fell back to browser")
		}
	}
}

func originStores(t *testing.T) (*ProtectedStore, *ProtectedStore, *protectedNativeFixture, *protectedNativeFixture) {
	t.Helper()
	b, bn, _, _ := newProtectedFixture(t)
	d, dn, _, _ := newProtectedFixture(t)
	bn.raw, dn.raw = nil, nil
	br, dr := &protectedNativeFixture{}, &protectedNativeFixture{}
	b.RepairNative, d.RepairNative = br, dr
	b.CustodyLock, d.CustodyLock = bn, bn
	b.RepairLock, d.RepairLock = br, br
	b.AuthOrigin, d.AuthOrigin = "browser", "desktop"
	return b, d, bn, dn
}

func TestProtectedOriginLogoutDoesNotDeleteOtherLoggedOutRepair(t *testing.T) {
	for _, logoutOrigin := range []string{"browser", "desktop"} {
		t.Run(logoutOrigin, func(t *testing.T) {
			b, d, _, _ := originStores(t)
			for _, s := range []*ProtectedStore{b, d} {
				if _, err := s.BeginPendingRepair(t.Context(), testRepairFingerprint()); err != nil {
					t.Fatal(err)
				}
			}
			own, other := b, d
			if logoutOrigin == "desktop" {
				own, other = d, b
			}
			before := bytes.Clone(other.RepairNative.(*protectedNativeFixture).raw)
			unlock, err := own.Lock(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			err = own.ClearLocalRequestsOnLogout(t.Context())
			unlock()
			if err != nil || !bytes.Equal(before, other.RepairNative.(*protectedNativeFixture).raw) {
				t.Fatal("logout erased other origin's repair", err)
			}
			if p, err := other.LoadPendingRepair(t.Context()); err != nil || p == nil || p.SessionID != "" {
				t.Fatal("logged-out repair lost custody", err)
			}
		})
	}
}

func TestProtectedOriginSameSessionBindingCannotChange(t *testing.T) {
	b, _, n, _ := originStores(t)
	_, _, c, link := newProtectedFixture(t)
	c.Binding = custodyBinding("browser")
	if err := saveProtectedCredential(t, b, c); err != nil {
		t.Fatal(err)
	}
	if _, err := b.BeginPendingLink(t.Context(), link); err != nil {
		t.Fatal(err)
	}
	original := bytes.Clone(n.raw)
	for _, mutate := range []func(*userauth.CredentialBinding){
		func(x *userauth.CredentialBinding) { x.AccountID = "another-account" },
		func(x *userauth.CredentialBinding) { x.AuthOrigin = "desktop" },
		func(x *userauth.CredentialBinding) { x.Resource = "https://other.test" },
		func(x *userauth.CredentialBinding) { x.KeyFingerprint = testRepairFingerprint() },
	} {
		next := c
		binding := *c.Binding
		next.Binding = &binding
		mutate(next.Binding)
		if userauth.SameCredentialBinding(next.Binding, c.Binding) {
			next.Binding.KeyFingerprint = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		}
		if err := saveProtectedCredential(t, b, next); err == nil || !bytes.Equal(original, n.raw) {
			t.Fatal("same session rebound or damaged old state")
		}
	}
	next := c
	next.Binding = nil
	if err := saveProtectedCredential(t, b, next); err == nil || !bytes.Equal(original, n.raw) {
		t.Fatal("same session downgraded to legacy")
	}
	if _, err := b.Load(t.Context()); errors.Is(err, userauth.ErrNotAuthenticated) {
		t.Fatal("rejected binding erased login")
	}
}
