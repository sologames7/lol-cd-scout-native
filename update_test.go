package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v1.0.1", "v1.0.0", true},
		{"v1.1.0", "v1.0.9", true},
		{"v2.0.0", "v1.9.9", true},
		{"v1.0.0", "v1.0.0", false},
		{"v1.0.0", "v1.0.1", false},
		{"v1.0", "v1.0.0", false}, // composants manquants = 0
		{"v1.0.0", "v1.0", false},
		{"v1.0.1", "v1.0", true},
		{"1.2.3", "v1.2.2", true},
		{"v1.2.3-beta", "v1.2.3", false},
		{"pas-une-version", "v1.0.0", false},
		{"v1.0.1", "dev", true}, // garde-fou : checkAndStage exclut déjà "dev"
	}
	for _, c := range cases {
		if got := isNewer(c.latest, c.current); got != c.want {
			t.Errorf("isNewer(%q, %q) = %v, attendu %v", c.latest, c.current, got, c.want)
		}
	}
}

// fakeExe : entête PE minimale + rembourrage au-delà du seuil de 1 Mo.
func fakeExe() []byte { return append([]byte("MZ"), bytes.Repeat([]byte{0x90}, 1<<20)...) }

func TestDownloadTo(t *testing.T) {
	payload := fakeExe()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(payload)
	}))
	defer srv.Close()

	dir := t.TempDir()
	dst := filepath.Join(dir, "app.exe.new")
	if err := downloadTo(srv.URL, dst, int64(len(payload))); err != nil {
		t.Fatalf("downloadTo : %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatalf("contenu téléchargé incorrect (err=%v)", err)
	}
	if !looksLikeExe(dst) {
		t.Error("looksLikeExe devrait accepter le fichier téléchargé")
	}
	// Aucun temporaire ne doit subsister à côté.
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("fichiers résiduels dans le dossier : %d", len(entries))
	}
}

func TestDownloadToRejectsWrongSize(t *testing.T) {
	payload := fakeExe()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(payload)
	}))
	defer srv.Close()

	dir := t.TempDir()
	dst := filepath.Join(dir, "app.exe.new")
	if err := downloadTo(srv.URL, dst, int64(len(payload))+42); err == nil {
		t.Fatal("une taille incohérente doit être rejetée")
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Error("aucun fichier ne doit être laissé après un échec")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("temporaires non nettoyés : %d", len(entries))
	}
}

func TestLooksLikeExeRejectsGarbage(t *testing.T) {
	dir := t.TempDir()
	small := filepath.Join(dir, "small.exe")
	os.WriteFile(small, []byte("MZ court"), 0o644)
	if looksLikeExe(small) {
		t.Error("un fichier trop petit doit être rejeté")
	}
	html := filepath.Join(dir, "page.exe")
	os.WriteFile(html, append([]byte("<!doctype html>"), bytes.Repeat([]byte{0}, 1<<20)...), 0o644)
	if looksLikeExe(html) {
		t.Error("un fichier sans entête MZ doit être rejeté")
	}
}
