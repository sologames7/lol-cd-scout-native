//go:build windows

package main

import (
	"testing"
	"unsafe"
)

func TestSendKbdSize(t *testing.T) {
	if n := unsafe.Sizeof(sendKbd{}); n != 40 {
		t.Fatalf("sendKbd = %d octets, SendInput x64 attend 40", n)
	}
}
