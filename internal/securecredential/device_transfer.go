package securecredential

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math"
	"strconv"
	"time"
)

// DeviceTransferProof binds a server-approved transfer to the exact source link.
// It never changes a key, saved login, device or link by itself.
type DeviceTransferProof struct {
	DeviceLinkProof
	PreviousDeviceID     string `json:"previous_device_id"`
	PreviousLinkRevision string `json:"previous_link_revision"`
}

func (DeviceTransferProof) String() string   { return "[redacted device transfer proof]" }
func (DeviceTransferProof) GoString() string { return "[redacted device transfer proof]" }

func (p DeviceTransferProof) message(key InstallationPublicKey, now time.Time) ([]byte, error) {
	if _, err := p.DeviceLinkProof.message(key, now); err != nil {
		return nil, err
	}
	revision, err := strconv.ParseUint(p.PreviousLinkRevision, 10, 64)
	if err != nil || revision == 0 || revision == math.MaxUint64 || strconv.FormatUint(revision, 10) != p.PreviousLinkRevision ||
		!deviceLinkID(p.PreviousDeviceID) || p.PreviousDeviceID == p.ExpectedDeviceID {
		return nil, ErrUnavailable
	}
	if _, err := installationURLBytes(p.IntentID, 32); err != nil {
		return nil, err
	}
	return p.framedMessage(), nil
}

func (p DeviceTransferProof) framedMessage() []byte {
	return installationFrame([]byte("riido-cli-device-transfer.v1"), []byte("transfer"), []byte(p.Resource), []byte(p.ClientID),
		[]byte(p.InstallationID), []byte(p.IntentID), []byte(p.ChallengeHash), []byte(p.KeyFingerprint), []byte(p.ExpectedDeviceID),
		[]byte(p.ApprovedPrincipalID), []byte(p.ExpiresAt), []byte(p.PreviousDeviceID), []byte(p.PreviousLinkRevision))
}

func (s *Store) SignDeviceTransferProof(ctx context.Context, key InstallationPublicKey, proof DeviceTransferProof) (string, error) {
	message, err := proof.message(key, time.Now())
	if err != nil {
		return "", err
	}
	data, err := s.installationCall(ctx, proof.Resource, "installation-device-transfer-proof", proof)
	if err != nil {
		return "", err
	}
	if _, err := installationObject(data, "algorithm", "signature"); err != nil {
		return "", err
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
	public, err := key.publicKey()
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(message)
	if !ecdsa.VerifyASN1(public, digest[:], raw) {
		return "", ErrUnavailable
	}
	if _, err = proof.message(key, time.Now()); err != nil {
		return "", err
	}
	return result.Signature, nil
}
