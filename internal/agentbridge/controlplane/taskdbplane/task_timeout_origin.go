package taskdbplane

import "github.com/teamswyg/riido-contracts/task"

func timeoutCanOriginate(state task.TaskState) bool {
	switch state.Code() {
	case task.TaskStateCodeRunning, task.TaskStateCodeNeedsInput, task.TaskStateCodeBlocked, task.TaskStateCodeValidating, task.TaskStateCodeHumanReview:
		return true
	default:
		return false
	}
}
