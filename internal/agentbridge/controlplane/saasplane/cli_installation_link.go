package saasplane

import (
	"context"
	"net/http"

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
	if err := c.postConfirmation(ctx, proof, signature); err != nil {
		return contextcli.CLIInstallationLinkData{}, err
	}
	return contextcli.CLIInstallationLinkData{Resource: c.cfg.Resource, DeviceID: c.cfg.DeviceID, IntentID: proof.IntentID, Confirmed: true}, nil
}

func cliLinkFailure(code string) error { return &contextcli.CLIInstallationLinkError{Code: code} }
