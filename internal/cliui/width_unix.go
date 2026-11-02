//go:build darwin || linux || freebsd || netbsd || openbsd

package cliui

import (
	"os"
	"syscall"
	"unsafe"
)

func terminalWidth(file *os.File) int {
	var size struct{ rows, cols, x, y uint16 }
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&size)))
	if errno != 0 {
		return 0
	}
	return int(size.cols)
}
