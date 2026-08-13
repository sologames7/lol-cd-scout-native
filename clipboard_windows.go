//go:build windows

package main

// Presse-papier Windows en syscall pur : le HUD n'a souvent pas le focus
// (WS_EX_NOACTIVATE + raccourcis Alt+1…5), donc navigator.clipboard échoue.
// League ignore le presse-papier OS (clipboard interne) : Ctrl+V depuis
// l'extérieur ne colle rien. Après la copie on frappe le texte en Unicode
// dans le client (SendInput) — le chat doit être ouvert (Entrée), on n'envoie
// pas le message tout seul.

import (
	"errors"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

const (
	cfUnicodeText      = 13
	gmemMoveable       = 0x0002
	inputKeyboard      = 1
	keyeventfKeyup     = 0x0002
	keyeventfUnicode   = 0x0004
	leagueChatMaxRunes = 250
)

var (
	clipUser32                   = syscall.NewLazyDLL("user32.dll")
	clipKernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procOpenClipboard            = clipUser32.NewProc("OpenClipboard")
	procCloseClipboard           = clipUser32.NewProc("CloseClipboard")
	procEmptyClipboard           = clipUser32.NewProc("EmptyClipboard")
	procSetClipboardData         = clipUser32.NewProc("SetClipboardData")
	procSendInput                = clipUser32.NewProc("SendInput")
	procSetForegroundWindow      = clipUser32.NewProc("SetForegroundWindow")
	procAllowSetForegroundWindow = clipUser32.NewProc("AllowSetForegroundWindow")
	procGlobalAlloc              = clipKernel32.NewProc("GlobalAlloc")
	procGlobalLock               = clipKernel32.NewProc("GlobalLock")
	procGlobalUnlock             = clipKernel32.NewProc("GlobalUnlock")
	procGlobalFree               = clipKernel32.NewProc("GlobalFree")
)

// sendKbd = INPUT clavier 64-bit (40 octets). Go aligne Extra sur 8.
type sendKbd struct {
	Type  uint32
	_     uint32
	Vk    uint16
	Scan  uint16
	Flags uint32
	Time  uint32
	Extra uintptr
	_     [8]byte
}

func setClipboard(s string) error {
	utf16, err := syscall.UTF16FromString(s)
	if err != nil {
		return err
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var opened uintptr
	for i := 0; i < 12; i++ {
		opened, _, _ = procOpenClipboard.Call(0)
		if opened != 0 {
			break
		}
		time.Sleep(8 * time.Millisecond)
	}
	if opened == 0 {
		return errors.New("presse-papier occupé")
	}
	defer procCloseClipboard.Call()
	procEmptyClipboard.Call()
	if s == "" {
		return nil
	}

	nbytes := uintptr(len(utf16) * 2)
	h, _, _ := procGlobalAlloc.Call(gmemMoveable, nbytes)
	if h == 0 {
		return errors.New("GlobalAlloc")
	}
	ptr, _, _ := procGlobalLock.Call(h)
	if ptr == 0 {
		procGlobalFree.Call(h)
		return errors.New("GlobalLock")
	}
	dst := unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), len(utf16))
	copy(dst, utf16)
	procGlobalUnlock.Call(h)

	ok, _, _ := procSetClipboardData.Call(cfUnicodeText, h)
	if ok == 0 {
		procGlobalFree.Call(h)
		return errors.New("SetClipboardData")
	}
	if s != "" {
		go typeIntoLeagueChat(s)
	}
	return nil
}

func leagueGameHWND() syscall.Handle {
	if h := findWindow("League of Legends (TM) Client"); h != 0 {
		return h
	}
	return findWindow("League of Legends (TM)")
}

func typeIntoLeagueChat(s string) {
	if s == "" {
		return
	}
	runes := []rune(s)
	if len(runes) > leagueChatMaxRunes {
		s = string(runes[:leagueChatMaxRunes])
	}
	for i := 0; i < 30; i++ {
		if !keyDown(vkMenu) && !keyDown(vkShift) && !keyDown(vkControl) {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	hwnd := leagueGameHWND()
	if hwnd == 0 {
		return
	}
	procAllowSetForegroundWindow.Call(^uintptr(0))
	procSetForegroundWindow.Call(uintptr(hwnd))
	time.Sleep(25 * time.Millisecond)
	utf16, err := syscall.UTF16FromString(s)
	if err != nil || len(utf16) < 2 {
		return
	}
	for _, ch := range utf16[:len(utf16)-1] {
		sendUnicodeKey(ch, 0)
		sendUnicodeKey(ch, keyeventfKeyup)
	}
}

func sendUnicodeKey(ch uint16, up uint32) {
	in := sendKbd{
		Type:  inputKeyboard,
		Scan:  ch,
		Flags: keyeventfUnicode | up,
	}
	procSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
}
