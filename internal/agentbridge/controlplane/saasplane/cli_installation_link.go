package saasplane

import (
	"context"
	"encoding/base64"
	"math"
	"net/http"
	"strconv"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// CLIInstallationLinkConfig is daemon-owned. No field is accepted from IPC.
type CLIInstallationLinkConfig struct {
	BaseURL, Resource, DeviceID, DeviceSecret, ExecutionCoverageGeneration string
	HTTPClient                                                             *http.Client
}

func (CLIInstallationLinkConfig) String() string   { return "[redacted daemon link configuration]" }
func (CLIInstallationLinkConfig) GoString() string { return "[redacted daemon link configuration]" }

type CLIInstallationLinkClient struct{ cfg CLIInstallationLinkConfig }

func NewCLIInstallationLinkClient(cfg CLIInstallationLinkConfig) (*CLIInstallationLinkClient, error) {
	if !cliLinkOrigin(cfg.BaseURL) || !cliLinkOrigin(cfg.Resource) || !validTerminationHeader(cfg.DeviceID, 512) || !validTerminationHeader(cfg.DeviceSecret, 8192) || ValidateExecutionCoverage(cfg.DeviceID, cfg.DeviceSecret, cfg.ExecutionCoverageGeneration) != nil {
		return nil, cliLinkFailure("LINK_UNSUPPORTED")
	}
	return &CLIInstallationLinkClient{cfg: cfg}, nil
}

func (c *CLIInstallationLinkClient) Info() contextcli.CLIInstallationLinkData {
	return contextcli.CLIInstallationLinkData{Resource: c.cfg.Resource, DeviceID: c.cfg.DeviceID}
}

func (c *CLIInstallationLinkClient) Confirm(ctx context.Context, proof contextcli.CLIInstallationLinkProof, signature string) (contextcli.CLIInstallationLinkData, error) {
	if !c.validProof(proof) || !cliLinkSignature(signature) {
		return contextcli.CLIInstallationLinkData{}, cliLinkFailure("LINK_INVALID_REQUEST")
	}
	return c.confirmRevision(ctx, proof, signature, "1")
}

func (c *CLIInstallationLinkClient) ConfirmTransfer(ctx context.Context, proof contextcli.CLIInstallationLinkProof, previousDevice, previousRevision, signature string) (contextcli.CLIInstallationLinkData, error) {
	rev, err := strconv.ParseUint(previousRevision, 10, 64)
	id, idErr := base64.RawURLEncoding.DecodeString(proof.IntentID)
	if !c.validProof(proof) || !cliLinkSignature(signature) || !cliLinkID(previousDevice) || previousDevice == proof.ExpectedDeviceID || rev == 0 || rev == math.MaxUint64 || err != nil || strconv.FormatUint(rev, 10) != previousRevision || idErr != nil || len(id) != 32 || base64.RawURLEncoding.EncodeToString(id) != proof.IntentID {
		return contextcli.CLIInstallationLinkData{}, cliLinkFailure("LINK_INVALID_REQUEST")
	}
	return c.confirmRevision(ctx, proof, signature, strconv.FormatUint(rev+1, 10))
}

func (c *CLIInstallationLinkClient) confirmRevision(ctx context.Context, proof contextcli.CLIInstallationLinkProof, signature, revision string) (contextcli.CLIInstallationLinkData, error) {
	if err := c.postConfirmation(ctx, proof, signature, revision); err != nil {
		return contextcli.CLIInstallationLinkData{}, err
	}
	return contextcli.CLIInstallationLinkData{Resource: c.cfg.Resource, DeviceID: c.cfg.DeviceID, IntentID: proof.IntentID, Confirmed: true}, nil
}

func cliLinkFailure(code string) error { return &contextcli.CLIInstallationLinkError{Code: code} }
