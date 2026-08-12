//go:build !windows
package main
type dummyProcAttr struct{}
func windowsHiddenProcAttr() any { return nil }
