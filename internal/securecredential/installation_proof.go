package securecredential

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"net/url"
	"strconv"
	"time"
)

const InstallationKeyAlgorithm = "p256-sha256-x962-der-v1"

type InstallationPublicKey struct {
	Algorithm      string `json:"algorithm"`
	PublicKey      string `json:"public_key"`
	KeyFingerprint string `json:"key_fingerprint"`
}

// InstallationProof is the fixed read-only OAuth authorization proof. It is not
// an arbitrary signing interface. Code/state never go to command output or logs.
type InstallationProof struct {
	Resource       string `json:"resource"`
	ClientID       string `json:"client_id"`
	Intent         string `json:"installation_intent"`
	ChallengeHash  string `json:"challenge_hash"`
	KeyFingerprint string `json:"key_fingerprint"`
	Code           string `json:"code"`
	RedirectURI    string `json:"redirect_uri"`
	CodeChallenge  string `json:"code_challenge"`
	State          string `json:"state"`
	ExpiresAt      int64  `json:"expires_at"`
}

func installationResource(resource string) bool {
	u, err := url.Parse(resource)
	if err != nil || len(resource) > 2048 || u.Hostname() == "" || u.User != nil || u.Opaque != "" ||
		u.Path != "" || u.RawPath != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.String() != resource {
		return false
	}
	if u.Port() != "" {
		port, parseErr := strconv.Atoi(u.Port())
		if parseErr != nil || port < 1 || port > 65535 {
			return false
		}
	}
	return u.Scheme == "https" || u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1")
}

func installationRedirect(value string) bool {
	u, err := url.Parse(value)
	if err != nil || len(value) > 2048 || u.Scheme != "http" || (u.Hostname() != "127.0.0.1" && u.Hostname() != "::1") ||
		u.User != nil || u.Opaque != "" || u.Path != "/callback" || u.RawPath != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.String() != value {
		return false
	}
	port, err := strconv.Atoi(u.Port())
	return err == nil && port > 0 && port <= 65535
}

func installationFrame(fields ...[]byte) []byte {
	var result []byte
	for _, field := range fields {
		result = binary.BigEndian.AppendUint32(result, uint32(len(field)))
		result = append(result, field...)
	}
	return result
}

func installationStorageKey(resource string) string {
	digest := sha256.Sum256(installationFrame([]byte("riido-cli-installation-key.v1"), []byte(resource), []byte("riido-cli")))
	return hex.EncodeToString(digest[:])
}

func installationURLBytes(value string, size int) ([]byte, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(raw) != size || base64.RawURLEncoding.EncodeToString(raw) != value {
		return nil, ErrUnavailable
	}
	return raw, nil
}

func installationHash(value string) ([]byte, error) {
	raw, err := hex.DecodeString(value)
	if err != nil || len(raw) != sha256.Size || hex.EncodeToString(raw) != value {
		return nil, ErrUnavailable
	}
	return raw, nil
}

func (k InstallationPublicKey) publicKey() (*ecdsa.PublicKey, error) {
	raw, err := installationURLBytes(k.PublicKey, 65)
	if err != nil || k.Algorithm != InstallationKeyAlgorithm {
		return nil, ErrUnavailable
	}
	key, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), raw)
	digest := sha256.Sum256(raw)
	if err != nil || hex.EncodeToString(digest[:]) != k.KeyFingerprint {
		return nil, ErrUnavailable
	}
	return key, nil
}

func (p InstallationProof) message(k InstallationPublicKey, now time.Time) ([]byte, error) {
	if _, err := k.publicKey(); err != nil {
		return nil, err
	}
	if !installationResource(p.Resource) || p.ClientID != "riido-cli" || !installationRedirect(p.RedirectURI) ||
		p.Code == "" || len(p.Code) > 256 || p.KeyFingerprint != k.KeyFingerprint || p.ExpiresAt <= 0 || !time.Unix(p.ExpiresAt, 0).After(now) {
		return nil, ErrUnavailable
	}
	for _, value := range []string{p.Intent, p.State, p.CodeChallenge} {
		if _, err := installationURLBytes(value, 32); err != nil {
			return nil, err
		}
	}
	challenge, err := installationHash(p.ChallengeHash)
	if err != nil {
		return nil, err
	}
	keyDigest, err := installationHash(p.KeyFingerprint)
	if err != nil {
		return nil, err
	}
	codeHash, stateHash := sha256.Sum256([]byte(p.Code)), sha256.Sum256([]byte(p.State))
	return installationFrame([]byte("riido-cli-installation-proof.v1"), []byte("authorization_code"), []byte(p.Resource),
		[]byte(p.ClientID), []byte(p.Intent), challenge, keyDigest, codeHash[:], []byte(p.RedirectURI), []byte(p.CodeChallenge),
		[]byte("S256"), stateHash[:], []byte("read"), []byte(strconv.FormatInt(p.ExpiresAt, 10))), nil
}
