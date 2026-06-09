//go:build windows

package childreg

// reapProcessGroup is a no-op on Windows: provider children are not placed in
// process groups (Setpgid is Unix-only), so there is no group to kill yet. A
// reliable Windows reaper is tracked separately (D7).
func reapProcessGroup(int) bool { return false }
