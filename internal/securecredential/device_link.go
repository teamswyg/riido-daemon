package securecredential

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"
)

// DeviceLinkProof is the fixed control-plane installation link intent. The
// caller must obtain it through the authenticated server flow. Signing it does
// not bypass server owner, current policy, intent expiry or consumption checks.
type DeviceLinkProof struct {
	Resource            string `json:"resource"`
	ClientID            string `json:"client_id"`
	InstallationID      string `json:"installation_id"`
	IntentID            string `json:"intent_id"`
	ChallengeHash       string `json:"challenge_hash"`
	KeyFingerprint      string `json:"key_fingerprint"`
	ExpectedDeviceID    string `json:"expected_device_id"`
	ApprovedPrincipalID string `json:"approved_principal_id"`
	ExpiresAt           string `json:"expires_at"`
}

func (DeviceLinkProof) String() string   { return "[redacted device link proof]" }
func (DeviceLinkProof) GoString() string { return "[redacted device link proof]" }

func deviceLinkID(value string) bool {
	return value != "" && len(value) <= 512 && utf8.ValidString(value) && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "/\\\x00\r\n\t")
}

func (p DeviceLinkProof) message(key InstallationPublicKey, now time.Time) ([]byte, error) {
	if _, err := key.publicKey(); err != nil {
		return nil, err
	}
	expiry, err := time.Parse(time.RFC3339Nano, p.ExpiresAt)
	if err != nil || expiry.UTC().Format(time.RFC3339Nano) != p.ExpiresAt || !expiry.After(now) ||
		!installationResource(p.Resource) || !strings.HasPrefix(p.Resource, "https://") || p.ClientID != "riido-cli" || p.KeyFingerprint != key.KeyFingerprint {
		return nil, ErrUnavailable
	}
	for _, id := range []string{p.InstallationID, p.IntentID, p.ExpectedDeviceID, p.ApprovedPrincipalID} {
		if !deviceLinkID(id) {
			return nil, ErrUnavailable
		}
	}
	if _, err := installationHash(p.ChallengeHash); err != nil {
		return nil, err
	}
	return p.framedMessage(), nil
}

func (p DeviceLinkProof) framedMessage() []byte {
	return installationFrame([]byte("riido-cli-device-link.v1"), []byte("link"), []byte(p.Resource), []byte(p.ClientID), []byte(p.InstallationID), []byte(p.IntentID),
		[]byte(p.ChallengeHash), []byte(p.KeyFingerprint), []byte(p.ExpectedDeviceID), []byte(p.ApprovedPrincipalID), []byte(p.ExpiresAt))
}

// SignDeviceLinkProof reuses an existing protected installation key without
// starting daemon or reading user credentials. The native helper authenticates
// the OS user. A missing key, expired proof or invalid helper result is denied.
func (s *Store) SignDeviceLinkProof(ctx context.Context, key InstallationPublicKey, proof DeviceLinkProof) (string, error) {
	message, err := proof.message(key, time.Now())
	if err != nil {
		return "", err
	}
	data, err := s.installationCall(ctx, proof.Resource, "installation-device-link-proof", proof)
	if err != nil {
		return "", err
	}
	fields, err := installationObject(data, "algorithm", "signature")
	if err != nil || len(fields) != 2 {
		return "", ErrUnavailable
	}
	var result struct {
		Algorithm string `json:"algorithm"`
		Signature string `json:"signature"`
	}
	if json.Unmarshal(data, &result) != nil || result.Algorithm != InstallationKeyAlgorithm || len(result.Signature) > 96 {
		return "", ErrUnavailable
	}
	raw, err := base64.RawURLEncoding.DecodeString(result.Signature)
	if err != nil || len(raw) > 72 || base64.RawURLEncoding.EncodeToString(raw) != result.Signature {
		return "", ErrUnavailable
	}
	publicKey, err := key.publicKey()
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(message)
	if !ecdsa.VerifyASN1(publicKey, digest[:], raw) {
		return "", ErrUnavailable
	}
	if _, err = proof.message(key, time.Now()); err != nil {
		return "", err
	}
	return result.Signature, nil
}
