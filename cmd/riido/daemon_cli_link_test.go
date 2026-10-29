package main

import (
	"encoding/json"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/logging"
)

func cliLinkIPCFixture(t *testing.T, settings daemonSettings, raw string) []byte {
	t.Helper()
	client, server := net.Pipe()
	defer client.Close()
	_ = client.SetDeadline(time.Now().Add(5 * time.Second))
	go handleDaemonConn(server, startFlags{}, settings, time.Now(), nil, nil, nil, nil, logging.NewWriterLogger(io.Discard))
	if _, err := io.WriteString(client, raw+"\n"); err != nil {
		t.Fatal(err)
	}
	response, err := io.ReadAll(client)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func TestCLILinkIPCDoesNotSubstituteDaemonID(t *testing.T) {
	settings := daemonSettings{CLIAPIResource: "https://resource.test", SaaSURL: "https://cp.test", DaemonID: "display-daemon", DeviceSecret: "synthetic-secret"}
	response := cliLinkIPCFixture(t, settings, `{"method":"cli_installation_link_info","schema_version":"riido-cli-device-link-ipc.v1"}`)
	var reply contextcli.CLIInstallationLinkReply
	if json.Unmarshal(response, &reply) != nil || reply.OK || reply.Error == nil || reply.Error.Code != "LINK_UNSUPPORTED" || strings.Contains(string(response), settings.DaemonID) {
		t.Fatal("linking substituted the display daemon ID")
	}
}

func TestCLILinkIPCRejectsTrailingRequestPreservingLegacyDecode(t *testing.T) {
	settings := loadDaemonSettingsForTest(t, saasDeviceCredentialEnv())
	response := cliLinkIPCFixture(t, settings, `{"method":"cli_installation_link_info","schema_version":"riido-cli-device-link-ipc.v1"}{}`)
	var reply contextcli.CLIInstallationLinkReply
	if json.Unmarshal(response, &reply) != nil || reply.Error == nil || reply.Error.Code != "LINK_INVALID_REQUEST" {
		t.Fatal("trailing new-method request accepted")
	}
	legacy := cliLinkIPCFixture(t, settings, `{"method":"health","future_extension":"preserved"}`)
	if !strings.Contains(string(legacy), `"health":"ok"`) {
		t.Fatal("new strict decoding changed legacy health compatibility")
	}
}

func TestCLILinkIPCInfoUsesExplicitResourceAndCurrentDevice(t *testing.T) {
	env := saasDeviceCredentialEnv()
	env["RIIDO_CLI_API_URL"] = "https://monolith.example.test"
	settings := loadDaemonSettingsForTest(t, env)
	body := cliLinkIPCFixture(t, settings, `{"method":"cli_installation_link_info","schema_version":"riido-cli-device-link-ipc.v1"}`)
	var reply struct {
		SchemaVersion string `json:"schema_version"`
		OK            bool   `json:"ok"`
		Data          struct {
			Resource string `json:"resource"`
			DeviceID string `json:"device_id"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &reply) != nil || reply.SchemaVersion != "riido-cli-device-link-ipc.v1" || !reply.OK || reply.Data.Resource != env["RIIDO_CLI_API_URL"] || reply.Data.DeviceID != settings.DeviceID {
		t.Fatal("explicit device link information is unavailable or incorrectly bound")
	}
	if strings.Contains(string(body), settings.DeviceSecret) || strings.Contains(string(body), settings.SaaSURL) {
		t.Fatal("info exposed daemon credential or transport settings")
	}
}

func TestCLILinkIPCDoesNotUseDisplayURLOrDefaultResource(t *testing.T) {
	env := saasDeviceCredentialEnv()
	env[envServerURL] = "https://display-only.example.test"
	settings := loadDaemonSettingsForTest(t, env)
	body := cliLinkIPCFixture(t, settings, `{"method":"cli_installation_link_info","schema_version":"riido-cli-device-link-ipc.v1"}`)
	var reply struct {
		OK    bool `json:"ok"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &reply) != nil || reply.OK || reply.Error.Code != "LINK_UNSUPPORTED" {
		t.Fatal("missing explicit resource did not close only the new linking surface")
	}
}
