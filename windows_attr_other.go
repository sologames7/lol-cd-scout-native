//go:build !windows

package main

import "os"

type dummyProcAttr struct{}

func windowsHiddenProcAttr() any { return nil }

func terminateSelf() { os.Exit(0) }
