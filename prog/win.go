//go:build windows

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
		size struct {
			col uint16
			row uint16
		}
	}
	dll := syscall.NewLazyDLL("kernel32.dll")
	buf := dll.NewProc("GetConsoleScreenBufferInfo")
	mode := dll.NewProc("SetConsoleMode")
	hdl, _ := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)

	mode.Call(uintptr(hdl), uintptr(7))
	buf.Call(uintptr(hdl), uintptr(unsafe.Pointer(&win)), 0)
	return int(win.size.row), int(win.size.col)
}
