package contextcli

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestCLILinkIPCDecodeIsExactAndRedacted(t *testing.T) {
	valid := `{"method":"cli_installation_link_confirm","schema_version":"riido-cli-device-link-ipc.v1","proof":{"resource":"https://api.test","client_id":"riido-cli","installation_id":"install","intent_id":"intent","challenge_hash":"digest","key_fingerprint":"fingerprint","expected_device_id":"device","approved_principal_id":"owner","expires_at":"time"},"signature":"synthetic-signature"}`
	request, err := DecodeCLIInstallationLinkRequest([]byte(valid))
	if err != nil || request.Proof == nil || request.Proof.ExpectedDeviceID != "device" {
		t.Fatal("valid exact request rejected")
	}
	for name, raw := range map[string]string{
		"duplicate":        strings.Replace(valid, `"signature":`, `"signature":"old","signature":`, 1),
		"case alias":       strings.Replace(valid, `"signature":`, `"Signature":`, 1),
		"nested duplicate": strings.Replace(valid, `"intent_id":`, `"intent_id":"old","intent_id":`, 1),
		"nested unknown":   strings.Replace(valid, `"intent_id":`, `"url":"https://other.test","intent_id":`, 1),
		"null":             strings.Replace(valid, `"signature":"synthetic-signature"`, `"signature":null`, 1),
		"extra object":     valid + `{}`,
		"caller authority": strings.TrimSuffix(valid, "}") + `,"authorization":"synthetic-token"}`,
		"run capability":   strings.TrimSuffix(valid, "}") + `,"run_capability":"synthetic-token"}`,
		"oversize":         strings.Replace(valid, "synthetic-signature", strings.Repeat("x", 16<<10), 1),
		"wrong schema":     strings.Replace(valid, "device-link-ipc.v1", "device-link-ipc.v2", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeCLIInstallationLinkRequest([]byte(raw)); err == nil || err.Error() != "LINK_INVALID_REQUEST" {
				t.Fatal("invalid request was accepted or reflected")
			}
		})
	}
	for _, value := range []any{request, *request.Proof, CLIInstallationLinkReply{Data: &CLIInstallationLinkData{IntentID: "intent"}}} {
		for _, format := range []string{"%v", "%+v", "%#v"} {
			formatted := fmt.Sprintf(format, value)
			if strings.Contains(formatted, "synthetic-signature") || strings.Contains(formatted, "fingerprint") || strings.Contains(formatted, "\"intent\"") {
				t.Fatal("protected IPC data formatting leaked")
			}
		}
	}
	if _, err := json.Marshal(request); err != nil {
		t.Fatal(err)
	}
}
