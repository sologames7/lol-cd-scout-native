package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	verJSON = regexp.MustCompile(`^api/versions\.json$`)
	cdnPath = regexp.MustCompile(`(?i)^[0-9]+(?:\.[0-9]+)+/(?:img/(?:spell|item|champion)/[A-Za-z0-9._-]+|data/[A-Za-z_]+/item\.json)$`)
)

func main() {
	wd, _ := os.Getwd()
	root := wd
	if _, err := os.Stat(filepath.Join(wd, "landing", "index.html")); err == nil {
		root = filepath.Join(wd, "landing")
	} else if _, err := os.Stat(filepath.Join(wd, "index.html")); err != nil {
		fmt.Fprintln(os.Stderr, "pas de landing/index.html ici :", wd)
		os.Exit(1)
	}
	addr := "127.0.0.1:27183"
	fmt.Println("landing → http://" + addr + "  (" + root + ")")

	files := http.FileServer(http.Dir(root))
	mux := http.NewServeMux()
	mux.HandleFunc("/ddragon/", proxyDDragon)
	mux.Handle("/", files)

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 8 * time.Second}
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func proxyDDragon(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/ddragon/")
	var src string
	switch {
	case verJSON.MatchString(rest):
		src = "https://ddragon.leagueoflegends.com/" + rest
	case strings.HasPrefix(rest, "cdn/") && cdnPath.MatchString(strings.TrimPrefix(rest, "cdn/")):
		src = "https://ddragon.leagueoflegends.com/" + rest
	default:
		http.NotFound(w, r)
		return
	}
	resp, err := http.Get(src)
	if err != nil {
		http.Error(w, "ddragon", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	for _, k := range []string{"Content-Type", "Cache-Control"} {
		if v := resp.Header.Get(k); v != "" {
			w.Header().Set(k, v)
		}
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
