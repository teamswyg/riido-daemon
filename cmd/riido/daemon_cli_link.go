package main

import (
	"context"
	"errors"
	"net"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func writeCLIInstallationLink(ctx context.Context, conn net.Conn, request *contextcli.CLIInstallationLinkRequest, settings daemonSettings) {
	reply := contextcli.CLIInstallationLinkReply{SchemaVersion: contextcli.CLIInstallationLinkIPCSchema}
	client, err := saasplane.NewCLIInstallationLinkClient(saasplane.CLIInstallationLinkConfig{
		BaseURL: settings.SaaSURL, Resource: settings.CLIAPIResource, DeviceID: settings.DeviceID,
		DeviceSecret: settings.DeviceSecret, ExecutionCoverageGeneration: settings.ExecutionCoverageGeneration,
		HTTPClient: settings.cliLinkHTTPClient,
	})
	var data contextcli.CLIInstallationLinkData
	if err == nil {
		switch {
		case request == nil:
			err = &contextcli.CLIInstallationLinkError{Code: "LINK_INVALID_REQUEST"}
		case request.Method == contextcli.IPCMethodCLIInstallationLinkInfo:
			data = client.Info()
		case request.Method == contextcli.IPCMethodCLIInstallationLinkConfirm && request.Proof != nil:
			data, err = client.Confirm(ctx, *request.Proof, request.Signature)
		default:
			err = &contextcli.CLIInstallationLinkError{Code: "LINK_INVALID_REQUEST"}
		}
	}
	if err != nil {
		if !errors.As(err, &reply.Error) {
			reply.Error = &contextcli.CLIInstallationLinkError{Code: "LINK_UNAVAILABLE"}
		}
	} else {
		reply.OK, reply.Data = true, &data
	}
	_ = writeDaemonJSON(conn, reply)
}
