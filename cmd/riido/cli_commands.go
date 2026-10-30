package main

type mainCommand string

const (
	mainCommandMwsd                mainCommand = "mwsd"
	mainCommandTask                mainCommand = "task"
	mainCommandServe               mainCommand = "serve"
	mainCommandAPI                 mainCommand = "api"
	mainCommandBridge              mainCommand = "bridge"
	mainCommandClaudePermissionMCP mainCommand = "claude-permission-mcp"
	mainCommandDaemon              mainCommand = "daemon"
	mainCommandVersion             mainCommand = "version"
	mainCommandWork                mainCommand = "work"
	mainCommandResource            mainCommand = "resource"
	mainCommandRepo                mainCommand = "repo"
	mainCommandContext             mainCommand = "context"
	mainCommandAgents              mainCommand = "agents"
	mainCommandMeeting             mainCommand = "meeting"
	mainCommandAuth                mainCommand = "auth"
	mainCommandWhoami              mainCommand = "whoami"
	mainCommandDoctor              mainCommand = "doctor"
	mainCommandCapabilities        mainCommand = "capabilities"
)

type daemonCommand string

const (
	daemonCommandStart   daemonCommand = "start"
	daemonCommandStatus  daemonCommand = "status"
	daemonCommandHealth  daemonCommand = "health"
	daemonCommandReady   daemonCommand = "ready"
	daemonCommandMetrics daemonCommand = "metrics"
	daemonCommandStop    daemonCommand = "stop"
	daemonCommandLogs    daemonCommand = "logs"
)

type daemonMethod string

const (
	daemonMethodDefault                    daemonMethod = ""
	daemonMethodExecutionBindingUpdate     daemonMethod = "execution_binding_update"
	daemonMethodExecutionBindingRemove     daemonMethod = "execution_binding_remove"
	daemonMethodStatus                     daemonMethod = "status"
	daemonMethodHealth                     daemonMethod = "health"
	daemonMethodReady                      daemonMethod = "ready"
	daemonMethodMetrics                    daemonMethod = "metrics"
	daemonMethodShutdown                   daemonMethod = "shutdown"
	daemonMethodToolApproval               daemonMethod = "tool_approval"
	daemonMethodContextRead                daemonMethod = "context_read"
	daemonMethodCLIInstallationLinkInfo    daemonMethod = "cli_installation_link_info"
	daemonMethodCLIInstallationLinkConfirm daemonMethod = "cli_installation_link_confirm"
)

type taskCommand string

const (
	taskCommandList       taskCommand = "list"
	taskCommandTransition taskCommand = "transition"
	taskCommandEvidence   taskCommand = "evidence"
	taskCommandValidate   taskCommand = "validate"
)

type apiCommand string

const (
	apiCommandStatus     apiCommand = "status"
	apiCommandTasks      apiCommand = "tasks"
	apiCommandReviewDemo apiCommand = "review-demo"
	apiCommandTransition apiCommand = "transition"
	apiCommandEvidence   apiCommand = "evidence"
	apiCommandValidate   apiCommand = "validate"
)

type bridgeCommand string

const (
	bridgeCommandProviders bridgeCommand = "providers"
	bridgeCommandDetect    bridgeCommand = "detect"
)

type mwsdCommand string

const (
	mwsdCommandSnapshot      mwsdCommand = "snapshot"
	mwsdCommandProjection    mwsdCommand = "projection"
	mwsdCommandSync          mwsdCommand = "sync"
	mwsdCommandOrchestration mwsdCommand = "orchestration"
	mwsdCommandProjects      mwsdCommand = "projects"
	mwsdCommandStatus        mwsdCommand = "status"
)
