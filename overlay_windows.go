//go:build windows

package main

// HUD en jeu : canvas plein écran WS_POPUP, widgets indépendants.
// Click-through hors des blocs : SetWindowRgn sur le parent ET les HWND enfants
// WebView2 (Chrome_WidgetWin_* / Intermediate D3D Window) + WM_NCHITTEST.
// Le hit-test Windows interroge l'enfant d'abord : une RGN parent seule ne perce pas Chromium.
// Fond contrôleur A:0 (pas de slab navy). Jamais toggler WS_EX_TRANSPARENT / LAYERED au survol.
// Alt+1–5 Flash, Alt+Maj+1–5 ult, Tab cartes.

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/jchv/go-webview2/pkg/edge"
	"golang.org/x/sys/windows"
)

const hudWindowTitle = "CD Scout HUD"

var (
	user32                    = syscall.NewLazyDLL("user32.dll")
	gdi32                     = syscall.NewLazyDLL("gdi32.dll")
	dwmapi                    = syscall.NewLazyDLL("dwmapi.dll")
	kernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procIsWindow              = user32.NewProc("IsWindow")
	procEnumWindows           = user32.NewProc("EnumWindows")
	procEnumChildWindows      = user32.NewProc("EnumChildWindows")
	procGetClassNameW         = user32.NewProc("GetClassNameW")
	procGetWindowTextW        = user32.NewProc("GetWindowTextW")
	procOutputDebugStringW    = kernel32.NewProc("OutputDebugStringW")
	procSetWindowPos          = user32.NewProc("SetWindowPos")
	procGetAsyncKeyState      = user32.NewProc("GetAsyncKeyState")
	procGetSystemMetrics      = user32.NewProc("GetSystemMetrics")
	procShowWindow            = user32.NewProc("ShowWindow")
	procPostMessageW          = user32.NewProc("PostMessageW")
	procReleaseCapture        = user32.NewProc("ReleaseCapture")
	procGetCursorPos          = user32.NewProc("GetCursorPos")
	procGetForegroundWindow   = user32.NewProc("GetForegroundWindow")
	procGetWindowRect         = user32.NewProc("GetWindowRect")
	procScreenToClient        = user32.NewProc("ScreenToClient")
	procRegisterClassExW      = user32.NewProc("RegisterClassExW")
	procCreateWindowExW       = user32.NewProc("CreateWindowExW")
	procDefWindowProcW        = user32.NewProc("DefWindowProcW")
	procGetMessageW           = user32.NewProc("GetMessageW")
	procTranslateMessage      = user32.NewProc("TranslateMessage")
	procDispatchMessageW      = user32.NewProc("DispatchMessageW")
	procPostQuitMessage       = user32.NewProc("PostQuitMessage")
	procDestroyWindow         = user32.NewProc("DestroyWindow")
	procLoadCursorW           = user32.NewProc("LoadCursorW")
	procGetStockObject        = gdi32.NewProc("GetStockObject")
	procCreateRectRgn         = gdi32.NewProc("CreateRectRgn")
	procCombineRgn            = gdi32.NewProc("CombineRgn")
	procDeleteObject          = gdi32.NewProc("DeleteObject")
	procSetWindowRgn          = user32.NewProc("SetWindowRgn")
	procDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
	ole32                     = syscall.NewLazyDLL("ole32.dll")
	procCoInitializeEx        = ole32.NewProc("CoInitializeEx")
	procCoUninitialize        = ole32.NewProc("CoUninitialize")
)

const (
	hwndTopmost   = ^uintptr(0)
	swpNoSize     = 0x0001
	swpNoMove     = 0x0002
	swpNoZOrder   = 0x0004
	swpNoActivate = 0x0010

	wsPopup        = 0x80000000
	wsClipChildren = 0x02000000
	wsClipSiblings = 0x04000000
	wsExNoActivate = 0x08000000
	wsExToolWindow = 0x00000080
	wsExTopmost    = 0x00000008

	csHRedraw = 0x0002
	csVRedraw = 0x0001

	swShowNA = 8

	wmDestroy       = 0x0002
	wmMove          = 0x0003
	wmSize          = 0x0005
	wmClose         = 0x0010
	wmEraseBkgnd    = 0x0014
	wmMouseActivate = 0x0021
	wmDisplayChange = 0x007E
	wmExitSizeMove  = 0x0232
	wmApp           = 0x8000
	wmHudBounds     = wmApp + 1
	wmHudDrag       = wmApp + 2
	wmHudClose      = wmApp + 3
	wmHudTopmost    = wmApp + 4
	wmHudReset      = wmApp + 5
	wmHudHits       = wmApp + 6

	maNoActivate = 3
	idcArrow     = 32512
	blackBrush   = 4
	rgnOr        = 2
	vkLButton    = 0x01
	vkTab        = 0x09
	vkShift      = 0x10
	vkControl    = 0x11
	vkMenu       = 0x12

	wmNCHitTest   = 0x0084
	htClient      = 1
	htTransparent = ^uintptr(0) // HTTRANSPARENT = -1

	nullBrush = 5

	smCXScreen = 0
	smCYScreen = 1

	dwmwaWindowCornerPreference = 33
	dwmwcpDoNotRound            = 1
	dwmwaBorderColor            = 34
	dwmwaCaptionColor           = 35
	dwmColorNone                = 0xFFFFFFFE

	inputEventBuffer = 32
	hudClassName     = "CdScoutHudWidget"
	coInitApartment  = 0x2
)

var hudWndProcCB = syscall.NewCallback(hudWndProc)

type winPoint struct{ X, Y int32 }
type winRect struct{ Left, Top, Right, Bottom int32 }

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   syscall.Handle
	Icon       syscall.Handle
	Cursor     syscall.Handle
	Background syscall.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     syscall.Handle
}

type winMsg struct {
	Hwnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      winPoint
	Private uint32
}

func setDPIAware() {
	if p := user32.NewProc("SetProcessDpiAwarenessContext"); p.Find() == nil {
		p.Call(^uintptr(3)) // PER_MONITOR_AWARE_V2
		return
	}
	shcore := syscall.NewLazyDLL("shcore.dll")
	if p := shcore.NewProc("SetProcessDpiAwareness"); p.Find() == nil {
		p.Call(2) // PROCESS_PER_MONITOR_DPI_AWARE
		return
	}
	user32.NewProc("SetProcessDPIAware").Call()
}

func hudSupported() bool { return true }

func dwmAttr(hwnd syscall.Handle, attr uint32, value uint32) {
	procDwmSetWindowAttribute.Call(uintptr(hwnd), uintptr(attr), uintptr(unsafe.Pointer(&value)), 4)
}

func hudSetControllerTransparent(cr *edge.Chromium) {
	if cr == nil {
		return
	}
	if c := cr.GetController(); c != nil {
		if c2 := c.GetICoreWebView2Controller2(); c2 != nil {
			// A:0 : pas de slab navy. Si CSS opacity ne laisse pas voir la map
			// à travers les widgets, fallback = WS_EX_LAYERED dans CreateWindowEx
			// (jamais via SetWindowLong) + SetLayeredWindowAttributes.
			_ = c2.PutDefaultBackgroundColor(edge.COREWEBVIEW2_COLOR{A: 0, R: 18, G: 28, B: 44})
		}
	}
}

func sysMetric(index int) int {
	v, _, _ := procGetSystemMetrics.Call(uintptr(index))
	return int(int32(v))
}

func windowAlive(hwnd syscall.Handle) bool {
	if hwnd == 0 {
		return false
	}
	r, _, _ := procIsWindow.Call(uintptr(hwnd))
	return r != 0
}

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

var (
	hudChildMu      sync.Mutex
	hudChildFound   []syscall.Handle
	hudChildNames   []string
	hudChildLogOnce sync.Once
	hudChildEnumCB  = syscall.NewCallback(hudEnumChildProc)
)

func hudIsWebViewChildClass(cls string) bool {
	switch {
	case strings.HasPrefix(cls, "Chrome_WidgetWin_"):
		return true
	case cls == "Intermediate D3D Window":
		return true
	case strings.HasPrefix(cls, "Chrome_RenderWidgetHostHWND"):
		return true
	default:
		return false
	}
}

func hudEnumChildProc(hwnd syscall.Handle, _ uintptr) uintptr {
	buf := make([]uint16, 256)
	n, _, _ := procGetClassNameW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return 1
	}
	cls := syscall.UTF16ToString(buf[:n])
	hudChildNames = append(hudChildNames, cls)
	if hudIsWebViewChildClass(cls) {
		hudChildFound = append(hudChildFound, hwnd)
	}
	return 1
}

func hudWebViewChildHWNDs(parent syscall.Handle) []syscall.Handle {
	if !windowAlive(parent) {
		return nil
	}
	hudChildMu.Lock()
	defer hudChildMu.Unlock()
	hudChildFound = hudChildFound[:0]
	hudChildNames = hudChildNames[:0]
	procEnumChildWindows.Call(uintptr(parent), hudChildEnumCB, 0)
	names := append([]string(nil), hudChildNames...)
	out := append([]syscall.Handle(nil), hudChildFound...)
	if len(names) > 0 {
		hudChildLogOnce.Do(func() {
			msg, err := syscall.UTF16PtrFromString("cdscout hud WebView2 children: " + strings.Join(names, ", "))
			if err == nil {
				procOutputDebugStringW.Call(uintptr(unsafe.Pointer(msg)))
			}
		})
	}
	return out
}

var hudPin = struct {
	mu         sync.Mutex
	pinned     bool
	hold       bool
	found      bool
	loop       bool
	userClosed bool
	hwnd       syscall.Handle
	web        *edge.Chromium
}{}

var autoHud = struct {
	mu      sync.Mutex
	key     string
	lastTry time.Time
}{}

// hudCloseIdle ferme l'overlay sans bloquer la réouverture auto de la
// prochaine partie (contrairement à la croix, qui pose userClosed).
func hudCloseIdle() {
	hudPin.mu.Lock()
	hudPin.userClosed = false
	hwnd := hudPin.hwnd
	hudPin.mu.Unlock()
	if windowAlive(hwnd) {
		procPostMessageW.Call(uintptr(hwnd), wmHudClose, 0, 0)
	}
}

// autoOpenHudForGame ouvre le widget dès qu'une partie live est détectée.
// Hors partie (et hors démo / devmode), l'overlay se ferme.
func autoOpenHudForGame(key string) {
	key = hudKeepKey(key)
	autoHud.mu.Lock()
	if key == "" {
		autoHud.key = ""
		autoHud.mu.Unlock()
		hudCloseIdle()
		return
	}
	hudPin.mu.Lock()
	closed, loop, hwnd := hudPin.userClosed, hudPin.loop, hudPin.hwnd
	hudPin.mu.Unlock()
	if closed && autoHud.key == key {
		autoHud.mu.Unlock()
		return
	}
	if windowAlive(hwnd) || loop {
		autoHud.key = key
		autoHud.mu.Unlock()
		return
	}
	if autoHud.key == key && time.Since(autoHud.lastTry) < 8*time.Second {
		autoHud.mu.Unlock()
		return
	}
	autoHud.key = key
	autoHud.lastTry = time.Now()
	autoHud.mu.Unlock()
	go func() { _ = hudOpen(hudURL()) }()
}

var hudHits = struct {
	mu sync.Mutex
	rs []hudHit
}{}

func hudSetHits(rs []hudHit) {
	rs = hudHitsClean(rs)
	hudHits.mu.Lock()
	same := hudHitsEqual(hudHits.rs, rs)
	if !same {
		hudHits.rs = rs
	}
	hudHits.mu.Unlock()
	if same {
		return
	}
	if h := hudHWND(); windowAlive(h) {
		procPostMessageW.Call(uintptr(h), wmHudHits, 0, 0)
	}
}

func hudBuildRgn(rs []hudHit) uintptr {
	if len(rs) == 0 {
		rgn, _, _ := procCreateRectRgn.Call(0, 0, 1, 1)
		return rgn
	}
	r0 := rs[0]
	rgn, _, _ := procCreateRectRgn.Call(uintptr(r0.X), uintptr(r0.Y), uintptr(r0.X+r0.W), uintptr(r0.Y+r0.H))
	if rgn == 0 {
		return 0
	}
	for _, h := range rs[1:] {
		part, _, _ := procCreateRectRgn.Call(uintptr(h.X), uintptr(h.Y), uintptr(h.X+h.W), uintptr(h.Y+h.H))
		if part == 0 {
			continue
		}
		procCombineRgn.Call(rgn, rgn, part, rgnOr)
		procDeleteObject.Call(part)
	}
	return rgn
}

func applyHudRegion(hwnd syscall.Handle) {
	if !windowAlive(hwnd) {
		return
	}
	hudHits.mu.Lock()
	rs := append([]hudHit(nil), hudHits.rs...)
	hudHits.mu.Unlock()
	// SetWindowRgn prend ownership de l'HRGN : une copie par HWND (parent + enfants Chromium).
	targets := append([]syscall.Handle{hwnd}, hudWebViewChildHWNDs(hwnd)...)
	for _, h := range targets {
		if !windowAlive(h) {
			continue
		}
		rgn := hudBuildRgn(rs)
		if rgn == 0 {
			continue
		}
		procSetWindowRgn.Call(uintptr(h), rgn, 1)
	}
}

func hudNCHitTest(hwnd syscall.Handle, lParam uintptr) uintptr {
	hudHits.mu.Lock()
	rs := hudHits.rs
	hudHits.mu.Unlock()
	if len(rs) == 0 {
		return htTransparent
	}
	pt := winPoint{
		X: int32(int16(lParam)),
		Y: int32(int16(lParam >> 16)),
	}
	procScreenToClient.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pt)))
	if hudHitContains(rs, int(pt.X), int(pt.Y)) {
		return htClient
	}
	return htTransparent
}

func hudStatus() (pinned, noActivate, found bool) {
	hudPin.mu.Lock()
	defer hudPin.mu.Unlock()
	return hudPin.pinned, true, hudPin.found
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
}

func hudSetPin(pinned, _ bool) {
	hudPin.mu.Lock()
	hudPin.pinned = pinned
	hwnd := hudPin.hwnd
	hudPin.mu.Unlock()
	if windowAlive(hwnd) {
		procPostMessageW.Call(uintptr(hwnd), wmHudTopmost, 0, 0)
	}
}

func hudHWND() syscall.Handle {
	hudPin.mu.Lock()
	defer hudPin.mu.Unlock()
	if windowAlive(hudPin.hwnd) {
		return hudPin.hwnd
	}
	return 0
}

var hudSize = struct {
	mu        sync.Mutex
	w, h      int
	x, y      int
	hasPos    bool
	reset     bool
	ready     chan error
	readyOnce sync.Once
}{w: hudMiniW, h: hudMiniH}

var hudWidgets = struct {
	mu sync.Mutex
	g  hudGeomDisk
}{}

var hudDragLive = struct {
	mu sync.Mutex
	on bool
	r  hudDragReq
}{}

func hudScreenSize() (int, int) {
	sw, sh := sysMetric(smCXScreen), sysMetric(smCYScreen)
	if sw < 1 {
		sw = 1920
	}
	if sh < 1 {
		sh = 1080
	}
	return sw, sh
}

func hudSetBounds(_, _ int) {
	sw, sh := hudScreenSize()
	hudSize.mu.Lock()
	hudSize.w, hudSize.h = sw, sh
	hudSize.x, hudSize.y, hudSize.hasPos = 0, 0, true
	hudSize.mu.Unlock()
	if h := hudHWND(); windowAlive(h) {
		procPostMessageW.Call(uintptr(h), wmHudBounds, 0, 0)
	}
}

func hudResetPos() {
	sw, sh := hudScreenSize()
	g := hudGeomDisk{V: 4, Widgets: hudDefaultWidgets(sw, sh)}
	hudGeomReplace(g)
	hudEvalGeom(g)
	if h := hudHWND(); windowAlive(h) {
		procPostMessageW.Call(uintptr(h), wmHudBounds, 0, 0)
	}
}

func hudBeginDrag() {}

func hudBeginWidgetDrag(req hudDragReq) {
	if !validHudWidgetID(req.ID) {
		return
	}
	hwnd := hudHWND()
	if windowAlive(hwnd) {
		go hudWidgetDragLoop(hwnd, req)
	}
}

func hudEval(js string) {
	hudPin.mu.Lock()
	web := hudPin.web
	hudPin.mu.Unlock()
	if web != nil && js != "" {
		web.Eval(js)
	}
}

func hudEvalMove(id string, g hudWidgetGeom) {
	b, err := json.Marshal(map[string]any{"id": id, "x": g.X, "y": g.Y, "scale": g.scaleOr1()})
	if err != nil {
		return
	}
	hudEval("window.__hudMove&&window.__hudMove(" + string(b) + ")")
}

func hudEvalGeom(g hudGeomDisk) {
	b, err := json.Marshal(g)
	if err != nil {
		return
	}
	hudEval("window.__hudApplyGeom&&window.__hudApplyGeom(" + string(b) + ")")
}

func hudDragLiveCopy() (hudDragReq, bool) {
	hudDragLive.mu.Lock()
	defer hudDragLive.mu.Unlock()
	if !hudDragLive.on {
		return hudDragReq{}, false
	}
	return hudDragLive.r, true
}

func hudCloseWindow() {
	hudPin.mu.Lock()
	hudPin.userClosed = true
	hwnd := hudPin.hwnd
	hudPin.mu.Unlock()
	if windowAlive(hwnd) {
		procPostMessageW.Call(uintptr(hwnd), wmHudClose, 0, 0)
		return
	}
	hudPin.mu.Lock()
	hudPin.hwnd, hudPin.found, hudPin.loop, hudPin.web = 0, false, false, nil
	hudPin.mu.Unlock()
}

func hudProfileDir() string {
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		return filepath.Join(local, "lol-cd-scout", "hud-webview2")
	}
	return filepath.Join(os.TempDir(), "cdscout-hud-webview2")
}

func hudGeomPath() string {
	return filepath.Join(filepath.Dir(hudProfileDir()), "hud-geom.json")
}

func hudGeomSnapshot() hudGeomDisk {
	sw, sh := hudScreenSize()
	hudWidgets.mu.Lock()
	g := hudGeomMerge(hudWidgets.g, sw, sh)
	hudWidgets.mu.Unlock()
	return g
}

func hudGeomReplace(g hudGeomDisk) {
	sw, sh := hudScreenSize()
	g = hudGeomMerge(g, sw, sh)
	hudWidgets.mu.Lock()
	hudWidgets.g = g
	hudWidgets.mu.Unlock()
	saveHudGeomFile(g)
}

func hudWidgetsPut(id string, w hudWidgetGeom) {
	if !validHudWidgetID(id) {
		return
	}
	sw, sh := hudScreenSize()
	w = clampHudWidget(w, sw, sh)
	hudWidgets.mu.Lock()
	if hudWidgets.g.Widgets == nil {
		hudWidgets.g.Widgets = map[string]hudWidgetGeom{}
	}
	hudWidgets.g.V = 3
	hudWidgets.g.Widgets[id] = w
	hudWidgets.mu.Unlock()
	hudDragLive.mu.Lock()
	hudDragLive.on = true
	hudDragLive.r = hudDragReq{ID: id, X: w.X, Y: w.Y, Scale: w.scaleOr1()}
	hudDragLive.mu.Unlock()
}

func loadHudGeomFile() hudGeomDisk {
	sw, sh := hudScreenSize()
	b, err := os.ReadFile(hudGeomPath())
	if err != nil {
		return hudGeomDisk{V: 4, Widgets: hudDefaultWidgets(sw, sh)}
	}
	var g hudGeomDisk
	if json.Unmarshal(b, &g) != nil {
		return hudGeomDisk{V: 4, Widgets: hudDefaultWidgets(sw, sh)}
	}
	return hudGeomMerge(g, sw, sh)
}

func saveHudGeomFile(g hudGeomDisk) {
	_ = os.MkdirAll(filepath.Dir(hudGeomPath()), 0o755)
	b, err := json.Marshal(g)
	if err != nil {
		return
	}
	_ = os.WriteFile(hudGeomPath(), b, 0o644)
}

func saveHudGeom(_ syscall.Handle) {
	saveHudGeomFile(hudGeomSnapshot())
}

func webview2Available() bool {
	bases := []string{
		os.Getenv("PROGRAMFILES(X86)"),
		os.Getenv("PROGRAMFILES"),
		os.Getenv("LOCALAPPDATA"),
	}
	rels := []string{
		`Microsoft\EdgeWebView\Application`,
		`Microsoft\Edge\Application`,
	}
	for _, rel := range rels {
		for _, base := range bases {
			if base == "" {
				continue
			}
			if st, err := os.Stat(filepath.Join(base, rel)); err == nil && st.IsDir() {
				return true
			}
		}
	}
	return false
}

func hudOpen(url string) error {
	if !webview2Available() {
		return errors.New("WebView2 introuvable : installe Microsoft Edge WebView2 Runtime")
	}
	hudPin.mu.Lock()
	hudPin.userClosed = false
	if hudPin.loop && windowAlive(hudPin.hwnd) {
		hudPin.mu.Unlock()
		return nil
	}
	if hudPin.loop {
		hudPin.mu.Unlock()
		return errors.New("HUD en cours d'ouverture")
	}
	hudPin.loop = true
	hudPin.pinned = true
	hudPin.mu.Unlock()

	ready := make(chan error, 1)
	hudSize.mu.Lock()
	hudSize.ready = ready
	hudSize.readyOnce = sync.Once{}
	hudSize.mu.Unlock()

	go hudUIThread(url)
	inputStart()
	select {
	case err := <-ready:
		return err
	case <-time.After(20 * time.Second):
		hudPin.mu.Lock()
		hudPin.loop = false
		hudPin.mu.Unlock()
		return errors.New("HUD : délai d'ouverture WebView2")
	}
}

func signalHudReady(err error) {
	hudSize.mu.Lock()
	ch := hudSize.ready
	hudSize.mu.Unlock()
	if ch == nil {
		return
	}
	hudSize.readyOnce.Do(func() { ch <- err })
}

func hudUIThread(url string) {
	runtime.LockOSThread()
	defer func() {
		hudPin.mu.Lock()
		hudPin.hwnd, hudPin.found, hudPin.loop, hudPin.web = 0, false, false, nil
		hudPin.mu.Unlock()
	}()

	procCoInitializeEx.Call(0, coInitApartment)
	defer procCoUninitialize.Call()

	hwnd, err := createHudWindow()
	if err != nil {
		signalHudReady(err)
		return
	}

	cr := edge.NewChromium()
	cr.DataPath = hudProfileDir()
	cr.AcceleratorKeyCallback = func(vk uint) bool {
		return vk == 0x7B // F12
	}
	hudPin.mu.Lock()
	hudPin.hwnd, hudPin.found, hudPin.web = hwnd, true, cr
	hudPin.mu.Unlock()
	procShowWindow.Call(uintptr(hwnd), swShowNA)
	applyHudTopmost(hwnd)

	if !cr.Embed(uintptr(hwnd)) {
		signalHudReady(errors.New("WebView2 : impossible d'embarquer le moteur"))
		procDestroyWindow.Call(uintptr(hwnd))
		return
	}
	if settings, err := cr.GetSettings(); err == nil {
		_ = settings.PutAreDevToolsEnabled(false)
		_ = settings.PutAreDefaultContextMenusEnabled(false)
		_ = settings.PutIsStatusBarEnabled(false)
		_ = settings.PutIsZoomControlEnabled(false)
		_ = settings.PutAreBrowserAcceleratorKeysEnabled(false)
		_ = settings.PutIsSwipeNavigationEnabled(false)
	}
	hudSetControllerTransparent(cr)
	// Cette WebView2 n'affiche que le widget : hudmode même si l'URL perd ?hud=1.
	cr.Init(`(function(){document.documentElement.classList.add('hudmode');function b(){if(document.body)document.body.classList.add('hudmode');document.title='CD Scout HUD'}if(document.body)b();else document.addEventListener('DOMContentLoaded',b)})()`)
	cr.NavigationCompletedCallback = func(*edge.ICoreWebView2, *edge.ICoreWebView2NavigationCompletedEventArgs) {
		hudSetControllerTransparent(cr)
		cr.Resize()
		cr.Eval(`document.documentElement.classList.add('hudmode');if(document.body)document.body.classList.add('hudmode');document.title='CD Scout HUD'`)
		hudEvalGeom(hudGeomSnapshot())
		if h := hudHWND(); windowAlive(h) {
			procPostMessageW.Call(uintptr(h), wmHudHits, 0, 0)
		}
	}
	cr.Resize()
	cr.Navigate(url)
	applyHudRegion(hwnd)
	signalHudReady(nil)

	go pinLoop()

	var msg winMsg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func createHudWindow() (syscall.Handle, error) {
	var inst windows.Handle
	if err := windows.GetModuleHandleEx(0, nil, &inst); err != nil {
		return 0, err
	}
	className, _ := syscall.UTF16PtrFromString(hudClassName)
	title, _ := syscall.UTF16PtrFromString(hudWindowTitle)
	cursor, _, _ := procLoadCursorW.Call(0, uintptr(idcArrow))
	brush, _, _ := procGetStockObject.Call(nullBrush)

	wc := wndClassEx{
		Size:       uint32(unsafe.Sizeof(wndClassEx{})),
		Style:      csHRedraw | csVRedraw,
		WndProc:    hudWndProcCB,
		Instance:   syscall.Handle(inst),
		Cursor:     syscall.Handle(cursor),
		Background: syscall.Handle(brush),
		ClassName:  className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	sw, sh := hudScreenSize()
	g := loadHudGeomFile()
	hudWidgets.mu.Lock()
	hudWidgets.g = g
	hudWidgets.mu.Unlock()
	hudSize.mu.Lock()
	hudSize.x, hudSize.y, hudSize.w, hudSize.h, hudSize.hasPos = 0, 0, sw, sh, true
	hudSize.mu.Unlock()

	ex := uintptr(wsExTopmost | wsExToolWindow | wsExNoActivate)
	style := uintptr(wsPopup | wsClipChildren | wsClipSiblings)
	hwnd, _, err2 := procCreateWindowExW.Call(
		ex,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		style,
		0, 0, uintptr(sw), uintptr(sh),
		0, 0, uintptr(inst), 0,
	)
	if hwnd == 0 {
		if err2 != nil && err2 != syscall.Errno(0) {
			return 0, err2
		}
		return 0, errors.New("CreateWindowEx HUD")
	}
	hnd := syscall.Handle(hwnd)
	dwmAttr(hnd, dwmwaWindowCornerPreference, dwmwcpDoNotRound)
	dwmAttr(hnd, dwmwaBorderColor, dwmColorNone)
	dwmAttr(hnd, dwmwaCaptionColor, dwmColorNone)
	return hnd, nil
}

func hudWndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmEraseBkgnd:
		return 1
	case wmMouseActivate:
		return maNoActivate
	case wmNCHitTest:
		return hudNCHitTest(hwnd, lParam)
	case wmSize, wmHudBounds:
		applyHudBounds(hwnd)
		hudPin.mu.Lock()
		web := hudPin.web
		hudPin.mu.Unlock()
		if web != nil {
			web.Resize()
		}
		applyHudRegion(hwnd)
		return 0
	case wmMove:
		hudPin.mu.Lock()
		web := hudPin.web
		hudPin.mu.Unlock()
		if web != nil {
			_ = web.NotifyParentWindowPositionChanged()
		}
		return 0
	case wmExitSizeMove, wmDisplayChange:
		saveHudGeom(hwnd)
		applyHudBounds(hwnd)
		return 0
	case wmHudDrag:
		return 0
	case wmHudHits:
		applyHudRegion(hwnd)
		return 0
	case wmHudTopmost:
		applyHudTopmost(hwnd)
		return 0
	case wmHudReset:
		applyHudBounds(hwnd)
		return 0
	case wmHudClose, wmClose:
		procDestroyWindow.Call(uintptr(hwnd))
		return 0
	case wmDestroy:
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

func applyHudBounds(hwnd syscall.Handle) {
	if !windowAlive(hwnd) {
		return
	}
	sw, sh := hudScreenSize()
	var rc winRect
	procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rc)))
	curX, curY := int(rc.Left), int(rc.Top)
	curW, curH := int(rc.Right-rc.Left), int(rc.Bottom-rc.Top)
	if curX == 0 && curY == 0 && curW == sw && curH == sh {
		return
	}
	hudSize.mu.Lock()
	hudSize.x, hudSize.y, hudSize.w, hudSize.h, hudSize.hasPos = 0, 0, sw, sh, true
	hudSize.mu.Unlock()
	procSetWindowPos.Call(uintptr(hwnd), hwndTopmost, 0, 0, uintptr(sw), uintptr(sh), swpNoActivate)
	hudPin.mu.Lock()
	web := hudPin.web
	hudPin.mu.Unlock()
	if web != nil {
		web.Resize()
	}
	applyHudRegion(hwnd)
}

func applyHudTopmost(hwnd syscall.Handle) {
	hudPin.mu.Lock()
	pinned := hudPin.pinned
	hudPin.mu.Unlock()
	if !pinned || !windowAlive(hwnd) {
		return
	}
	procSetWindowPos.Call(uintptr(hwnd), hwndTopmost, 0, 0, 0, 0, swpNoSize|swpNoMove|swpNoActivate)
}

func clientCursor(hwnd syscall.Handle) (x, y int, ok bool) {
	var pt winPoint
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procScreenToClient.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pt)))
	return int(pt.X), int(pt.Y), true
}

func hudWidgetDragLoop(hwnd syscall.Handle, req hudDragReq) {
	if !windowAlive(hwnd) || !validHudWidgetID(req.ID) {
		return
	}
	cx0, cy0, ok := clientCursor(hwnd)
	if !ok {
		return
	}
	start := hudWidgetGeom{X: req.X, Y: req.Y, Scale: req.Scale}
	if start.Scale <= 0 {
		start.Scale = 1
	}
	d0 := math.Hypot(float64(cx0-start.X), float64(cy0-start.Y))
	if d0 < 8 {
		d0 = 8
	}
	procReleaseCapture.Call()
	hudDragLive.mu.Lock()
	hudDragLive.on = true
	hudDragLive.r = req
	hudDragLive.mu.Unlock()
	defer func() {
		hudDragLive.mu.Lock()
		hudDragLive.on = false
		hudDragLive.mu.Unlock()
		saveHudGeom(hwnd)
	}()
	for keyDown(vkLButton) {
		if !windowAlive(hwnd) {
			return
		}
		cx, cy, ok := clientCursor(hwnd)
		if !ok {
			time.Sleep(12 * time.Millisecond)
			continue
		}
		g := start
		if req.Mode == "resize" {
			d1 := math.Hypot(float64(cx-start.X), float64(cy-start.Y))
			g.Scale = start.Scale * d1 / d0
		} else {
			g.X = start.X + (cx - cx0)
			g.Y = start.Y + (cy - cy0)
		}
		hudWidgetsPut(req.ID, g)
		hudWidgets.mu.Lock()
		cur := hudWidgets.g.Widgets[req.ID]
		hudWidgets.mu.Unlock()
		hudEvalMove(req.ID, cur)
		time.Sleep(12 * time.Millisecond)
	}
}

func pinLoop() {
	for {
		time.Sleep(2 * time.Second)
		hwnd := hudHWND()
		if hwnd == 0 {
			return
		}
		applyHudTopmost(hwnd)
		// L'enfant Chromium peut être recréé ; reclipper depuis le thread UI.
		procPostMessageW.Call(uintptr(hwnd), wmHudHits, 0, 0)
	}
}

func leagueForeground() bool {
	h := leagueGameHWND()
	if h == 0 {
		return false
	}
	fg, _, _ := procGetForegroundWindow.Call()
	return syscall.Handle(fg) == h
}

func watchAllsum(step *int, held *bool, last *time.Time) {
	if !leagueForeground() || keyDown(vkMenu) || keyDown(vkControl) {
		*step, *held = 0, false
		return
	}
	if *step > 0 && time.Since(*last) > 900*time.Millisecond {
		*step, *held = 0, false
	}
	expect := [...]int{0xBF, 0x41, 0x4C, 0x4C, 0x53, 0x55, 0x4D} // / A L L S U M
	if *step >= len(expect) {
		*step = 0
		return
	}
	down := keyDown(expect[*step])
	if down && !*held {
		*step++
		*last = time.Now()
		*held = true
		if *step == len(expect) {
			pushInput("allsum", 0)
			*step, *held = 0, false
		}
		return
	}
	if !down {
		*held = false
	}
}

func keyDown(vk int) bool {
	r, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return r&0x8000 != 0
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
		allsumStep, allsumHeld := 0, false
		var allsumLast time.Time
		for {
			alt, shift := keyDown(vkMenu), keyDown(vkShift)
			tab := keyDown(vkTab)
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
			watchAllsum(&allsumStep, &allsumHeld, &allsumLast)
			inputHub.mu.Lock()
			inputHub.tab = tab
			inputHub.mu.Unlock()
			time.Sleep(16 * time.Millisecond)
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
