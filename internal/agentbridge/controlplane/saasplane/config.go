package saasplane

import (
	"net/http"
	"time"
)

type Config struct {
	BaseURL        string
	DaemonID       string
	DeviceID       string
	DeviceSecret   string
	Profile        string
	AppVersion     string
	PID            int
	StartedAt      time.Time
	Agents         []AgentBinding
	BearerToken    string
	HTTPClient     *http.Client
	RequestTimeout time.Duration
	LongPollWait   time.Duration
}
