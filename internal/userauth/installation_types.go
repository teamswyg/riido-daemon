package userauth

import "context"

type InstallationKey struct {
	Algorithm      string `json:"algorithm"`
	PublicKey      string `json:"public_key"`
	KeyFingerprint string `json:"key_fingerprint"`
}

// InstallationProof is a typed, read-only authorization intent, never an
// arbitrary signing payload. OS protection belongs to the caller's signer.
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

func (InstallationProof) String() string   { return "[redacted installation authorization proof]" }
func (InstallationProof) GoString() string { return "[redacted installation authorization proof]" }

type InstallationSigner func(context.Context, InstallationProof) (string, error)
