package codex

type rawFrameType string

const (
	rawFrameMalformed rawFrameType = "malformed"
	rawFrameError     rawFrameType = "error"
	rawFrameResponse  rawFrameType = "response"

	rawFrameNotificationPrefix  = "notification:"
	rawFrameServerRequestPrefix = "server_request:"
)
