//go:build darwin || linux || freebsd || netbsd || openbsd

package cliui

import (
	"os"
	"syscall"
	"unsafe"
)

// isTerminal asks the terminal driver for the window size; files, pipes and
// other character devices such as /dev/null answer ENOTTY.
func isTerminal(file *os.File) bool {
	var size struct{ rows, cols, x, y uint16 }
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&size)))
	return errno == 0
}

func terminalWidth(file *os.File) int {
	var size struct{ rows, cols, x, y uint16 }
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&size)))
	if errno != 0 {
		return 0
	}
	return int(size.cols)
}
