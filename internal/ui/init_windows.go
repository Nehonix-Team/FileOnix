//go:build windows

package ui

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	// ENABLE_VIRTUAL_TERMINAL_PROCESSING enables ANSI escape sequences on Windows 10+
	ENABLE_VIRTUAL_TERMINAL_PROCESSING uint32 = 0x0004
)

func init() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")

	// Enable Virtual Terminal Processing for Stdout
	out := syscall.Handle(os.Stdout.Fd())
	var outMode uint32
	ret, _, _ := getConsoleMode.Call(uintptr(out), uintptr(unsafe.Pointer(&outMode)))
	if ret != 0 {
		outMode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING
		setConsoleMode.Call(uintptr(out), uintptr(outMode))
	}

	// Enable Virtual Terminal Processing for Stderr
	errOut := syscall.Handle(os.Stderr.Fd())
	var errMode uint32
	ret, _, _ = getConsoleMode.Call(uintptr(errOut), uintptr(unsafe.Pointer(&errMode)))
	if ret != 0 {
		errMode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING
		setConsoleMode.Call(uintptr(errOut), uintptr(errMode))
	}

	// Set console output code page to UTF-8 (65001)
	if setConsoleOutputCP.Find() == nil {
		setConsoleOutputCP.Call(uintptr(65001))
	}
}
