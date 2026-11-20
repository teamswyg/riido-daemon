package securecredential

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestDeviceBootstrapMatchesNativeAndIndependentFrame(t *testing.T) {
	_, proof, _ := bootstrapFixture(t)
	proof.Resource = "https://cli.fixture.invalid"
	proof.KeyFingerprint = strings.Repeat("b", 64)
	proof.ExpiresAt = "2026-10-28T01:02:03.123Z"
	got := sha256.Sum256(proof.framedMessage(strings.Repeat("c", 64)))
	// Python uint32 big-endian framing and the native Swift fixture share this vector.
	if hex.EncodeToString(got[:]) != "3cc36ae5f4cf0c3264b303853809ae2d0085eb466e3eedda70651ca17c06d009" {
		t.Fatal("bootstrap fixed domain or field order changed")
	}
}
