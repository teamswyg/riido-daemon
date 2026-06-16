package openclaw

type wireFrameType string

const (
	wireFrameMalformed  wireFrameType = "malformed"
	wireFrameFullResult wireFrameType = "full_result"

	wireFrameNDJSONPrefix = "ndjson:"
)

type wireNDJSONEvent string

const (
	wireNDJSONText    wireNDJSONEvent = "text"
	wireNDJSONLog     wireNDJSONEvent = "log"
	wireNDJSONError   wireNDJSONEvent = "error"
	wireNDJSONSession wireNDJSONEvent = "session"
	wireNDJSONUsage   wireNDJSONEvent = "usage"
)
