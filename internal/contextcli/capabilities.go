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
	noteUserMode     = "user mode: 서명된 OS 인증 보호 도우미와 OAuth 로그인이 필요합니다; 앱 로그인과 daemon은 사용하지 않습니다"
	noteUnsupported  = "아직 구현되지 않은 명령입니다. 실행하면 원격 조회 없이 exit 8을 반환합니다"
)

// ReadCommandUnavailable guards the agent broker; user-only reads dispatch separately.
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
		if name == "work activity" {
			return CommandCapability{Name: name, Modes: []Mode{ModeUser}, Status: CommandAvailable, Note: noteUserMode + "; 작업 ID와 --team-id, --page/--limit으로 live 활동을 조회합니다; 서버 활동 조회 권한 배포가 필요하며 agent mode는 exit 8입니다"}
		}
		if name == "whoami" || name == "work get" || name == "work tree" || name == "work comments" || name == "work relations" || name == "context get" || name == "meeting get" {
			note := noteDaemonBroker + "; " + noteUserMode
			if name == "work get" {
				note += "; user mode는 작업 ID와 --team-id만 지원합니다"
			}
			if name == "work comments" {
				note += "; user mode는 작업 ID와 --team-id, --page/--limit을 사용하고 --kind 선택 시 source version으로 이어 읽습니다; agent mode의 구형 스냅샷은 고정 20개 --cursor, 새 맥락은 parts의 context part 명령으로 이어 읽습니다"
			}
			if name == "context get" {
				note += "; user mode는 --work ID와 --team-id로 직접 연결된 맥락을 조회하며 서버 read-scope 계약 배포가 필요합니다"
			}
			if name == "meeting get" {
				note += "; user mode는 미팅 ID와 --team-id, --cursor/--limit으로 보드를 이어 읽으며 서버 read-scope 계약 배포가 필요합니다"
			}
			if name == "work tree" {
				note += "; user mode는 작업 ID와 --team-id로 프로젝트·목표·상위 작업과 하위 최대 10개를 한 단계 조회합니다; 서버 read-scope 계약 배포가 필요합니다"
			}
			if name == "work relations" {
				note += "; user mode는 작업 ID와 --team-id로 직접 관계를 실시간 조회하며 페이지 옵션은 없습니다; agent mode는 기존 snapshot을 사용합니다; 서버 관계 조회 권한 배포가 필요합니다"
			}
			return CommandCapability{Name: name, Modes: []Mode{ModeAgent, ModeUser}, Status: CommandAvailable, Note: note}
		}
		if ReadCommandUnavailable(name) {
			return CommandCapability{Name: name, Modes: []Mode{ModeAgent}, Status: CommandUnavailable, Note: noteUnsupported}
		}
		return CommandCapability{Name: name, Modes: []Mode{ModeAgent}, Status: CommandNeedsDaemon, Note: noteDaemonBroker}
	}
	local := func(name string) CommandCapability {
		return CommandCapability{Name: name, Modes: []Mode{ModeAgent, ModeUser}, Status: CommandAvailable}
	}
	commands := []CommandCapability{
		{Name: "repo list", Modes: []Mode{ModeUser}, Status: CommandAvailable, Note: noteUserMode + "; --work ID와 --team-id로 저장된 저장소·PR 메타데이터를 조회합니다; 로컬 Git/GitHub/SSH 인증은 사용·확인하지 않습니다; 큰 메타데이터 자동 이어 읽기 한도 16 MiB, 초과 시 work body 사용; agent mode는 exit 8입니다"},
		{Name: "resource list", Modes: []Mode{ModeUser}, Status: CommandAvailable, Note: noteUserMode + "; --work ID와 --team-id로 링크·첨부·Figma 참조를 조회합니다; 큰 메타데이터 자동 이어 읽기 한도 16 MiB, 초과 시 work body 사용; agent mode는 exit 8입니다"},
		local("capabilities"),
		local("doctor"),
		{Name: "auth status", Modes: []Mode{ModeAgent, ModeUser}, Status: CommandAvailable, Note: "user mode는 서명된 보호 도우미의 저장 여부만 확인합니다; API 인증 확인은 whoami를 사용하세요"},
		{Name: "auth logout", Modes: []Mode{ModeUser}, Status: CommandAvailable, Note: noteUserMode},
		{Name: "auth login", Modes: []Mode{ModeUser}, Status: CommandAvailable, Note: noteUserMode + "; 기본/--browser PKCE만 지원합니다; --desktop/--with-token은 exit 8입니다"},
		agentRead("whoami"),
		agentRead("context get"),
		agentRead("context part"),
		agentRead("context changes"),
	}
	for _, sub := range []string{"get", "tree", "relations", "comments", "activity", "search"} {
		commands = append(commands, agentRead("work "+sub))
	}
	commands = append(commands, CommandCapability{Name: "work replies", Modes: []Mode{ModeUser}, Status: CommandAvailable, Note: noteUserMode + "; 작업 ID, --team-id, --comment-id가 필요합니다; 기존 component 조회에는 페이지가 없고 --kind 선택 시 --page/--limit과 원 --version을 사용합니다; 서버 댓글 조회 계약 배포가 필요합니다"})
	commands = append(commands, CommandCapability{Name: "work body", Modes: []Mode{ModeUser}, Status: CommandAvailable, Note: noteUserMode + "; 작업 ID, --team-id, 원 version·byte offset으로 본문과 큰 typed 자료를 읽습니다; contentPage.selector는 work body의 --comment-id/--page/--comments-limit 또는 meeting body의 --board-cursor/--board-limit으로 보존하세요; 서버 read-scope 계약 배포가 필요합니다"})
	commands = append(commands, agentRead("agents list"), agentRead("agents show"))
	commands = append(commands, agentRead("meeting list"), agentRead("meeting get"))
	commands = append(commands, CommandCapability{Name: "meeting body", Modes: []Mode{ModeUser}, Status: CommandAvailable, Note: noteUserMode + "; 미팅 ID, --team-id, 원 version·byte offset으로 본문과 큰 typed 자료를 읽습니다; contentPage.selector는 work body의 --comment-id/--page/--comments-limit 또는 meeting body의 --board-cursor/--board-limit으로 보존하세요; 서버 read-scope 계약 배포가 필요합니다"})
	return CapabilitiesData{
		CLIVersion:            cliVersion,
		EnvelopeSchemaVersion: EnvelopeSchemaVersion,
		ContextIPCMethod:      IPCMethodContextRead,
		ContextIPCSchema:      ContextIPCSchemaVersion,
		Commands:              commands,
	}
}
