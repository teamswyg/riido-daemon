package main

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type cliLinkDaemon struct{ io contextIO }

func (d cliLinkDaemon) Info(ctx context.Context) (standalonecli.LinkDevice, error) {
	data, err := d.call(ctx, contextcli.CLIInstallationLinkRequest{Method: contextcli.IPCMethodCLIInstallationLinkInfo, SchemaVersion: contextcli.CLIInstallationLinkIPCSchema})
	return standalonecli.LinkDevice{Resource: data.Resource, DeviceID: data.DeviceID}, err
}

func (d cliLinkDaemon) Confirm(ctx context.Context, proof userauth.DeviceLinkProof, signature string) error {
	p := contextcli.CLIInstallationLinkProof(proof)
	data, err := d.call(ctx, contextcli.CLIInstallationLinkRequest{Method: contextcli.IPCMethodCLIInstallationLinkConfirm, SchemaVersion: contextcli.CLIInstallationLinkIPCSchema, Proof: &p, Signature: signature})
	if err != nil {
		return err
	}
	if !data.Confirmed || data.Resource != proof.Resource || data.DeviceID != proof.ExpectedDeviceID || data.IntentID != proof.IntentID {
		return &contextcli.CLIInstallationLinkError{Code: "LINK_INVALID_RESPONSE"}
	}
	return nil
}

func (d cliLinkDaemon) call(ctx context.Context, request contextcli.CLIInstallationLinkRequest) (contextcli.CLIInstallationLinkData, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	fail := &contextcli.CLIInstallationLinkError{Code: "LINK_DAEMON_UNAVAILABLE"}
	if err := ctx.Err(); err != nil {
		return contextcli.CLIInstallationLinkData{}, err
	}
	socket, err := d.io.daemonSocket(d.io.environment())
	if err != nil {
		return contextcli.CLIInstallationLinkData{}, fail
	}
	deadline, _ := ctx.Deadline()
	conn, err := d.io.dial(socket, time.Until(deadline))
	if err != nil {
		return contextcli.CLIInstallationLinkData{}, fail
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	if err := conn.SetDeadline(deadline); err != nil {
		return contextcli.CLIInstallationLinkData{}, fail
	}
	if err := json.NewEncoder(conn).Encode(request); err != nil {
		return contextcli.CLIInstallationLinkData{}, fail
	}
	body, err := io.ReadAll(io.LimitReader(conn, (16<<10)+1))
	if err != nil {
		return contextcli.CLIInstallationLinkData{}, fail
	}
	return contextcli.DecodeCLIInstallationLinkReply(body, request.Method)
}
