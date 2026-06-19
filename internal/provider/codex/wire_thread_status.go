package codex

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
