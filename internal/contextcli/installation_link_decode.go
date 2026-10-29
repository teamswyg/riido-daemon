package contextcli

import (
	"bytes"
	"encoding/json"
	"io"
	"slices"
)

func DecodeCLIInstallationLinkRequest(raw []byte) (CLIInstallationLinkRequest, error) {
	var request CLIInstallationLinkRequest
	fail := &CLIInstallationLinkError{Code: "LINK_INVALID_REQUEST"}
	if len(raw) == 0 || len(raw) > 16<<10 || json.Unmarshal(raw, &request) != nil || request.SchemaVersion != CLIInstallationLinkIPCSchema {
		return CLIInstallationLinkRequest{}, fail
	}
	switch request.Method {
	case IPCMethodCLIInstallationLinkInfo:
		if _, err := cliLinkIPCObject(raw, "method", "schema_version"); err != nil {
			return CLIInstallationLinkRequest{}, fail
		}
	case IPCMethodCLIInstallationLinkConfirm:
		fields, err := cliLinkIPCObject(raw, "method", "schema_version", "proof", "signature")
		if err != nil || request.Proof == nil || request.Signature == "" {
			return CLIInstallationLinkRequest{}, fail
		}
		if _, err = cliLinkIPCObject(fields["proof"], "resource", "client_id", "installation_id", "intent_id", "challenge_hash", "key_fingerprint", "expected_device_id", "approved_principal_id", "expires_at"); err != nil {
			return CLIInstallationLinkRequest{}, fail
		}
	default:
		return CLIInstallationLinkRequest{}, fail
	}
	return request, nil
}

func cliLinkIPCObject(raw []byte, allowed ...string) (map[string]json.RawMessage, error) {
	fail := &CLIInstallationLinkError{Code: "LINK_INVALID_REQUEST"}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, fail
	}
	fields := map[string]json.RawMessage{}
	for decoder.More() {
		token, err = decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || !slices.Contains(allowed, key) || fields[key] != nil {
			return nil, fail
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, fail
		}
		fields[key] = value
	}
	if token, err = decoder.Token(); err != nil || token != json.Delim('}') || len(fields) != len(allowed) || decoder.Decode(new(any)) != io.EOF {
		return nil, fail
	}
	return fields, nil
}
