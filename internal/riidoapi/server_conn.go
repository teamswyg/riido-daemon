package riidoapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
)

func (s Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	var req requestEnvelope
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		_ = writeResponse(conn, responseEnvelope{OK: false, Error: fmt.Sprintf("decode request: %v", err)})
		return
	}
	response := s.handleRequest(ctx, req)
	_ = writeResponse(conn, response)
}
