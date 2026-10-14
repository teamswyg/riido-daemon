package main

import (
	"encoding/json"
	"net"

	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
)

func readMwsdTestMethod(conn net.Conn) (string, bool) {
	var req struct {
		Method string `json:"method"`
	}
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		return "", false
	}
	return req.Method, true
}

func mwsdTestResponse(snapshot mwsdbridge.Snapshot, method string) (any, bool) {
	switch method {
	case "status":
		return snapshot.Status, true
	case "graph":
		return snapshot.Graph, true
	case "domain":
		return snapshot.Domain, true
	case "harness":
		return snapshot.Harness, true
	case "orchestration":
		return snapshot.Orchestration, true
	case "projects":
		return snapshot.Projects, true
	default:
		return nil, false
	}
}

func writeMwsdTestResponse(conn net.Conn, method string, data any) {
	body, _ := json.Marshal(data)
	_ = json.NewEncoder(conn).Encode(struct {
		OK     bool            `json:"ok"`
		Method string          `json:"method"`
		Data   json.RawMessage `json:"data"`
	}{
		OK:     true,
		Method: method,
		Data:   body,
	})
}
