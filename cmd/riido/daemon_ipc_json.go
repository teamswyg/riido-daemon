package main

import (
	"encoding/json"
	"net"
)

func writeDaemonJSON(conn net.Conn, value any) error {
	return json.NewEncoder(conn).Encode(value)
}

func writeUnknownDaemonMethod(conn net.Conn, method daemonMethod) error {
	return writeDaemonJSON(conn, map[string]any{
		"schema_version": DaemonStatusSchemaVersion,
		"error":          "unknown method",
		"method":         string(method),
	})
}
