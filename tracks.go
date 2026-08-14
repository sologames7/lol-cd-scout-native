package main

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

// Timers de CD partagés entre le widget permanent et la fenêtre tracker
// (deux WebView2 = deux JS : sans ça, Alt+1 sur l'un n'update pas l'autre).

type clientTrack struct {
	End   float64 `json:"end"`
	Total float64 `json:"total"`
}

var hudTracks = struct {
	mu sync.Mutex
	m  map[string]clientTrack
}{m: map[string]clientTrack{}}

func hudTracksGet() map[string]clientTrack {
	hudTracks.mu.Lock()
	defer hudTracks.mu.Unlock()
	out := make(map[string]clientTrack, len(hudTracks.m))
	for k, v := range hudTracks.m {
		out[k] = v
	}
	return out
}

func hudTracksPut(m map[string]clientTrack) {
	hudTracks.mu.Lock()
	defer hudTracks.mu.Unlock()
	if m == nil {
		hudTracks.m = map[string]clientTrack{}
		return
	}
	hudTracks.m = m
}

func hudTracksClear() {
	hudTracksPut(nil)
}

func apiTracks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, hudTracksGet())
	case http.MethodPost:
		var m map[string]clientTrack
		if err := json.NewDecoder(io.LimitReader(r.Body, 64<<10)).Decode(&m); err != nil && err != io.EOF {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		hudTracksPut(m)
		writeJSON(w, map[string]any{"ok": true})
	default:
		http.Error(w, "GET/POST", http.StatusMethodNotAllowed)
	}
}

// Démo HUD : le front poste l'état fictif pour que la 2e fenêtre le voie aussi.

var demoLive = struct {
	mu    sync.Mutex
	on    bool
	state LiveState
}{}

func setDemoLive(s *LiveState) {
	demoLive.mu.Lock()
	defer demoLive.mu.Unlock()
	if s == nil {
		demoLive.on = false
		demoLive.state = LiveState{}
		return
	}
	st := *s
	st.Active = true
	st.Demo = true
	demoLive.on, demoLive.state = true, st
}

func demoLiveCopy() (LiveState, bool) {
	demoLive.mu.Lock()
	defer demoLive.mu.Unlock()
	if !demoLive.on {
		return LiveState{}, false
	}
	return demoLive.state, true
}

func demoLiveOn() bool {
	demoLive.mu.Lock()
	defer demoLive.mu.Unlock()
	return demoLive.on
}

func apiDemo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		setDemoLive(nil)
		hudTracksClear()
		writeJSON(w, map[string]any{"ok": true})
	case http.MethodPost:
		var s LiveState
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&s); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		setDemoLive(&s)
		writeJSON(w, map[string]any{"ok": true})
	default:
		http.Error(w, "POST/DELETE", http.StatusMethodNotAllowed)
	}
}
