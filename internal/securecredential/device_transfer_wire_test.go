package securecredential

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestDeviceTransferMatchesCurrentCPAndNativeFrame(t *testing.T) {
	_, proof, _ := transferFixture(t)
	proof.Resource = "https://cli.fixture.invalid"
	proof.KeyFingerprint = strings.Repeat("b", 64)
	proof.ExpiresAt = "2026-10-28T01:02:03.123Z"
	digest := sha256.Sum256(proof.framedMessage())
	// Executed from CP 0e7515de cliInstallationLinkMessage; Swift uses this vector too.
	if hex.EncodeToString(digest[:]) != "f1dc69e70077293465bd980931f9e52eb55796fbf6793b537cc53a0180cf441c" {
		t.Fatal("transfer wire differs from CP contract")
	}
}
