package saasplane

import "testing"

func newTestPlaneWithToken(t *testing.T, baseURL string, agents []AgentBinding, token string) *Plane {
	t.Helper()
	plane, err := New(Config{
		BaseURL:     baseURL,
		DaemonID:    "daemon-1",
		DeviceID:    "device-1",
		Agents:      agents,
		BearerToken: token,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return plane
}
