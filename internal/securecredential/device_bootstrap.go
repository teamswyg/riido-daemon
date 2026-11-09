package securecredential

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"
)

// DeviceBootstrapProof is a fixed server intent. The current boot digest is
// deliberately absent: only the verified native helper may supply it.
type DeviceBootstrapProof struct {
	DeviceLinkProof
	Phase              string `json:"phase"`
	TransitionID       string `json:"transition_id"`
	PreviousBootDigest string `json:"previous_boot_digest"`
}

type DeviceBootstrapSignature struct {
	BootDigest string `json:"boot_digest"`
	Signature  string `json:"signature"`
}

func (DeviceBootstrapProof) String() string       { return "[redacted device bootstrap proof]" }
func (DeviceBootstrapProof) GoString() string     { return "[redacted device bootstrap proof]" }
func (DeviceBootstrapSignature) String() string   { return "[redacted device bootstrap signature]" }
func (DeviceBootstrapSignature) GoString() string { return "[redacted device bootstrap signature]" }

func (p DeviceBootstrapProof) validate(key InstallationPublicKey, now time.Time) error {
	if _, err := p.DeviceLinkProof.message(key, now); err != nil {
		return err
	}
	for _, id := range []string{p.TransitionID, p.IntentID} {
		if _, err := installationURLBytes(id, 32); err != nil {
			return err
		}
	}
	switch p.Phase {
	case "prepare":
		if p.PreviousBootDigest != "" {
			return ErrUnavailable
		}
	case "activate":
		if _, err := installationHash(p.PreviousBootDigest); err != nil {
			return err
		}
	default:
		return ErrUnavailable
	}
	return nil
}

func (p DeviceBootstrapProof) message(key InstallationPublicKey, bootDigest string, now time.Time) ([]byte, error) {
	if err := p.validate(key, now); err != nil {
		return nil, err
	}
	if _, err := installationHash(bootDigest); err != nil {
		return nil, err
	}
	if p.Phase == "activate" && p.PreviousBootDigest == bootDigest {
		return nil, ErrUnavailable
	}
	return p.framedMessage(bootDigest), nil
}

func (p DeviceBootstrapProof) framedMessage(bootDigest string) []byte {
	return installationFrame([]byte("riido-cli-device-bootstrap.v1"), []byte(p.Phase), []byte(p.Resource), []byte(p.ClientID),
		[]byte(p.InstallationID), []byte(p.TransitionID), []byte(p.IntentID), []byte(p.ChallengeHash), []byte(p.KeyFingerprint),
		[]byte(p.ExpectedDeviceID), []byte(p.ApprovedPrincipalID), []byte(p.ExpiresAt), []byte(p.PreviousBootDigest), []byte(bootDigest))
}

// SignDeviceBootstrapProof does not change a credential, key or device. Server
// validation, fencing and a durable transition remain necessary after signing.
func (s *Store) SignDeviceBootstrapProof(ctx context.Context, key InstallationPublicKey, proof DeviceBootstrapProof) (DeviceBootstrapSignature, error) {
	if err := proof.validate(key, time.Now()); err != nil {
		return DeviceBootstrapSignature{}, err
	}
	data, err := s.installationCall(ctx, proof.Resource, "installation-device-bootstrap-proof", proof)
	if err != nil {
		return DeviceBootstrapSignature{}, err
	}
	if _, err := installationObject(data, "algorithm", "signature", "boot_digest"); err != nil {
		return DeviceBootstrapSignature{}, err
	}
	var result struct {
		Algorithm string `json:"algorithm"`
		DeviceBootstrapSignature
	}
	if json.Unmarshal(data, &result) != nil || result.Algorithm != InstallationKeyAlgorithm || len(result.Signature) > 96 {
		return DeviceBootstrapSignature{}, ErrUnavailable
	}
	message, err := proof.message(key, result.BootDigest, time.Now())
	if err != nil {
		return DeviceBootstrapSignature{}, err
	}
	raw, err := base64.RawURLEncoding.DecodeString(result.Signature)
	if err != nil || len(raw) > 72 || base64.RawURLEncoding.EncodeToString(raw) != result.Signature {
		return DeviceBootstrapSignature{}, ErrUnavailable
	}
	public, err := key.publicKey()
	if err != nil {
		return DeviceBootstrapSignature{}, err
	}
	digest := sha256.Sum256(message)
	if !ecdsa.VerifyASN1(public, digest[:], raw) {
		return DeviceBootstrapSignature{}, ErrUnavailable
	}
	return result.DeviceBootstrapSignature, nil
}
