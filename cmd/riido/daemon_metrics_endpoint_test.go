package main

import "testing"

type daemonMetricsPayload struct {
	Metrics struct {
		RuntimeCount        int `json:"runtime_count"`
		RuntimeResponding   int `json:"runtime_responding"`
		ProviderAvailable   int `json:"provider_available"`
		ProviderUnavailable int `json:"provider_unavailable"`
		RunningTasks        int `json:"running_tasks"`
	} `json:"metrics"`
}

func TestDaemonMetricsEndpoint(t *testing.T) {
	daemon := startForegroundDaemonForStatus(t)
	defer assertForegroundDaemonExits(t, daemon.cancel, daemon.errCh)

	out := daemonEndpointOutput(t, daemon.socket, daemonCommandMetrics)
	metrics := decodeDaemonEndpointJSON[daemonMetricsPayload](t, out)
	if metrics.Metrics.RuntimeCount != 4 || metrics.Metrics.RuntimeResponding != 4 {
		t.Fatalf("metrics payload mismatch: %+v\n%s", metrics.Metrics, out)
	}
	if metrics.Metrics.ProviderAvailable+metrics.Metrics.ProviderUnavailable != 4 {
		t.Fatalf("provider metric count mismatch: %+v\n%s", metrics.Metrics, out)
	}
}
