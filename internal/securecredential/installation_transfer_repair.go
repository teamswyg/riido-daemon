package securecredential

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
)

// InstallationTransferRepairProof admits only a new human-consent request for an
// existing key. It carries a recovery hash, never the secret or arbitrary bytes.
type InstallationTransferRepairProof struct {
	Resource          string `json:"resource"`
	ClientID          string `json:"client_id"`
	RequestID         string `json:"request_id"`
	OriginalRequestID string `json:"original_request_id"`
	KeyFingerprint    string `json:"key_fingerprint"`
	RecoveryHash      string `json:"recovery_hash"`
}

func (InstallationTransferRepairProof) String() string {
	return "[redacted installation transfer repair proof]"
}

func (InstallationTransferRepairProof) GoString() string {
	return "[redacted installation transfer repair proof]"
}

func (p InstallationTransferRepairProof) message(key InstallationPublicKey) ([]byte, error) {
	if _, err := key.publicKey(); err != nil {
		return nil, err
	}
	if !installationResource(p.Resource) || !strings.HasPrefix(p.Resource, "https://") || p.ClientID != "riido-cli" || p.KeyFingerprint != key.KeyFingerprint {
		return nil, ErrUnavailable
	}
	if _, err := installationURLBytes(p.RequestID, 32); err != nil {
		return nil, err
	}
	if _, err := installationURLBytes(p.OriginalRequestID, 32); err != nil || p.OriginalRequestID == p.RequestID {
		return nil, ErrUnavailable
	}
	fingerprint, err := installationHash(p.KeyFingerprint)
	if err != nil {
		return nil, err
	}
	recovery, err := installationHash(p.RecoveryHash)
	if err != nil {
		return nil, err
	}
	return installationFrame([]byte("riido-cli-installation-transfer-repair-request.v1"), []byte(p.Resource), []byte(p.ClientID), []byte(p.RequestID), []byte(p.OriginalRequestID), fingerprint, recovery), nil
}

// The helper loads only the existing installation key and authenticates the OS
// user. Unknown/older operations cannot fall back to another signing domain.
func (s *Store) SignInstallationTransferRepairProof(ctx context.Context, key InstallationPublicKey, proof InstallationTransferRepairProof) (string, error) {
	if s == nil {
		return "", ErrUnavailable
	}
	message, err := proof.message(key)
	if err != nil {
		return "", err
	}
	data, err := s.installationCall(ctx, proof.Resource, "installation-transfer-repair-proof", proof)
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
	return result.Signature, nil
}
