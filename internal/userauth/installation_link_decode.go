package userauth

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

func validDeviceLinkClient(c *Client) bool {
	if c == nil || c.http == nil || c.resource == nil || c.config.ClientID != "riido-cli" || len(c.config.Scopes) != 1 || c.config.Scopes[0] != "read" {
		return false
	}
	u, err := url.Parse(c.config.Resource)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && u.Path == "" && u.RawPath == "" && u.RawQuery == "" && !u.ForceQuery && u.Fragment == "" && u.Opaque == "" && c.config.Resource == "https://"+u.Host
}

func deviceLinkID(value string) bool {
	return value != "" && len(value) <= 512 && utf8.ValidString(value) && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "/\\\x00\r\n\t")
}

func deviceLinkRandom(value string) bool {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(raw) == 32 && base64.RawURLEncoding.EncodeToString(raw) == value
}

func deviceLinkDigest(value string) bool {
	raw, err := hex.DecodeString(value)
	return err == nil && len(raw) == 32 && hex.EncodeToString(raw) == value
}

func validDeviceLinkRequest(r DeviceLinkRequest) bool {
	return deviceLinkRandom(r.RequestID) && deviceLinkRandom(r.RecoverySecret) && deviceLinkID(r.ExpectedDeviceID)
}

// Exact key parsing rejects duplicates, case aliases, nulls and trailing JSON.
func deviceLinkObject(raw []byte, keys ...string) (map[string]json.RawMessage, error) {
	if len(raw) == 0 || len(raw) > deviceLinkLimit || !utf8.Valid(raw) {
		return nil, linkFailure("LINK_INVALID_RESPONSE")
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	token, err := dec.Token()
	if err != nil || token != json.Delim('{') {
		return nil, linkFailure("LINK_INVALID_RESPONSE")
	}
	fields := map[string]json.RawMessage{}
	for dec.More() {
		token, err = dec.Token()
		key, ok := token.(string)
		if err != nil || !ok || !slices.Contains(keys, key) || fields[key] != nil {
			return nil, linkFailure("LINK_INVALID_RESPONSE")
		}
		var value json.RawMessage
		if dec.Decode(&value) != nil || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, linkFailure("LINK_INVALID_RESPONSE")
		}
		fields[key] = value
	}
	if token, err = dec.Token(); err != nil || token != json.Delim('}') || len(fields) != len(keys) || dec.Decode(new(any)) != io.EOF {
		return nil, linkFailure("LINK_INVALID_RESPONSE")
	}
	return fields, nil
}

func decodeDeviceLink(fields map[string]json.RawMessage, result any) error {
	raw, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, result)
}

func (c *Client) decodeDeviceLinkRecovery(raw []byte, input DeviceLinkRecoveryRequest) (DeviceLinkRecoveryResult, error) {
	// Inspect the state only to choose the exact permitted key set. The final
	// parser still checks duplicate keys in the original response, not a map copy.
	var shape struct {
		State string `json:"state"`
	}
	if json.Unmarshal(raw, &shape) != nil {
		return DeviceLinkRecoveryResult{}, linkFailure("LINK_INVALID_RESPONSE")
	}
	keys := []string{"request_id", "state"}
	if shape.State == "proof_required" {
		keys = append(keys, "proof")
	}
	fields, err := deviceLinkObject(raw, keys...)
	if err != nil {
		return DeviceLinkRecoveryResult{}, err
	}
	var result DeviceLinkRecoveryResult
	if decodeDeviceLink(fields, &result) != nil || result.RequestID != input.RequestID {
		return DeviceLinkRecoveryResult{}, linkFailure("LINK_INVALID_RESPONSE")
	}
	switch result.State {
	case "awaiting_approval", "proof_expired", "linked":
		return result, nil
	case "proof_required":
	default:
		return DeviceLinkRecoveryResult{}, linkFailure("LINK_INVALID_RESPONSE")
	}
	proofFields, err := deviceLinkObject(fields["proof"], "resource", "client_id", "installation_id", "intent_id", "challenge_hash", "key_fingerprint", "expected_device_id", "approved_principal_id", "expires_at")
	if err != nil {
		return DeviceLinkRecoveryResult{}, err
	}
	var proof DeviceLinkProof
	if decodeDeviceLink(proofFields, &proof) != nil || proof.Resource != c.config.Resource || proof.ClientID != c.config.ClientID || proof.KeyFingerprint != input.KeyFingerprint || proof.ExpectedDeviceID != input.ExpectedDeviceID || !deviceLinkDigest(proof.ChallengeHash) || !deviceLinkRandom(proof.IntentID) || !deviceLinkID(proof.InstallationID) || !deviceLinkID(proof.ApprovedPrincipalID) {
		return DeviceLinkRecoveryResult{}, linkFailure("LINK_INVALID_RESPONSE")
	}
	expiry, err := time.Parse(time.RFC3339Nano, proof.ExpiresAt)
	if err != nil || expiry.IsZero() || expiry.UTC().Format(time.RFC3339Nano) != proof.ExpiresAt {
		return DeviceLinkRecoveryResult{}, linkFailure("LINK_INVALID_RESPONSE")
	}
	if !expiry.After(time.Now()) {
		return DeviceLinkRecoveryResult{}, linkFailure("LINK_PROOF_EXPIRED")
	}
	result.Proof = &proof
	return result, nil
}
