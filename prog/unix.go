//go:build unix

package prog

import (
	"fmt"
	"syscall"
	"unsafe"
)

func up(count int) {
	fmt.Printf("\x1b[%dA", count)
}

func down(count int) {
	fmt.Printf("\x1b[%dB", count)
}

func clear() {
	fmt.Print("\x1b[2K\r")
}

func size() (int, int) {
	var win struct {
		row uint16
		col uint16
	}
	syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&win)),
	)
	return int(win.row), int(win.col)
}
