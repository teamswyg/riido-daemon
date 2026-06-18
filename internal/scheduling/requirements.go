package scheduling

import "github.com/teamswyg/riido-contracts/provider/capability"

// TaskRequirements is the C5 view of a task before claim/execute.
type TaskRequirements struct {
	Provider                 capability.ProviderKind
	RequiredSurfaces         []RequiredSurface
	AllowExperimentalRuntime bool
}

// RuntimeCapability is the C5 subset of C3 ProviderCapability needed for
// scheduler eligibility. It intentionally excludes process/runtime details.
type RuntimeCapability struct {
	RuntimeID                 capability.RuntimeID
	Provider                  capability.ProviderKind
	CapabilityFingerprint     capability.CapabilityFingerprint
	SlotLimit                 int
	SlotsInUse                int
	Available                 bool
	CompatibilityStatus       capability.CompatibilityStatus
	RequiresExperimentalOptIn bool
	SupportsStreaming         bool
	SupportsResume            bool
	SupportsSystem            bool
	SupportsMaxTurns          bool
	SupportsMCP               bool
	SupportsToolHooks         bool
	SupportsUsage             bool
	SupportsWorktree          bool
}

func (c RuntimeCapability) compatibilityBlocked() bool {
	return c.CompatibilityStatus == capability.CompatBlocked
}

func (c RuntimeCapability) slotsExhausted() bool {
	return c.SlotLimit > 0 && c.SlotsInUse >= c.SlotLimit
}
