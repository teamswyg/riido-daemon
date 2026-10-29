package contextcli

const (
	CLIInstallationLinkIPCSchema        = "riido-cli-device-link-ipc.v1"
	IPCMethodCLIInstallationLinkInfo    = "cli_installation_link_info"
	IPCMethodCLIInstallationLinkConfirm = "cli_installation_link_confirm"
)

// CLIInstallationLinkProof mirrors the fixed approved server proof. Possession
// of these fields alone grants no authority; CP verifies its stored intent/key.
type CLIInstallationLinkProof struct {
	Resource            string `json:"resource"`
	ClientID            string `json:"client_id"`
	InstallationID      string `json:"installation_id"`
	IntentID            string `json:"intent_id"`
	ChallengeHash       string `json:"challenge_hash"`
	KeyFingerprint      string `json:"key_fingerprint"`
	ExpectedDeviceID    string `json:"expected_device_id"`
	ApprovedPrincipalID string `json:"approved_principal_id"`
	ExpiresAt           string `json:"expires_at"`
}

func (CLIInstallationLinkProof) String() string   { return "[redacted installation link proof]" }
func (CLIInstallationLinkProof) GoString() string { return "[redacted installation link proof]" }

type CLIInstallationLinkRequest struct {
	Method        string                    `json:"method"`
	SchemaVersion string                    `json:"schema_version"`
	Proof         *CLIInstallationLinkProof `json:"proof,omitempty"`
	Signature     string                    `json:"signature,omitempty"`
}

func (CLIInstallationLinkRequest) String() string   { return "[redacted installation link request]" }
func (CLIInstallationLinkRequest) GoString() string { return "[redacted installation link request]" }

type CLIInstallationLinkData struct {
	Resource  string `json:"resource"`
	DeviceID  string `json:"device_id"`
	IntentID  string `json:"intent_id,omitempty"`
	Confirmed bool   `json:"confirmed,omitempty"`
}

type CLIInstallationLinkError struct {
	Code string `json:"code"`
}

func (e *CLIInstallationLinkError) Error() string { return e.Code }

type CLIInstallationLinkReply struct {
	SchemaVersion string                    `json:"schema_version"`
	OK            bool                      `json:"ok"`
	Data          *CLIInstallationLinkData  `json:"data,omitempty"`
	Error         *CLIInstallationLinkError `json:"error,omitempty"`
}

func (CLIInstallationLinkReply) String() string   { return "[redacted installation link reply]" }
func (CLIInstallationLinkReply) GoString() string { return "[redacted installation link reply]" }
