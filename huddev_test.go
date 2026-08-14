package main

import "testing"

func TestHudDevAllowed(t *testing.T) {
	old := appVersion
	t.Cleanup(func() { appVersion = old })

	t.Setenv("CDSCOUT_DEV", "")
	appVersion = "v1.2.3"
	if hudDevAllowed() {
		t.Fatal("exe release : devmode doit être refusé")
	}
	appVersion = "dev"
	if !hudDevAllowed() {
		t.Fatal("build run.ps1 : devmode doit être autorisé")
	}
	appVersion = "v1.2.3"
	t.Setenv("CDSCOUT_DEV", "1")
	if !hudDevAllowed() {
		t.Fatal("CDSCOUT_DEV=1 : devmode doit être autorisé")
	}
}

func TestHudKeepKey(t *testing.T) {
	setDemoLive(nil)
	if g := hudKeepKey("CLASSIC|Zed"); g != "CLASSIC|Zed" {
		t.Fatalf("live: %q", g)
	}
	hudDev.mu.Lock()
	was := hudDev.force
	hudDev.force = true
	hudDev.mu.Unlock()
	t.Cleanup(func() {
		hudDev.mu.Lock()
		hudDev.force = was
		hudDev.mu.Unlock()
	})
	if g := hudKeepKey(""); g != "dev" {
		t.Fatalf("force hors partie: %q", g)
	}
}
