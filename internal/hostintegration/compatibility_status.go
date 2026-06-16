package hostintegration

import "github.com/teamswyg/riido-contracts/provider/capability"

func validCompatibilityStatus(status capability.CompatibilityStatus) bool {
	switch status {
	case capability.CompatSupported,
		capability.CompatDegraded,
		capability.CompatExperimental,
		capability.CompatBlocked:
		return true
	default:
		return false
	}
}
