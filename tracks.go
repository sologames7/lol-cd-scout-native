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
	Auto  bool    `json:"auto,omitempty"`
}

type tracksPayload struct {
	Tracks map[string]clientTrack `json:"tracks,omitempty"`
	CI     []string               `json:"ci,omitempty"`
}

var hudTracks = struct {
	mu sync.Mutex
	m  map[string]clientTrack
}{m: map[string]clientTrack{}}

var hudCI = struct {
	mu  sync.Mutex
	ids []string
}{}

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
	hudCIPut(nil)
}

func hudCIGet() []string {
	hudCI.mu.Lock()
	defer hudCI.mu.Unlock()
	out := make([]string, len(hudCI.ids))
	copy(out, hudCI.ids)
	return out
}

func hudCIPut(ids []string) {
	hudCI.mu.Lock()
	defer hudCI.mu.Unlock()
	if len(ids) == 0 {
		hudCI.ids = nil
		return
	}
	out := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	hudCI.ids = out
}

func apiTracks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, tracksPayload{Tracks: hudTracksGet(), CI: hudCIGet()})
	case http.MethodPost:
		raw, err := io.ReadAll(io.LimitReader(r.Body, 64<<10))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(raw) > 0 {
			var wrap tracksPayload
			if json.Unmarshal(raw, &wrap) == nil && (wrap.Tracks != nil || wrap.CI != nil) {
				if wrap.Tracks != nil {
					hudTracksPut(wrap.Tracks)
				}
				if wrap.CI != nil {
					hudCIPut(wrap.CI)
				}
			} else {
				var m map[string]clientTrack
				if err := json.Unmarshal(raw, &m); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				hudTracksPut(m)
			}
		}
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
