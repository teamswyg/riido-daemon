package securecredential

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"time"
)

// EnsureInstallationKey uses the helper's OS-user/resource/client identity,
// independently of the currently signed-in account and credential item.
func (s *Store) EnsureInstallationKey(ctx context.Context, resource string) (InstallationPublicKey, error) {
	var key InstallationPublicKey
	if !installationResource(resource) {
		return key, ErrUnavailable
	}
	payload := struct {
		Resource string `json:"resource"`
		ClientID string `json:"client_id"`
	}{Resource: resource, ClientID: "riido-cli"}
	data, err := s.installationCall(ctx, resource, "installation-key", payload)
	if err != nil {
		return key, err
	}
	if _, err = installationObject(data, "algorithm", "public_key", "key_fingerprint"); err != nil || json.Unmarshal(data, &key) != nil {
		return InstallationPublicKey{}, ErrUnavailable
	}
	if _, err = key.publicKey(); err != nil {
		return InstallationPublicKey{}, err
	}
	return key, nil
}

// SignInstallationProof accepts only a server-bound read-only authorization
// intent and verifies the helper's DER signature before returning it to OAuth.
func (s *Store) SignInstallationProof(ctx context.Context, key InstallationPublicKey, proof InstallationProof) (string, error) {
	message, err := proof.message(key, time.Now())
	if err != nil {
		return "", err
	}
	data, err := s.installationCall(ctx, proof.Resource, "installation-proof", proof)
	if err != nil {
		return "", err
	}
	if _, err = installationObject(data, "algorithm", "signature"); err != nil {
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
	publicKey, err := key.publicKey()
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(message)
	if !ecdsa.VerifyASN1(publicKey, digest[:], raw) || !time.Unix(proof.ExpiresAt, 0).After(time.Now()) {
		return "", ErrUnavailable
	}
	return result.Signature, nil
}

func (s *Store) installationCall(ctx context.Context, resource, operation string, payload any) (json.RawMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !s.Interactive {
		return nil, ErrInteractionRequired
	}
	ctx, cancel := context.WithTimeout(ctx, helperTimeout)
	defer cancel()
	value, err := json.Marshal(payload)
	if err != nil || len(value) > 8192 {
		return nil, ErrUnavailable
	}
	bound := *s
	bound.Key = installationStorageKey(resource)
	output, err := bound.invoke(ctx, operation, value)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, err
	}
	root, err := installationObject(output, "version", "ok", "data", "error")
	if err != nil || len(root) != 3 {
		return nil, ErrUnavailable
	}
	var version int
	var ok bool
	if json.Unmarshal(root["version"], &version) != nil || version != 1 || json.Unmarshal(root["ok"], &ok) != nil {
		return nil, ErrUnavailable
	}
	if ok {
		if root["data"] == nil || root["error"] != nil {
			return nil, ErrUnavailable
		}
		return root["data"], nil
	}
	if root["data"] != nil {
		return nil, ErrUnavailable
	}
	failure, err := installationObject(root["error"], "code", "message")
	if err != nil {
		return nil, ErrUnavailable
	}
	var code string
	if json.Unmarshal(failure["code"], &code) != nil {
		return nil, ErrUnavailable
	}
	switch code {
	case "INTERACTION_REQUIRED":
		return nil, ErrInteractionRequired
	case "CANCELLED":
		return nil, ErrCancelled
	case "NOT_FOUND":
		return nil, ErrNotFound
	case "TIMED_OUT":
		return nil, context.DeadlineExceeded
	default:
		return nil, ErrUnavailable
	}
}

func installationObject(raw []byte, fields ...string) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, ErrUnavailable
	}
	result := map[string]json.RawMessage{}
	for decoder.More() {
		token, err = decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || !slices.Contains(fields, key) || result[key] != nil {
			return nil, ErrUnavailable
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil || bytes.Equal(value, []byte("null")) {
			return nil, ErrUnavailable
		}
		result[key] = value
	}
	if token, err = decoder.Token(); err != nil || token != json.Delim('}') {
		return nil, ErrUnavailable
	}
	var extra any
	if !errors.Is(decoder.Decode(&extra), io.EOF) {
		return nil, ErrUnavailable
	}
	return result, nil
}
