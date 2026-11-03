package standalonecli

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type linkKeyFixture struct {
	*protectedNativeFixture
	key                      securecredential.InstallationPublicKey
	keyErr                   error
	existing, signs, ensures int
	proof                    securecredential.DeviceLinkProof
	loads                    atomic.Int32
}

func (n *linkKeyFixture) Load(ctx context.Context) ([]byte, error) {
	n.loads.Add(1)
	return n.protectedNativeFixture.Load(ctx)
}

func (n *linkKeyFixture) ExistingInstallationKey(context.Context, string) (securecredential.InstallationPublicKey, error) {
	n.existing++
	return n.key, n.keyErr
}

func (n *linkKeyFixture) EnsureInstallationKey(context.Context, string) (securecredential.InstallationPublicKey, error) {
	n.ensures++
	return securecredential.InstallationPublicKey{}, securecredential.ErrUnavailable
}

func (n *linkKeyFixture) SignDeviceLinkProof(_ context.Context, _ securecredential.InstallationPublicKey, p securecredential.DeviceLinkProof) (string, error) {
	n.signs++
	n.proof = p
	return "synthetic-signature", nil
}

func TestProtectedLinkKeyAdapterUsesExistingIdentityAndFixedProof(t *testing.T) {
	s, n, _, b := newProtectedFixture(t)
	native := &linkKeyFixture{protectedNativeFixture: n, key: securecredential.InstallationPublicKey{Algorithm: "test-algorithm", PublicKey: "synthetic-public-only", KeyFingerprint: b.KeyFingerprint}}
	s.Native = native
	key, err := s.ExistingInstallationKey(t.Context(), s.Resource)
	if err != nil || key.KeyFingerprint != b.KeyFingerprint || native.existing != 1 || native.ensures != 0 {
		t.Fatal("link lookup used key creation", err)
	}
	proof := userauth.DeviceLinkProof{Resource: s.Resource, ClientID: "riido-cli", KeyFingerprint: b.KeyFingerprint, ExpectedDeviceID: b.ExpectedDeviceID, IntentID: "original-intent"}
	signature, err := s.SignDeviceLinkProof(t.Context(), key, proof)
	if err != nil || signature == "" || native.signs != 1 || native.proof != securecredential.DeviceLinkProof(proof) {
		t.Fatal("typed proof changed across adapter", err)
	}
	if _, err = s.ExistingInstallationKey(t.Context(), "https://foreign.example.test"); err == nil || native.existing != 1 {
		t.Fatal("foreign resource queried native key")
	}
	proof.ExpectedDeviceID = "same-device"
	proof.KeyFingerprint = "foreign-key"
	if _, err = s.SignDeviceLinkProof(t.Context(), key, proof); err == nil || native.signs != 1 {
		t.Fatal("foreign key reached signer")
	}
	if n.saves != 0 || n.deletes != 0 {
		t.Fatal("key adapter changed OAuth custody")
	}
}

func TestProtectedLinkMissingKeyDoesNotBecomeMissingOAuthCredential(t *testing.T) {
	s, n, _, _ := newProtectedFixture(t)
	native := &linkKeyFixture{protectedNativeFixture: n, keyErr: securecredential.ErrNotFound}
	s.Native = native
	if _, err := s.ExistingInstallationKey(t.Context(), s.Resource); !errors.Is(err, securecredential.ErrNotFound) || errors.Is(err, userauth.ErrNotAuthenticated) || native.ensures != 0 {
		t.Fatal("missing key fell back or reclassified OAuth", err)
	}
	if n.saves != 0 || n.deletes != 0 {
		t.Fatal("missing key removed current credential")
	}
	s.Native = n
	if _, err := s.ExistingInstallationKey(t.Context(), s.Resource); !errors.Is(err, securecredential.ErrUnavailable) {
		t.Fatal("older native adapter became supported", err)
	}
}
