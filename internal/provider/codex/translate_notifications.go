package codex

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func translateNotification(method codexMethod, p map[string]any) []agentbridge.Event {
	switch method {
	case codexMethodThreadStarted, codexMethodThreadResumed:
		return codexThreadStartedEvent(stringField(p, "thread_id"))

	case codexMethodThreadStartedSlash, codexMethodThreadResumedSlash:
		return codexThreadStartedEvent(threadIDFromParams(p))

	case codexMethodTurnStarted, codexMethodTurnStartedSlash:
		return codexTurnStartedEvent()

	case codexMethodAgentMessage:
		return codexAgentTextDeltaEvent(p)

	case codexMethodItemAgentMessageDelta:
		return codexItemTextDeltaEvent(p)

	case codexMethodReasoning:
		return codexReasoningDeltaEvent(p)

	case codexMethodCommandStarted:
		return codexCommandStartedEvent(p)

	case codexMethodCommandOutput:
		return codexCommandOutputEvent(p)

	case codexMethodCommandCompleted:
		return codexCommandCompletedEvent(p)

	case codexMethodApplyPatchStart:
		return codexApplyPatchStartedEvent(p)

	case codexMethodApplyPatchDone:
		return codexApplyPatchDoneEvent(p)

	case codexMethodTurnCompleted, codexMethodTurnCompleteSlash:
		return codexTurnCompletedEvent(p)

	case codexMethodTurnError, codexMethodTurnErrorSlash, codexMethodTurnFailedSlash:
		return codexTurnFailedEvent(p)

	case codexMethodAccountRateLimitsUpdated, codexMethodAccountRateLimitsAlt:
		return codexRateLimitsUpdatedEvent()

	case codexMethodItemStarted, codexMethodItemUpdated, codexMethodItemCompleted,
		codexMethodHookStarted, codexMethodHookCompleted,
		codexMethodMCPStartupStatusUpdated,
		codexMethodRemoteControlChanged:
		return codexStructuralLifecycleEvent(method)

	case codexMethodUsage:
		return codexUsageDeltaEvent(p)

	case codexMethodThreadTokenUsage:
		return codexThreadTokenUsageEvent(p)
	default:
		return codexUnknownNotificationEvent(method)
	}
}
