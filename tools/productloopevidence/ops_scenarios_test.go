package main

func operationalScenarioIDs() []string {
	return []string{
		"ops.monitoring.client_surface_anomaly",
		"ops.ui.copy_regression",
		"ops.usability.agent_name_change",
		"ops.resilience.internet_disconnect_wait",
		"ops.stress.concurrent_users",
		"ops.stress.single_pc_agent_capacity",
		"ops.stress.boot_packet_burst",
		"ops.chaos.control_plane_restart_recovery",
		"ops.chaos.scale_out_rebalance",
		"ops.chaos.scale_out_duration",
		"ops.chaos.full_outage_daemon_backoff",
		"ops.chaos.recovery_packet_surge",
		"ops.scenario.exception_equals_weekend_open",
		"desktop.ui.hello_to_say_world_body_only",
		"release.backend_daemon_ready",
	}
}
