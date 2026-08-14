//go:build !windows

package main

func setClipboard(string) error { return nil }

func writeClipboard(string, int) error { return nil }
