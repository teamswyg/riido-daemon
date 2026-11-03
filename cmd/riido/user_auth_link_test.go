package main

import (
	"context"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestAuthLinkRejectsAgentBeforeOpeningUserStorage(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("agent opened user credentials")
		return nil, nil
	}
	if got := exitCodeOf(t, runAuth(h.io, []string{"link", "--json"})); got != 4 {
		t.Fatalf("exit=%d", got)
	}
	if strings.Contains(h.stdout.String(), testRunCapability) || len(h.daemon.requests) != 0 {
		t.Fatal("agent link leaked or contacted daemon")
	}
}

func TestAuthLinkHelpAndInvalidArgsHaveNoSideEffects(t *testing.T) {
	for _, args := range [][]string{{"link", "--help"}, {"link", "--with-token", "synthetic-secret", "--json"}} {
		h := newContextHarness(t, nil)
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
			t.Fatal("help/invalid opened user credentials")
			return nil, nil
		}
		_ = runAuth(h.io, args)
		if len(h.daemon.requests) != 0 {
			t.Fatal("help/invalid contacted daemon")
		}
	}
}

func TestAuthLinkDaemonIPCUsesOnlyTypedProof(t *testing.T) {
	h := newContextHarness(t, nil)
	h.daemon.replies[contextcli.IPCMethodCLIInstallationLinkInfo] = `{"schema_version":"riido-cli-device-link-ipc.v1","ok":true,"data":{"resource":"https://api.example.test","device_id":"device"}}`
	h.daemon.replies[contextcli.IPCMethodCLIInstallationLinkConfirm] = `{"schema_version":"riido-cli-device-link-ipc.v1","ok":true,"data":{"resource":"https://api.example.test","device_id":"device","intent_id":"intent","confirmed":true}}`
	daemon := cliLinkDaemon{io: h.io}
	device, err := daemon.Info(t.Context())
	if err != nil || device.DeviceID != "device" {
		t.Fatal("info failed", err)
	}
	proof := userauth.DeviceLinkProof{Resource: device.Resource, ExpectedDeviceID: device.DeviceID, IntentID: "intent"}
	if err := daemon.Confirm(t.Context(), proof, "synthetic-signature"); err != nil {
		t.Fatal("confirm failed", err)
	}
	request := h.daemon.lastRequest(t)
	if len(request) != 4 || request["method"] != contextcli.IPCMethodCLIInstallationLinkConfirm || request["run_capability"] != nil || request["recovery_secret"] != nil {
		t.Fatal("unexpected IPC authority")
	}
	proof.ExpectedDeviceID = "other-device"
	if err := daemon.Confirm(t.Context(), proof, "synthetic-signature"); err == nil {
		t.Fatal("mismatched ACK accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := daemon.Info(ctx); err == nil {
		t.Fatal("cancelled IPC succeeded")
	}
}
