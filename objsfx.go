package main

import (
	"bytes"
	"crypto/md5"
	"embed"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"sync"
)

//go:embed branding/alerts/*.png
var alertArtFS embed.FS

var alertArtName = regexp.MustCompile(`^(grubs|dragon|baron|herald|infernal|mountain|ocean|cloud|hextech|chemtech|soul|elder)\.png$`)

// SFX courts des objectifs (wiki LoL, extraits client). Proxifiés same-origin
// pour WebView2. Le front coupe à ~2 s : les thèmes de fight larves/héraut
// font 20–30 s, on n'en joue que l'attaque.

var objSFXFile = map[string]string{
	"dragon": "Infernal_Soul_trigger_SFX.ogg",          // sting drake (âme infernale)
	"elder":  "Elder_Dragon_execute_SFX.ogg",           // ancestral / âme
	"baron":  "Baron_pit_fight_start_SFX.ogg",          // rugissement d'aggro Nashor
	"grubs":  "Voidgrub_fight_background_music.ogg",    // thème de fight des larves
	"herald": "Rift_Herald_fight_background_music.ogg", // thème de fight du héraut
}

var objSFXCache = struct {
	mu   sync.Mutex
	data map[string][]byte
}{data: map[string][]byte{}}

func wikiaOggURL(filename string) string {
	sum := md5.Sum([]byte(filename))
	h := hex.EncodeToString(sum[:])
	return fmt.Sprintf("https://static.wikia.nocookie.net/leagueoflegends/images/%c/%s/%s", h[0], h[:2], filename)
}

func fetchObjSFX(key string) ([]byte, error) {
	file, ok := objSFXFile[key]
	if !ok {
		return nil, fmt.Errorf("inconnu")
	}
	objSFXCache.mu.Lock()
	if b, hit := objSFXCache.data[key]; hit {
		objSFXCache.mu.Unlock()
		return b, nil
	}
	objSFXCache.mu.Unlock()

	b, err := fetchVoiceOGG(wikiaOggURL(file))
	if err != nil {
		return nil, err
	}
	objSFXCache.mu.Lock()
	objSFXCache.data[key] = b
	objSFXCache.mu.Unlock()
	return b, nil
}

func apiObjSFX(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "GET", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("k")
	b, err := fetchObjSFX(key)
	if err != nil {
		http.Error(w, "sfx introuvable", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "audio/ogg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = io.Copy(w, bytes.NewReader(b))
}

func apiAlertArt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "GET", http.StatusMethodNotAllowed)
		return
	}
	name := path.Base(r.URL.Path)
	if !alertArtName.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	b, err := alertArtFS.ReadFile("branding/alerts/" + name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	ct := "image/png"
	if len(b) >= 3 && b[0] == 0xff && b[1] == 0xd8 {
		ct = "image/jpeg"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = io.Copy(w, bytes.NewReader(b))
}

func prefetchObjSFX() {
	objSFXOnce.Do(func() {
		for k := range objSFXFile {
			go func(key string) { _, _ = fetchObjSFX(key) }(k)
		}
	})
}

var objSFXOnce sync.Once
