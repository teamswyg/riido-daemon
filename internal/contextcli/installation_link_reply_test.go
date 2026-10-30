package contextcli

import (
	"strings"
	"testing"
)

func TestLinkReplyRequiresExactShape(t *testing.T) {
	valid := `{"schema_version":"riido-cli-device-link-ipc.v1","ok":true,"data":{"resource":"https://api.example.test","device_id":"device"}}`
	if data, err := DecodeCLIInstallationLinkReply([]byte(valid), IPCMethodCLIInstallationLinkInfo); err != nil || data.DeviceID != "device" {
		t.Fatal("valid info rejected", err)
	}
	for _, raw := range []string{
		strings.Replace(valid, `"ok":true`, `"ok":true,"ok":true`, 1),
		strings.Replace(valid, `"data":{`, `"data":{"extra":"value",`, 1),
		strings.Replace(valid, `"device_id":"device"`, `"device_id":null`, 1),
		strings.Replace(valid, `"device_id"`, `"Device_ID"`, 1),
		valid + `{}`, strings.Repeat(" ", 16<<10) + valid,
		`{"schema_version":"old","ok":true,"data":{}}`,
	} {
		if _, err := DecodeCLIInstallationLinkReply([]byte(raw), IPCMethodCLIInstallationLinkInfo); err == nil {
			t.Fatal("ambiguous daemon reply accepted")
		}
	}
	if _, err := DecodeCLIInstallationLinkReply([]byte(valid), IPCMethodCLIInstallationLinkConfirm); err == nil {
		t.Fatal("info substituted for confirmation")
	}
	failure := `{"schema_version":"riido-cli-device-link-ipc.v1","ok":false,"error":{"code":"synthetic-untrusted-text"}}`
	if _, err := DecodeCLIInstallationLinkReply([]byte(failure), IPCMethodCLIInstallationLinkInfo); err == nil || strings.Contains(err.Error(), "synthetic-untrusted-text") {
		t.Fatal("daemon error text leaked")
	}
}
