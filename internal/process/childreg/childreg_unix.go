//go:build !windows

package childreg

import "syscall"

// reapProcessGroup SIGKILLs the whole process group led by pid. It returns true
// only if the group still existed (a genuine orphan from a previous daemon).
// A negative pid targets the process group in syscall.Kill.
func reapProcessGroup(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(-pid, syscall.SIGKILL) == nil
}
