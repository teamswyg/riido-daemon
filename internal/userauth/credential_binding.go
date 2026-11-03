package userauth

import "strings"

// CredentialBinding is established by one verified proof-backed login. It is
// local custody metadata and never replaces the server's current authorization.
type CredentialBinding struct {
	Resource       string `json:"resource"`
	ClientID       string `json:"client_id"`
	AccountID      string `json:"account_id"`
	KeyFingerprint string `json:"key_fingerprint"`
	AuthOrigin     string `json:"auth_origin"`
}

const (
	AuthOriginBrowser = "browser"
	AuthOriginDesktop = "desktop"
)

func (CredentialBinding) String() string     { return "[redacted credential binding]" }
func (b CredentialBinding) GoString() string { return b.String() }

func (b CredentialBinding) Valid() bool {
	u, err := safeURL(b.Resource)
	return err == nil && b.Resource == u.Scheme+"://"+u.Host &&
		b.ClientID == "riido-cli" && deviceLinkID(b.AccountID) &&
		strings.TrimSpace(b.AccountID) == b.AccountID && deviceLinkDigest(b.KeyFingerprint) &&
		(b.AuthOrigin == AuthOriginBrowser || b.AuthOrigin == AuthOriginDesktop)
}

func SameCredentialBinding(a, b *CredentialBinding) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}

func (m *Manager) acceptsBinding(credential Credential) bool {
	b := credential.Binding
	return b == nil || m.client != nil && b.Valid() && b.Resource == m.client.config.Resource && b.ClientID == m.client.config.ClientID
}
