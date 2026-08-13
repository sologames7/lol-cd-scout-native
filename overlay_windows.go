//go:build windows

package main

// HUD en jeu : Chromium --app plein écran, sans cadre. Color-key #000001.
// WS_EX_TRANSPARENT hors des widgets opaques : curseur + clic droit League.
// Clic droit jamais menu Chrome (pass-through + contextmenu bloqué + --disable-dev-tools).
// Curseur Windows masqué (cursor:none) quand la souris est sur un panneau.

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const hudWindowTitle = "CD Scout HUD"

var (
	user32                         = syscall.NewLazyDLL("user32.dll")
	dwmapi                         = syscall.NewLazyDLL("dwmapi.dll")
	procEnumWindows                = user32.NewProc("EnumWindows")
	procGetWindowTextW             = user32.NewProc("GetWindowTextW")
	procIsWindow                   = user32.NewProc("IsWindow")
	procSetWindowPos               = user32.NewProc("SetWindowPos")
	procGetWindowLongPtrW          = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW          = user32.NewProc("SetWindowLongPtrW")
	procGetAsyncKeyState           = user32.NewProc("GetAsyncKeyState")
	procVkKeyScanW                 = user32.NewProc("VkKeyScanW")
	procGetSystemMetrics           = user32.NewProc("GetSystemMetrics")
	procShowWindow                 = user32.NewProc("ShowWindow")
	procIsWindowVisible            = user32.NewProc("IsWindowVisible")
	procSendMessageW               = user32.NewProc("SendMessageW")
	procPostMessageW               = user32.NewProc("PostMessageW")
	procReleaseCapture             = user32.NewProc("ReleaseCapture")
	procDwmSetWindowAttribute      = dwmapi.NewProc("DwmSetWindowAttribute")
	procGetCursorPos               = user32.NewProc("GetCursorPos")
	procScreenToClient             = user32.NewProc("ScreenToClient")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	procSetClassLongPtrW           = user32.NewProc("SetClassLongPtrW")
	procSetCursor                  = user32.NewProc("SetCursor")
)

const (
	hwndTopmost     = ^uintptr(0)
	hwndNoTopmost   = ^uintptr(1)
	swpNoSize       = 0x0001
	swpNoMove       = 0x0002
	swpNoActivate   = 0x0010
	swpFrameChanged = 0x0020

	wsCaption       = 0x00C00000
	wsThickFrame    = 0x00040000
	wsSysMenu       = 0x00080000
	wsMinimizeBox   = 0x00020000
	wsMaximizeBox   = 0x00010000
	wsBorder        = 0x00800000
	wsDlgFrame      = 0x00400000
	wsPopup         = 0x80000000
	wsClipChildren  = 0x02000000
	wsClipSiblings  = 0x04000000
	wsExNoActivate  = 0x08000000
	wsExToolWindow  = 0x00000080
	wsExAppWindow   = 0x00040000
	wsExLayered     = 0x00080000
	wsExTransparent = 0x00000020
	wsExNoRedirBmp  = 0x00200000 // WS_EX_NOREDIRECTIONBITMAP : incompatible avec layered + color-key
	lwaColorKey     = 0x00000001
	// COLORREF #000001 — fond HUD, jamais dans les cartes. DirectComposition off requis.
	hudChromaKey = 0x00010000

	swHide          = 0
	swShowNA        = 8
	wmClose         = 0x0010
	wmNCLButtonDown = 0x00A1
	htCaption       = 2

	dwmwaNcRenderingPolicy      = 2
	dwmNcRpDisabled             = 1
	dwmwaTransitionsForcedOff   = 3
	dwmwaWindowCornerPreference = 33
	dwmwcpDoNotRound            = 1
	dwmwaBorderColor            = 34
	dwmwaCaptionColor           = 35
	dwmColorNone                = 0xFFFFFFFE

	smCXScreen       = 0
	smCYScreen       = 1
	vkShift          = 0x10
	vkControl        = 0x11
	vkMenu           = 0x12
	vkRButton        = 0x02
	vkMButton        = 0x04
	vkOEM3           = 0xC0 // ² sur AZERTY (Backquote US, au-dessus de Tab)
	gclpHCursor      = int32(-12)
	inputEventBuffer = 32
)

var (
	gwlExStyle = int32(-20)
	gwlStyle   = int32(-16)
)

func hudSupported() bool { return true }

func idx(v int32) uintptr { return uintptr(int64(v)) }

func getLong(hwnd syscall.Handle, i int32) uintptr {
	r, _, _ := procGetWindowLongPtrW.Call(uintptr(hwnd), idx(i))
	return r
}

func setLong(hwnd syscall.Handle, i int32, v uintptr) {
	procSetWindowLongPtrW.Call(uintptr(hwnd), idx(i), v)
}

func dwmAttr(hwnd syscall.Handle, attr uint32, value uint32) {
	procDwmSetWindowAttribute.Call(uintptr(hwnd), uintptr(attr), uintptr(unsafe.Pointer(&value)), 4)
}

// ---------- Recherche de la fenêtre du HUD ----------

var (
	enumMu     sync.Mutex
	enumNeedle string
	enumFound  syscall.Handle
	enumProcCB = syscall.NewCallback(enumWindowProc)
)

func enumWindowProc(hwnd syscall.Handle, _ uintptr) uintptr {
	buf := make([]uint16, 320)
	n, _, _ := procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n > 0 && strings.Contains(syscall.UTF16ToString(buf[:n]), enumNeedle) {
		enumFound = hwnd
		return 0
	}
	return 1
}

func findWindow(titlePart string) syscall.Handle {
	enumMu.Lock()
	defer enumMu.Unlock()
	enumNeedle, enumFound = titlePart, 0
	procEnumWindows.Call(enumProcCB, 0)
	return enumFound
}

func windowAlive(hwnd syscall.Handle) bool {
	if hwnd == 0 {
		return false
	}
	r, _, _ := procIsWindow.Call(uintptr(hwnd))
	return r != 0
}

// ---------- Épinglage / chrome / visibilité ----------

var hudPin = struct {
	mu         sync.Mutex
	pinned     bool
	noActivate bool
	hold       bool
	found      bool
	loop       bool
	placed     bool
	shown      bool
	solid      bool      // souris capturée (drag)
	hideAfter  time.Time // hide ² différé : évite le voile gris au spam
	hwnd       syscall.Handle
}{}

type hudHit struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

var hudHits = struct {
	mu sync.Mutex
	rs []hudHit
}{}

func hudSetHits(rs []hudHit) {
	hudHits.mu.Lock()
	hudHits.rs = rs
	hudHits.mu.Unlock()
}

func hudSetSolid(on bool) {
	hudPin.mu.Lock()
	hudPin.solid = on
	hudPin.mu.Unlock()
	hudSyncPassThrough()
}

func hudStatus() (pinned, noActivate, found bool) {
	hudPin.mu.Lock()
	defer hudPin.mu.Unlock()
	return hudPin.pinned, hudPin.noActivate, hudPin.found
}

func hudHold() bool {
	hudPin.mu.Lock()
	defer hudPin.mu.Unlock()
	return hudPin.hold
}

func hudSetHold(on bool) {
	hudPin.mu.Lock()
	hudPin.hold = on
	hudPin.mu.Unlock()
	hudSyncVisibility()
}

func hudSetPin(pinned, noActivate bool) {
	hudPin.mu.Lock()
	hudPin.pinned, hudPin.noActivate = pinned, noActivate
	start := !hudPin.loop
	hudPin.loop = true
	hudPin.mu.Unlock()
	if start {
		go pinLoop()
	}
	applyPin()
}

func hudHWND() syscall.Handle {
	hudPin.mu.Lock()
	h := hudPin.hwnd
	hudPin.mu.Unlock()
	if windowAlive(h) {
		return h
	}
	h = findWindow(hudWindowTitle)
	hudPin.mu.Lock()
	if h != hudPin.hwnd {
		hudPin.placed = false
	}
	hudPin.hwnd, hudPin.found = h, h != 0
	hudPin.mu.Unlock()
	return h
}

// dressHUD retire le chrome Windows (Chrome le remet souvent : à rappeler en boucle).
func dressHUD(hwnd syscall.Handle, place bool) {
	style := getLong(hwnd, gwlStyle)
	style &^= wsCaption | wsThickFrame | wsSysMenu | wsMinimizeBox | wsMaximizeBox | wsBorder | wsDlgFrame
	// Ne pas forcer WS_VISIBLE : SW_HIDE l'enlève, le remettre ici
	// ressuscite une fenêtre Chromium non peinte (voile gris plein écran).
	style |= wsPopup | wsClipChildren | wsClipSiblings
	setLong(hwnd, gwlStyle, style)

	ex := getLong(hwnd, gwlExStyle)
	ex |= wsExToolWindow | wsExNoActivate | wsExLayered
	ex &^= wsExAppWindow | wsExNoRedirBmp | wsExTransparent
	hudPin.mu.Lock()
	if !hudPin.noActivate {
		ex &^= wsExNoActivate
	}
	hudPin.mu.Unlock()
	setLong(hwnd, gwlExStyle, ex)
	procSetLayeredWindowAttributes.Call(uintptr(hwnd), hudChromaKey, 255, lwaColorKey)
	procSetClassLongPtrW.Call(uintptr(hwnd), idx(gclpHCursor), 0)

	dwmAttr(hwnd, dwmwaNcRenderingPolicy, dwmNcRpDisabled)
	dwmAttr(hwnd, dwmwaTransitionsForcedOff, 1)
	dwmAttr(hwnd, dwmwaWindowCornerPreference, dwmwcpDoNotRound)
	dwmAttr(hwnd, dwmwaBorderColor, dwmColorNone)
	dwmAttr(hwnd, dwmwaCaptionColor, dwmColorNone)

	flags := uintptr(swpNoActivate | swpFrameChanged)
	var x, y, w, h uintptr
	if place {
		gw, gh, gx, gy := hudGeometry()
		x, y, w, h = uintptr(gx), uintptr(gy), uintptr(gw), uintptr(gh)
	} else {
		flags |= swpNoMove | swpNoSize
	}
	target := hwndNoTopmost
	hudPin.mu.Lock()
	if hudPin.pinned {
		target = hwndTopmost
	}
	hudPin.mu.Unlock()
	procSetWindowPos.Call(uintptr(hwnd), target, x, y, w, h, flags)
}

type winPoint struct{ X, Y int32 }

func cursorOnHud() bool {
	hwnd := hudHWND()
	if hwnd == 0 {
		return false
	}
	// Clic droit / milieu : toujours League (ping, caméra) — jamais le menu Chrome.
	if keyDown(vkRButton) || keyDown(vkMButton) {
		return false
	}
	hudPin.mu.Lock()
	solid := hudPin.solid
	hudPin.mu.Unlock()
	if solid {
		return true
	}
	var pt winPoint
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procScreenToClient.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pt)))
	hudHits.mu.Lock()
	defer hudHits.mu.Unlock()
	for _, r := range hudHits.rs {
		if int(pt.X) >= r.X && int(pt.X) < r.X+r.W && int(pt.Y) >= r.Y && int(pt.Y) < r.Y+r.H {
			return true
		}
	}
	return false
}

func hudSyncPassThrough() {
	hwnd := hudHWND()
	if hwnd == 0 {
		return
	}
	on := cursorOnHud()
	ex := getLong(hwnd, gwlExStyle)
	want := ex
	if on {
		want &^= wsExTransparent
	} else {
		want |= wsExTransparent
	}
	if want != ex {
		setLong(hwnd, gwlExStyle, want)
	}
	if on {
		procSetCursor.Call(0)
	}
}

func applyPin() {
	hwnd := hudHWND()
	if hwnd == 0 {
		return
	}
	dressHUD(hwnd, true)
	hudPin.mu.Lock()
	hudPin.placed = true
	hudPin.mu.Unlock()
}

func pinLoop() {
	for {
		time.Sleep(1500 * time.Millisecond)
		hudPin.mu.Lock()
		pinned := hudPin.pinned
		hudPin.mu.Unlock()
		if !pinned {
			continue
		}
		// Fenêtre cachée : ne pas SetWindowPos / restyler, ça la réaffiche grise.
		if hudWantVisible() {
			applyPin()
		}
		hudSyncVisibility()
	}
}

func hudWantVisible() bool {
	hudPin.mu.Lock()
	hold := hudPin.hold
	hudPin.mu.Unlock()
	inputHub.mu.Lock()
	tab := inputHub.tab
	inputHub.mu.Unlock()
	return hold || tab || !liveGameActive()
}

func windowIsVisible(hwnd syscall.Handle) bool {
	if hwnd == 0 {
		return false
	}
	r, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
	return r != 0
}

func hudSyncVisibility() {
	hwnd := hudHWND()
	if hwnd == 0 {
		return
	}
	want := hudWantVisible()
	now := time.Now()

	hudPin.mu.Lock()
	if want {
		hudPin.hideAfter = time.Time{}
	} else if hudPin.shown {
		if hudPin.hideAfter.IsZero() {
			hudPin.hideAfter = now.Add(220 * time.Millisecond)
		}
		if now.Before(hudPin.hideAfter) {
			hudPin.mu.Unlock()
			return
		}
	}
	hudPin.shown = want
	hudPin.mu.Unlock()

	visible := windowIsVisible(hwnd)
	if want == visible {
		return
	}
	if want {
		procShowWindow.Call(uintptr(hwnd), swShowNA)
		procSetLayeredWindowAttributes.Call(uintptr(hwnd), hudChromaKey, 255, lwaColorKey)
		return
	}
	procShowWindow.Call(uintptr(hwnd), swHide)
}

func hudBeginDrag() {
	hwnd := hudHWND()
	if hwnd == 0 {
		return
	}
	procReleaseCapture.Call()
	procSendMessageW.Call(uintptr(hwnd), wmNCLButtonDown, htCaption, 0)
}

func hudCloseWindow() {
	hwnd := hudHWND()
	if hwnd != 0 {
		procPostMessageW.Call(uintptr(hwnd), wmClose, 0, 0)
	}
	hudPin.mu.Lock()
	hudPin.hwnd, hudPin.found, hudPin.placed = 0, false, false
	hudPin.mu.Unlock()
	killHUDChrome()
}

func hudProfileDir() string {
	profile := filepath.Join(os.TempDir(), "cdscout-hud-profile-v5")
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		profile = filepath.Join(local, "lol-cd-scout", "hud-profile-v5")
	}
	return profile
}

func killHUDChrome() {
	// Chrome --app survit à os.Exit : tuer tout process dont la ligne de commande
	// contient notre profil (unique), pas le Chrome personnel.
	ps := `Get-CimInstance Win32_Process | Where-Object { $_.CommandLine -match 'hud-profile-v' } | ForEach-Object { Stop-Process -Id $_.ProcessId -Force -ErrorAction SilentlyContinue }`
	cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", ps)
	cmd.SysProcAttr = windowsHiddenProcAttr()
	_ = cmd.Run()
}

func sysMetric(index int) int {
	v, _, _ := procGetSystemMetrics.Call(uintptr(index))
	return int(int32(v))
}

func hudGeometry() (w, h, x, y int) {
	return sysMetric(smCXScreen), sysMetric(smCYScreen), 0, 0
}

func chromiumPath() string {
	rels := []string{
		`Google\Chrome\Application\chrome.exe`,
		`Microsoft\Edge\Application\msedge.exe`,
		`BraveSoftware\Brave-Browser\Application\brave.exe`,
		`Vivaldi\Application\vivaldi.exe`,
		`Chromium\Application\chrome.exe`,
	}
	bases := []string{os.Getenv("PROGRAMFILES"), os.Getenv("PROGRAMFILES(X86)"), os.Getenv("LOCALAPPDATA")}
	for _, rel := range rels {
		for _, base := range bases {
			if base == "" {
				continue
			}
			p := filepath.Join(base, rel)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}

func hudOpen(url string) error {
	exe := chromiumPath()
	if exe == "" {
		return errors.New("aucun navigateur Chromium détecté (Chrome, Edge, Brave)")
	}
	w, h, x, y := hudGeometry()
	profile := hudProfileDir()
	cmd := exec.Command(exe,
		"--app="+url,
		"--user-data-dir="+profile,
		fmt.Sprintf("--window-size=%d,%d", w, h),
		fmt.Sprintf("--window-position=%d,%d", x, y),
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-windows10-custom-titlebar",
		"--disable-direct-composition",
		"--disable-gpu-compositing",
		"--disable-dev-tools",
		"--disable-features=Translate,MediaRouter,CalculateNativeWinOcclusion,Windows11MicaTitlebar,DevToolsAvailability",
		"--autoplay-policy=no-user-gesture-required",
	)
	cmd.SysProcAttr = windowsHiddenProcAttr()
	if err := cmd.Start(); err != nil {
		return err
	}
	_ = cmd.Process.Release()
	inputStart()
	go func() {
		var hwnd syscall.Handle
		for i := 0; i < 60; i++ {
			time.Sleep(150 * time.Millisecond)
			hwnd = findWindow(hudWindowTitle)
			if hwnd != 0 {
				break
			}
		}
		hudPin.mu.Lock()
		hudPin.placed, hudPin.hwnd = false, 0
		hudPin.mu.Unlock()
		hudSetPin(true, true)
		for i := 0; i < 20; i++ {
			applyPin()
			time.Sleep(200 * time.Millisecond)
		}
		hudSyncVisibility()
	}()
	return nil
}

type inputEvent struct {
	Seq  uint64 `json:"seq"`
	Kind string `json:"kind"`
	Slot int    `json:"slot"`
}

var inputHub = struct {
	mu      sync.Mutex
	seq     uint64
	events  []inputEvent
	tab     bool
	started bool
}{}

func keyDown(vk int) bool {
	r, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return r&0x8000 != 0
}

// vkHoldKey : touche ² (activation HUD). VkKeyScanW suit le layout ;
// repli OEM_3 = position physique au-dessus de Tab en AZERTY.
func vkHoldKey() int {
	r, _, _ := procVkKeyScanW.Call(uintptr('²'))
	if int16(r) != -1 && r&0xFF != 0 {
		return int(r & 0xFF)
	}
	return vkOEM3
}

func inputStart() {
	inputHub.mu.Lock()
	if inputHub.started {
		inputHub.mu.Unlock()
		return
	}
	inputHub.started = true
	inputHub.mu.Unlock()

	go func() {
		prev := [5]bool{}
		vkHold := vkHoldKey()
		for {
			alt, shift := keyDown(vkMenu), keyDown(vkShift)
			tab := keyDown(vkHold)
			for i := 0; i < 5; i++ {
				down := keyDown(0x31 + i)
				if down && !prev[i] && alt {
					kind := "flash"
					if shift {
						kind = "ult"
					}
					pushInput(kind, i+1)
				}
				prev[i] = down
			}
			inputHub.mu.Lock()
			inputHub.tab = tab
			inputHub.mu.Unlock()
			hudSyncVisibility()
			hudSyncPassThrough()
			time.Sleep(30 * time.Millisecond)
		}
	}()
}

func pushInput(kind string, slot int) {
	inputHub.mu.Lock()
	defer inputHub.mu.Unlock()
	inputHub.seq++
	inputHub.events = append(inputHub.events, inputEvent{Seq: inputHub.seq, Kind: kind, Slot: slot})
	if len(inputHub.events) > inputEventBuffer {
		inputHub.events = inputHub.events[len(inputHub.events)-inputEventBuffer:]
	}
}

func inputSince(since uint64) (tab bool, seq uint64, out []inputEvent) {
	inputStart()
	inputHub.mu.Lock()
	defer inputHub.mu.Unlock()
	out = []inputEvent{}
	for _, e := range inputHub.events {
		if e.Seq > since {
			out = append(out, e)
		}
	}
	return inputHub.tab, inputHub.seq, out
}
