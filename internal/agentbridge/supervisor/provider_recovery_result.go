package supervisor

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func terminalRecoveryContext(a *Actor, task *runningTask, res agentbridge.Result) agentbridge.Result {
	if task == nil || task.recovery == nil {
		return res
	}
	c := task.recovery.Sanitized()
	if c.OutcomeCertainty == "started" {
		c.OutcomeCertainty = "unknown"
		c.State = "user_action_required"
		c.Reason = "unknown_outcome"
		switch res.Status {
		case agentbridge.ResultCompleted:
			c.State = "dispatched"
			if c.Attempt > 1 || c.SwitchCount > 0 {
				c.State = "recovered"
			}
			c.Reason = "none"
			c.OutcomeCertainty = "confirmed"
			c.Circuit = "closed"
		case agentbridge.ResultCancelled:
			c.State = "cancelled"
			c.Reason = "cancelled"
		default:
			// Every other terminal result remains an uncertain dispatched outcome.
		}
		if a.recovery != nil {
			if res.Status == agentbridge.ResultCancelled {
				a.recovery.Release(task.recoveryScope, task.taskID)
				if c.Circuit == "half_open" {
					c.Circuit = "open"
				}
			} else {
				a.recovery.Complete(task.recoveryScope, task.taskID, res.Status == agentbridge.ResultCompleted, time.Now())
				if res.Status != agentbridge.ResultCompleted {
					c.Circuit = "open"
				}
			}
		}
	}
	if task.cancelCause != nil || task.ctx != nil && task.ctx.Err() != nil {
		c.State = "cancelled"
		c.Reason = "cancelled"
	}
	res.RecoveryContext = &c
	return res
}
