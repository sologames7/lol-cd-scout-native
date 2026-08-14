package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// hudDevAllowed : le forçage overlay n'existe que sur un build local
// (`appVersion == "dev"`, donc `run.ps1`) ou CDSCOUT_DEV=1. Les exe
// GitHub Release ignorent « devmode on ».
func hudDevAllowed() bool {
	if v := strings.TrimSpace(os.Getenv("CDSCOUT_DEV")); v == "1" || strings.EqualFold(v, "true") {
		return true
	}
	v := strings.TrimSpace(appVersion)
	return v == "" || strings.EqualFold(v, "dev")
}

var hudDev = struct {
	mu    sync.Mutex
	force bool
}{}

func hudDevFile() string {
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		return filepath.Join(local, "lol-cd-scout", "devmode")
	}
	return filepath.Join(os.TempDir(), "cdscout-devmode")
}

func hudForceOn() bool {
	hudDev.mu.Lock()
	defer hudDev.mu.Unlock()
	return hudDev.force
}

func hudForceLoad() {
	if !hudDevAllowed() {
		return
	}
	b, err := os.ReadFile(hudDevFile())
	if err != nil {
		return
	}
	on := strings.TrimSpace(string(b)) == "1"
	hudDev.mu.Lock()
	hudDev.force = on
	hudDev.mu.Unlock()
}

func hudSetForce(on bool) bool {
	if !hudDevAllowed() {
		return false
	}
	hudDev.mu.Lock()
	hudDev.force = on
	hudDev.mu.Unlock()
	if on {
		_ = os.MkdirAll(filepath.Dir(hudDevFile()), 0700)
		_ = os.WriteFile(hudDevFile(), []byte("1\n"), 0600)
	} else {
		_ = os.Remove(hudDevFile())
	}
	return true
}

func hudMayOpen() bool {
	if hudForceOn() || demoLiveOn() {
		return true
	}
	return getLiveState().Active
}

func hudKeepKey(liveKey string) string {
	if liveKey != "" {
		return liveKey
	}
	if demoLiveOn() {
		return "demo"
	}
	if hudForceOn() {
		return "dev"
	}
	return ""
}

func apiHUDDevMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST", http.StatusMethodNotAllowed)
		return
	}
	if !hudDevAllowed() {
		http.Error(w, "devmode réservé au build local", http.StatusForbidden)
		return
	}
	var body struct {
		On *bool `json:"on"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 256)).Decode(&body)
	on := true
	if q := r.URL.Query().Get("on"); q == "0" || strings.EqualFold(q, "off") || strings.EqualFold(q, "false") {
		on = false
	}
	if body.On != nil {
		on = *body.On
	}
	hudSetForce(on)
	if on {
		go func() { _ = hudOpen(hudURL()) }()
	} else if !getLiveState().Active && !demoLiveOn() {
		hudCloseIdle()
	}
	writeJSON(w, map[string]any{"ok": true, "force": hudForceOn(), "dev": true})
}
