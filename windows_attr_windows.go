//go:build windows

package main

import "syscall"

func windowsHiddenProcAttr() *syscall.SysProcAttr { return &syscall.SysProcAttr{HideWindow: true} }
