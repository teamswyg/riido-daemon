package claude

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (d *protocolDriver) OnStart(ctx context.Context, io agentbridge.ProtocolIO) error {
	if d.written {
		return nil
	}
	body, err := marshalClaudeUserFrame(d.req.Prompt)
	if err != nil {
		return fmt.Errorf("claude driver: marshal user frame: %w", err)
	}
	if err := io.WriteStdin(ctx, body); err != nil {
		return fmt.Errorf("claude driver: write user frame: %w", err)
	}
	if err := io.CloseStdin(ctx); err != nil {
		return fmt.Errorf("claude driver: close stdin: %w", err)
	}
	d.written = true
	return nil
}
