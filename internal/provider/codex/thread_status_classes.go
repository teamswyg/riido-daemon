package codex

func codexStatusIsTerminal(status threadStatus) bool {
	switch status {
	case threadStatusIdle, threadStatusCompleted, threadStatusComplete, threadStatusFinished,
		threadStatusDone, threadStatusReady, threadStatusSucceeded:
		return true
	default:
		return false
	}
}

func codexStatusIsError(status threadStatus) bool {
	switch status {
	case threadStatusError, threadStatusErrored, threadStatusFailed, threadStatusAborted,
		threadStatusCancelled, threadStatusCanceled, threadStatusInterrupted:
		return true
	default:
		return false
	}
}

func codexStatusIsActive(status threadStatus) bool {
	switch status {
	case threadStatusRunning, threadStatusActive, threadStatusInProgress, threadStatusWorking,
		threadStatusStreaming, threadStatusThinking, threadStatusBusy, threadStatusGenerating,
		threadStatusTurnRunning:
		return true
	default:
		return false
	}
}
