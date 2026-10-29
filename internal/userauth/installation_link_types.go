package userauth

// These DTOs are caller inputs/results for fixed endpoints, not arbitrary HTTP
// requests. The caller must persist the original request/secret before create.
type DeviceLinkRequest struct{ RequestID, RecoverySecret, ExpectedDeviceID string }
type DeviceLinkRecoveryRequest struct{ RequestID, RecoverySecret, ExpectedDeviceID, KeyFingerprint string }
type DeviceLinkApproval struct {
	RequestID string `json:"request_id"`
	State     string `json:"state"`
	ExpiresAt int64  `json:"expires_at"`
}
type DeviceLinkProof struct {
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
type DeviceLinkRecoveryResult struct {
	RequestID string           `json:"request_id"`
	State     string           `json:"state"`
	Proof     *DeviceLinkProof `json:"proof,omitempty"`
}

func (DeviceLinkRequest) String() string           { return "[redacted device link request]" }
func (DeviceLinkRequest) GoString() string         { return "[redacted device link request]" }
func (DeviceLinkRecoveryRequest) String() string   { return "[redacted device link recovery]" }
func (DeviceLinkRecoveryRequest) GoString() string { return "[redacted device link recovery]" }
func (DeviceLinkApproval) String() string          { return "[redacted device link approval]" }
func (DeviceLinkApproval) GoString() string        { return "[redacted device link approval]" }
func (DeviceLinkProof) String() string             { return "[redacted device link proof]" }
func (DeviceLinkProof) GoString() string           { return "[redacted device link proof]" }
func (DeviceLinkRecoveryResult) String() string    { return "[redacted device link result]" }
func (DeviceLinkRecoveryResult) GoString() string  { return "[redacted device link result]" }

// DeviceLinkError never implements OAuth invalid_grant/refresh-replay semantics.
// It contains no provider body, URL, request ID or credential material.
type DeviceLinkError struct {
	Code       string
	StatusCode int
	Retryable  bool
}

func (*DeviceLinkError) Error() string { return "CLI device link operation failed" }
func linkFailure(code string) error    { return &DeviceLinkError{Code: code} }
