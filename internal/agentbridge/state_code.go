package agentbridge

import contractrunstate "github.com/teamswyg/riido-contracts/runstate"

type (
	RunStateCode   = contractrunstate.RunStateCode
	RunStateString = contractrunstate.RunStateString
)

const (
	RunStateCodeUnknown             = contractrunstate.RunStateCodeUnknown
	RunStateCodePending             = contractrunstate.RunStateCodePending
	RunStateCodePreparing           = contractrunstate.RunStateCodePreparing
	RunStateCodeStartingProvider    = contractrunstate.RunStateCodeStartingProvider
	RunStateCodeHandshaking         = contractrunstate.RunStateCodeHandshaking
	RunStateCodeRunning             = contractrunstate.RunStateCodeRunning
	RunStateCodeWaitingToolApproval = contractrunstate.RunStateCodeWaitingToolApproval
	RunStateCodeToolRunning         = contractrunstate.RunStateCodeToolRunning
	RunStateCodeWaitingProvider     = contractrunstate.RunStateCodeWaitingProvider
	RunStateCodeCompleting          = contractrunstate.RunStateCodeCompleting
	RunStateCodeCompleted           = contractrunstate.RunStateCodeCompleted
	RunStateCodeFailed              = contractrunstate.RunStateCodeFailed
	RunStateCodeCancelled           = contractrunstate.RunStateCodeCancelled
	RunStateCodeTimedOut            = contractrunstate.RunStateCodeTimedOut
	RunStateCodeIdleStopped         = contractrunstate.RunStateCodeIdleStopped
)

const (
	RunStateStringPending             = contractrunstate.RunStateStringPending
	RunStateStringPreparing           = contractrunstate.RunStateStringPreparing
	RunStateStringStartingProvider    = contractrunstate.RunStateStringStartingProvider
	RunStateStringHandshaking         = contractrunstate.RunStateStringHandshaking
	RunStateStringRunning             = contractrunstate.RunStateStringRunning
	RunStateStringWaitingToolApproval = contractrunstate.RunStateStringWaitingToolApproval
	RunStateStringToolRunning         = contractrunstate.RunStateStringToolRunning
	RunStateStringWaitingProvider     = contractrunstate.RunStateStringWaitingProvider
	RunStateStringCompleting          = contractrunstate.RunStateStringCompleting
	RunStateStringCompleted           = contractrunstate.RunStateStringCompleted
	RunStateStringFailed              = contractrunstate.RunStateStringFailed
	RunStateStringCancelled           = contractrunstate.RunStateStringCancelled
	RunStateStringTimedOut            = contractrunstate.RunStateStringTimedOut
	RunStateStringIdleStopped         = contractrunstate.RunStateStringIdleStopped
)

func ParseRunStateCode(value string) RunStateCode {
	return contractrunstate.ParseRunStateCode(value)
}

func AllRunStateCodes() []RunStateCode {
	return contractrunstate.AllRunStateCodes()
}
