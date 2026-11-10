package saasplane

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const deviceBootstrapSchema = "riido-device-bootstrap.v1"

func bootstrapFields(raw []byte, keys ...string) (map[string]string, error) {
	if len(raw) == 0 || len(raw) > 8192 || !utf8.Valid(raw) {
		return nil, cliLinkFailure("BOOTSTRAP_INVALID")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return nil, cliLinkFailure("BOOTSTRAP_INVALID")
	}
	values := make(map[string]string, len(keys))
	for d.More() {
		t, err = d.Token()
		k, ok := t.(string)
		if err != nil || !ok || !slices.Contains(keys, k) {
			return nil, cliLinkFailure("BOOTSTRAP_INVALID")
		}
		if _, exists := values[k]; exists {
			return nil, cliLinkFailure("BOOTSTRAP_INVALID")
		}
		t, err = d.Token()
		v, ok := t.(string)
		if err != nil || !ok || v == "" {
			return nil, cliLinkFailure("BOOTSTRAP_INVALID")
		}
		values[k] = v
	}
	t, err = d.Token()
	if err != nil || t != json.Delim('}') || len(values) != len(keys) {
		return nil, cliLinkFailure("BOOTSTRAP_INVALID")
	}
	if _, err = d.Token(); err != io.EOF {
		return nil, cliLinkFailure("BOOTSTRAP_INVALID")
	}
	return values, nil
}

func bootstrapRandomID(id string) bool {
	v, e := base64.RawURLEncoding.DecodeString(id)
	return e == nil && len(v) == 32 && base64.RawURLEncoding.EncodeToString(v) == id
}

func bootstrapTimestamp(s string) (time.Time, bool) {
	at, err := time.Parse(time.RFC3339Nano, s)
	return at, err == nil && !at.IsZero() && at.UTC().Format(time.RFC3339Nano) == s
}

type bootstrapRequest struct{ resource, origin, installation, key, device, secret, owner, id, nextHash string }

func (bootstrapRequest) String() string     { return "[redacted native bootstrap]" }
func (r bootstrapRequest) GoString() string { return r.String() }
func parseBootstrapInput(raw []byte) (bootstrapRequest, error) {
	f, e := bootstrapFields(raw, "resource", "control_plane_origin", "installation_id", "key_fingerprint", "device_id", "device_secret", "owner_principal_id", "transition_id", "next_secret_hash")
	if e != nil {
		return bootstrapRequest{}, e
	}
	q := bootstrapRequest{resource: f["resource"], installation: f["installation_id"], key: f["key_fingerprint"], device: f["device_id"], secret: f["device_secret"], owner: f["owner_principal_id"], id: f["transition_id"], nextHash: f["next_secret_hash"]}
	switch q.resource {
	case "https://monolith.riido.io":
		q.origin = "https://ai-api.riido.io"
	case "https://monolith.api.riido.io":
		q.origin = "https://staging.ai-api.riido.io"
	default:
		return bootstrapRequest{}, cliLinkFailure("BOOTSTRAP_INVALID")
	}
	if f["control_plane_origin"] != q.origin || !cliLinkID(q.installation) || !cliLinkID(q.owner) || !cliLinkDigest(q.key) || !cliLinkDigest(q.nextHash) || !bootstrapRandomID(q.id) || len(q.device) != 36 || !strings.HasPrefix(q.device, "dev_") || strings.Trim(q.device[4:], "0123456789abcdef") != "" || !validTerminationHeader(q.secret, 512) {
		return bootstrapRequest{}, cliLinkFailure("BOOTSTRAP_INVALID")
	}
	return q, nil
}
