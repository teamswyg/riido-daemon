package contextcli

// CommandStatus says whether a command works in this binary today.
type CommandStatus string

const (
	CommandAvailable   CommandStatus = "available"
	CommandNeedsDaemon CommandStatus = "needs_daemon_support"
	CommandUnavailable CommandStatus = "unavailable"
)

// CommandCapability describes one command and the modes it serves.
type CommandCapability struct {
	Name   string        `json:"name"`
	Modes  []Mode        `json:"modes"`
	Status CommandStatus `json:"status"`
	Note   string        `json:"note,omitempty"`
}

// CapabilitiesData is the data payload of `riido capabilities`.
type CapabilitiesData struct {
	CLIVersion            string              `json:"cli_version"`
	EnvelopeSchemaVersion string              `json:"envelope_schema_version"`
	ContextIPCMethod      string              `json:"context_ipc_method"`
	ContextIPCSchema      string              `json:"context_ipc_schema_version"`
	Commands              []CommandCapability `json:"commands"`
}

const (
	noteDaemonBroker = "agent mode: daemon의 context_read 브로커(RIIDO_CONTEXT_CLI=on)가 필요하고, Riido 데이터 조회는 control-plane 실행 grant 이후 가능합니다"
	noteUserMode     = "user mode: 서버 인증 계약(ADR) 확정 전까지 사용할 수 없습니다"
	noteUnsupported  = "아직 구현되지 않은 명령입니다. 실행하면 원격 조회 없이 exit 8을 반환합니다"
)

// ReadCommandUnavailable keeps command execution and availability advertising aligned.
func ReadCommandUnavailable(name string) bool {
	switch name {
	case "context changes", "work activity", "work search":
		return true
	default:
		return false
	}
}

// Capabilities returns the static capability table for cliVersion.
func Capabilities(cliVersion string) CapabilitiesData {
	agentRead := func(name string) CommandCapability {
		if ReadCommandUnavailable(name) {
			return CommandCapability{Name: name, Modes: []Mode{ModeAgent}, Status: CommandUnavailable, Note: noteUnsupported}
		}
		return CommandCapability{Name: name, Modes: []Mode{ModeAgent}, Status: CommandNeedsDaemon, Note: noteDaemonBroker}
	}
	local := func(name string) CommandCapability {
		return CommandCapability{Name: name, Modes: []Mode{ModeAgent, ModeUser}, Status: CommandAvailable}
	}
	commands := []CommandCapability{
		local("capabilities"),
		local("doctor"),
		local("auth status"),
		local("auth logout"),
		{Name: "auth login", Modes: []Mode{ModeUser}, Status: CommandUnavailable, Note: noteUserMode},
		agentRead("whoami"),
		agentRead("context get"),
		agentRead("context changes"),
	}
	for _, sub := range []string{"get", "tree", "relations", "comments", "activity", "search"} {
		commands = append(commands, agentRead("work "+sub))
	}
	commands = append(commands, agentRead("agents list"), agentRead("agents show"))
	commands = append(commands, agentRead("meeting list"), agentRead("meeting get"))
	return CapabilitiesData{
		CLIVersion:            cliVersion,
		EnvelopeSchemaVersion: EnvelopeSchemaVersion,
		ContextIPCMethod:      IPCMethodContextRead,
		ContextIPCSchema:      ContextIPCSchemaVersion,
		Commands:              commands,
	}
}
