package saasplane

import (
	"net/http"
	"time"
)

type Config struct {
	AssignmentEvents            bool // Device-scoped assignment notifications; legacy servers are supported.
	BaseURL                     string
	DaemonID                    string
	DeviceID                    string
	DeviceSecret                string
	ExecutionCoverageGeneration string
	Profile                     string
	AppVersion                  string
	PID                         int
	StartedAt                   time.Time
	Agents                      []AgentBinding
	BearerToken                 string
	HTTPClient                  *http.Client
	RequestTimeout              time.Duration
	LongPollWait                time.Duration
}
