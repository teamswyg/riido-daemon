package saasplane

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func validCLILinkACK(raw []byte, proof contextcli.CLIInstallationLinkProof) bool {
	expected := map[string]string{
		"schema_version": "riido-cli-installation-link.v1", "installation_id": proof.InstallationID,
		"resource": proof.Resource, "client_id": proof.ClientID, "key_fingerprint": proof.KeyFingerprint,
		"device_id": proof.ExpectedDeviceID, "link_approved_by": proof.ApprovedPrincipalID,
		"device_owner_at_link": proof.ApprovedPrincipalID, "link_revision": "1", "linked_at": "",
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return false
	}
	seen := map[string]bool{}
	for decoder.More() {
		token, err = decoder.Token()
		key, ok := token.(string)
		want, known := expected[key]
		if err != nil || !ok || !known || seen[key] {
			return false
		}
		var value *string
		if decoder.Decode(&value) != nil || value == nil {
			return false
		}
		if key == "linked_at" {
			when, parseErr := time.Parse(time.RFC3339Nano, *value)
			expiry, expiryErr := time.Parse(time.RFC3339Nano, proof.ExpiresAt)
			if parseErr != nil || expiryErr != nil || when.IsZero() || when.UTC().Format(time.RFC3339Nano) != *value || !when.Before(expiry) {
				return false
			}
		} else if *value != want {
			return false
		}
		seen[key] = true
	}
	token, err = decoder.Token()
	return err == nil && token == json.Delim('}') && len(seen) == len(expected) && decoder.Decode(new(any)) == io.EOF
}
