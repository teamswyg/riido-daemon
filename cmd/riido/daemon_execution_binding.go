package main

import "net"

func writeExecutionBindingUpdate(conn net.Conn, req daemonRequest, settings daemonSettings) {
	if settings.ExecutionBindingStore == nil {
		_ = writeDaemonJSON(conn, map[string]string{"error": "execution_binding_store_unavailable"})
		return
	}
	binding := req.ExecutionBinding
	if req.Method == daemonMethodExecutionBindingUpdate && binding == nil {
		_ = writeDaemonJSON(conn, map[string]string{"error": "execution_binding_required"})
		return
	}
	if req.Method == daemonMethodExecutionBindingRemove {
		binding = nil
	}
	if err := settings.ExecutionBindingStore.Replace(req.WorkspaceID, req.TaskID, binding); err != nil {
		_ = writeDaemonJSON(conn, map[string]string{"error": "execution_binding_update_failed"})
		return
	}
	_ = writeDaemonJSON(conn, map[string]string{"status": "updated", "effective": "subsequent_executions"})
}
