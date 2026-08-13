//go:build windows

package main

import (
	"os"
	"syscall"
)

func windowsHiddenProcAttr() *syscall.SysProcAttr { return &syscall.SysProcAttr{HideWindow: true} }

// terminateSelf tue le process tout de suite, sans attendre le teardown COM
// de WebView2 (DestroyWindow + os.Exit peuvent se deadlock et figent Quitter).
func terminateSelf() {
	k32 := syscall.NewLazyDLL("kernel32.dll")
	proc := k32.NewProc("TerminateProcess")
	proc.Call(^uintptr(0), 0) // GetCurrentProcess() == (HANDLE)-1
	os.Exit(0)
}
