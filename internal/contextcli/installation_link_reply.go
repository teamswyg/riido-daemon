package contextcli

import "encoding/json"

// DecodeCLIInstallationLinkReply rejects ambiguous and older daemon responses.
func DecodeCLIInstallationLinkReply(raw []byte, method string) (CLIInstallationLinkData, error) {
	fail := &CLIInstallationLinkError{Code: "LINK_INVALID_RESPONSE"}
	var reply CLIInstallationLinkReply
	if len(raw) == 0 || len(raw) > 16<<10 || json.Unmarshal(raw, &reply) != nil || reply.SchemaVersion != CLIInstallationLinkIPCSchema {
		return CLIInstallationLinkData{}, fail
	}
	if !reply.OK {
		fields, err := cliLinkIPCObject(raw, "schema_version", "ok", "error")
		if err != nil || reply.Error == nil || reply.Error.Code == "" {
			return CLIInstallationLinkData{}, fail
		}
		if _, err := cliLinkIPCObject(fields["error"], "code"); err != nil {
			return CLIInstallationLinkData{}, fail
		}
		// Never propagate arbitrary local process text to CLI diagnostics.
		return CLIInstallationLinkData{}, &CLIInstallationLinkError{Code: "LINK_DAEMON_REJECTED"}
	}
	fields, err := cliLinkIPCObject(raw, "schema_version", "ok", "data")
	if err != nil || reply.Data == nil || reply.Data.Resource == "" || reply.Data.DeviceID == "" {
		return CLIInstallationLinkData{}, fail
	}
	keys := []string{"resource", "device_id"}
	if method == IPCMethodCLIInstallationLinkConfirm {
		keys = append(keys, "intent_id", "confirmed")
		if reply.Data.IntentID == "" || !reply.Data.Confirmed {
			return CLIInstallationLinkData{}, fail
		}
	} else if method != IPCMethodCLIInstallationLinkInfo {
		return CLIInstallationLinkData{}, fail
	}
	if _, err := cliLinkIPCObject(fields["data"], keys...); err != nil {
		return CLIInstallationLinkData{}, fail
	}
	return *reply.Data, nil
}
