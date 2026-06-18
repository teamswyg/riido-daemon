package saasplane

import "testing"

func newTestPlane(t *testing.T, baseURL string, agents []AgentBinding) *Plane {
	t.Helper()
	return newTestPlaneWithToken(t, baseURL, agents, "")
}
