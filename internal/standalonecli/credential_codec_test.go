package standalonecli

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestNativeStorageCodecPreservesLoginSessionIdentity(t *testing.T) {
	credential := userauth.Credential{SessionID: "new-login-id", AccessToken: "fake-access", RefreshToken: "fake-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}
	raw, err := encodeCredential(credential)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := decodeCredential(raw)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SessionID != credential.SessionID || loaded.AccessToken != credential.AccessToken || loaded.RefreshToken != credential.RefreshToken {
		t.Fatal("native storage JSON lost login identity")
	}
}
