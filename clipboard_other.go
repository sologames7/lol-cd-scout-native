//go:build !windows

package main

func setClipboard(string) error { return nil }
