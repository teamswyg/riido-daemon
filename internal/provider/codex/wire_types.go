package codex

type rawFrameType string

const (
	rawFrameMalformed rawFrameType = "malformed"
	rawFrameError     rawFrameType = "error"
	rawFrameResponse  rawFrameType = "response"

	rawFrameNotificationPrefix  = "notification:"
	rawFrameServerRequestPrefix = "server_request:"
)

type codexMethod string

const (
	codexMethodError codexMethod = "error"

	codexMethodInitialize   codexMethod = "initialize"
	codexMethodInitialized  codexMethod = "initialized"
	codexMethodThreadStart  codexMethod = "thread/start"
	codexMethodThreadResume codexMethod = "thread/resume"
	codexMethodTurnStart    codexMethod = "turn/start"

	codexMethodThreadStarted       codexMethod = "thread_started"
	codexMethodThreadResumed       codexMethod = "thread_resumed"
	codexMethodThreadStartedSlash  codexMethod = "thread/started"
	codexMethodThreadResumedSlash  codexMethod = "thread/resumed"
	codexMethodThreadStatusChanged codexMethod = "thread/status/changed"
	codexMethodThreadStatusAlt     codexMethod = "thread_status_changed"

	codexMethodTurnStarted       codexMethod = "turn_started"
	codexMethodTurnStartedSlash  codexMethod = "turn/started"
	codexMethodTurnCompleted     codexMethod = "turn_completed"
	codexMethodTurnCompleteSlash codexMethod = "turn/completed"
	codexMethodTurnError         codexMethod = "turn_error"
	codexMethodTurnErrorSlash    codexMethod = "turn/error"
	codexMethodTurnFailedSlash   codexMethod = "turn/failed"

	codexMethodAgentMessage          codexMethod = "agent_message"
	codexMethodItemAgentMessageDelta codexMethod = "item/agentMessage/delta"
	codexMethodReasoning             codexMethod = "reasoning"

	codexMethodCommandStarted   codexMethod = "command_execution_started"
	codexMethodCommandOutput    codexMethod = "command_execution_output"
	codexMethodCommandCompleted codexMethod = "command_execution_completed"
	codexMethodApplyPatchStart  codexMethod = "apply_patch_started"
	codexMethodApplyPatchDone   codexMethod = "apply_patch_completed"

	codexMethodAccountRateLimitsUpdated codexMethod = "account/rateLimits/updated"
	codexMethodAccountRateLimitsAlt     codexMethod = "account_rate_limits_updated"
	codexMethodUsage                    codexMethod = "usage"
	codexMethodThreadTokenUsage         codexMethod = "thread/tokenUsage/updated"

	codexMethodItemStarted             codexMethod = "item/started"
	codexMethodItemUpdated             codexMethod = "item/updated"
	codexMethodItemCompleted           codexMethod = "item/completed"
	codexMethodHookStarted             codexMethod = "hook/started"
	codexMethodHookCompleted           codexMethod = "hook/completed"
	codexMethodMCPStartupStatusUpdated codexMethod = "mcpServer/startupStatus/updated"
	codexMethodRemoteControlChanged    codexMethod = "remoteControl/status/changed"

	codexMethodApproveCommand codexMethod = "approve_command"
	codexMethodApprovePatch   codexMethod = "approve_patch"
)

type threadStatus string

const (
	threadStatusIdle        threadStatus = "idle"
	threadStatusCompleted   threadStatus = "completed"
	threadStatusComplete    threadStatus = "complete"
	threadStatusFinished    threadStatus = "finished"
	threadStatusDone        threadStatus = "done"
	threadStatusReady       threadStatus = "ready"
	threadStatusSucceeded   threadStatus = "succeeded"
	threadStatusError       threadStatus = "error"
	threadStatusErrored     threadStatus = "errored"
	threadStatusFailed      threadStatus = "failed"
	threadStatusAborted     threadStatus = "aborted"
	threadStatusCancelled   threadStatus = "cancelled"
	threadStatusCanceled    threadStatus = "canceled"
	threadStatusInterrupted threadStatus = "interrupted"
	threadStatusRunning     threadStatus = "running"
	threadStatusActive      threadStatus = "active"
	threadStatusInProgress  threadStatus = "in_progress"
	threadStatusWorking     threadStatus = "working"
	threadStatusStreaming   threadStatus = "streaming"
	threadStatusThinking    threadStatus = "thinking"
	threadStatusBusy        threadStatus = "busy"
	threadStatusGenerating  threadStatus = "generating"
	threadStatusTurnRunning threadStatus = "turn_running"
)
