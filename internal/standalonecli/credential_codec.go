package standalonecli

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"slices"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

const (
	protectedSessionVersion = "riido-cli-protected-session.v3"
	protectedSessionV2      = "riido-cli-protected-session.v2"
	protectedSessionV1      = "riido-cli-protected-session.v1"
	protectedSessionLimit   = 32 << 10
)

type protectedSession struct {
	SchemaVersion string              `json:"schema_version"`
	Credential    userauth.Credential `json:"credential"`
	PendingLink   *PendingLink        `json:"pending_link"`
}

func (protectedSession) String() string   { return "[redacted protected CLI session]" }
func (protectedSession) GoString() string { return "[redacted protected CLI session]" }

func decodeCredential(raw []byte) (userauth.Credential, error) {
	record, err := decodeProtectedSession(raw)
	return record.Credential, err
}

func encodeCredential(credential userauth.Credential) ([]byte, error) {
	return encodeProtectedSession(protectedSession{SchemaVersion: protectedSessionVersion, Credential: credential})
}

func decodeProtectedSession(raw []byte) (protectedSession, error) {
	var record protectedSession
	if len(raw) == 0 || len(raw) > protectedSessionLimit {
		return record, securecredential.ErrUnavailable
	}
	fields, err := protectedObject(raw)
	if err != nil {
		return record, err
	}
	if _, versioned := fields["schema_version"]; !versioned {
		if !protectedFields(fields, "session_id", "access_token", "refresh_token", "expires_at", "token_type", "scope") {
			return protectedSession{}, securecredential.ErrUnavailable
		}
		// Read the original credential representation without rewriting it. The
		// next successful save is the only migration point.
		if json.Unmarshal(raw, &record.Credential) != nil {
			return protectedSession{}, securecredential.ErrUnavailable
		}
		record.SchemaVersion = protectedSessionVersion
	} else {
		var version string
		if json.Unmarshal(fields["schema_version"], &version) != nil || (version != protectedSessionV1 && version != protectedSessionV2 && version != protectedSessionVersion) {
			return protectedSession{}, securecredential.ErrUnavailable
		}
		credentialFields, credentialErr := protectedObject(fields["credential"])
		credentialKeys := []string{"session_id", "access_token", "refresh_token", "expires_at", "token_type", "scope"}
		if version == protectedSessionVersion {
			credentialKeys = append(credentialKeys, "refresh_recovery_secret")
			if value, present := credentialFields["refresh_recovery_secret"]; present {
				var secret string
				if json.Unmarshal(value, &secret) != nil || secret == "" {
					return protectedSession{}, securecredential.ErrUnavailable
				}
			}
		}
		if !protectedFields(fields, "schema_version", "credential", "pending_link") || len(fields) != 3 || credentialErr != nil || !protectedFields(credentialFields, credentialKeys...) {
			return protectedSession{}, securecredential.ErrUnavailable
		}
		if !bytes.Equal(bytes.TrimSpace(fields["pending_link"]), []byte("null")) {
			pendingFields, pendingErr := protectedObject(fields["pending_link"])
			allowed := []string{"resource", "client_id", "session_id", "key_fingerprint", "expected_device_id", "request_id", "recovery_secret"}
			if version != protectedSessionV1 {
				allowed = append(allowed, "phase")
			}
			if pendingErr != nil || !protectedFields(pendingFields, allowed...) {
				return protectedSession{}, securecredential.ErrUnavailable
			}
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&record) != nil || decoder.Decode(new(any)) != io.EOF {
			return protectedSession{}, securecredential.ErrUnavailable
		}
		if version == protectedSessionV1 && record.PendingLink != nil {
			record.PendingLink.Phase = LinkPrepared
		}
		record.SchemaVersion = protectedSessionVersion
	}
	if !validProtectedSession(record) {
		return protectedSession{}, securecredential.ErrUnavailable
	}
	return record, nil
}

// The protected format has only these three object levels. Check each object's
// exact keys before decoding so duplicate/case aliases cannot erase an unknown
// version or substitute a credential/session field through last-value wins.
func protectedObject(raw []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, securecredential.ErrUnavailable
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err = decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || fields[key] != nil {
			return nil, securecredential.ErrUnavailable
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return nil, securecredential.ErrUnavailable
		}
		fields[key] = value
	}
	if token, err = decoder.Token(); err != nil || token != json.Delim('}') || decoder.Decode(new(any)) != io.EOF {
		return nil, securecredential.ErrUnavailable
	}
	return fields, nil
}

func protectedFields(fields map[string]json.RawMessage, allowed ...string) bool {
	for key := range fields {
		if !slices.Contains(allowed, key) {
			return false
		}
	}
	return true
}

func encodeProtectedSession(record protectedSession) ([]byte, error) {
	if !validProtectedSession(record) {
		return nil, securecredential.ErrUnavailable
	}
	raw, err := json.Marshal(record)
	if err != nil || len(raw) > protectedSessionLimit {
		return nil, securecredential.ErrUnavailable
	}
	return raw, nil
}

func validProtectedSession(record protectedSession) bool {
	if record.SchemaVersion != protectedSessionVersion || record.Credential.AccessToken == "" || record.Credential.TokenType != "Bearer" {
		return false
	}
	if secret := record.Credential.RefreshRecoverySecret; secret != "" {
		raw, err := base64.RawURLEncoding.DecodeString(secret)
		if err != nil || len(raw) != 32 || base64.RawURLEncoding.EncodeToString(raw) != secret || record.Credential.SessionID == "" || record.Credential.RefreshToken == "" {
			return false
		}
	}
	return record.PendingLink == nil || (record.PendingLink.valid() && record.PendingLink.SessionID == record.Credential.SessionID)
}
