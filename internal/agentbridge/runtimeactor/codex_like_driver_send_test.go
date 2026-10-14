package runtimeactor

import (
	"context"
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (d *codexLikeDriver) send(
	ctx context.Context,
	io agentbridge.ProtocolIO,
	method string,
	params map[string]any,
) error {
	d.nextID++
	d.pending[d.nextID] = method
	frame := map[string]any{"jsonrpc": "2.0", "id": d.nextID, "method": method}
	if params != nil {
		frame["params"] = params
	}
	b, _ := json.Marshal(frame)
	return io.WriteStdin(ctx, append(b, '\n'))
}
