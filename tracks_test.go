package main

import "testing"

func TestHudTracksRoundtrip(t *testing.T) {
	hudTracksClear()
	hudTracksPut(map[string]clientTrack{"a#1|S0": {End: 120, Total: 300}})
	got := hudTracksGet()
	if got["a#1|S0"].Total != 300 {
		t.Fatalf("tracks: %+v", got)
	}
	hudTracksClear()
	if len(hudTracksGet()) != 0 {
		t.Fatal("clear incomplet")
	}
}

func TestHudCIRoundtrip(t *testing.T) {
	hudTracksClear()
	hudCIPut([]string{"Ahri#Ahri", "", "Ahri#Ahri", "Zed#Zed"})
	got := hudCIGet()
	if len(got) != 2 || got[0] != "Ahri#Ahri" || got[1] != "Zed#Zed" {
		t.Fatalf("ci: %#v", got)
	}
	hudTracksClear()
	if len(hudCIGet()) != 0 {
		t.Fatal("clear CI incomplet")
	}
}

func TestDemoLive(t *testing.T) {
	setDemoLive(nil)
	if demoLiveOn() {
		t.Fatal("démo devrait être off")
	}
	setDemoLive(&LiveState{GameTime: 100, Enemies: []LivePlayer{{Champion: "Zed"}}})
	s, ok := demoLiveCopy()
	if !ok || !s.Active || !s.Demo || s.GameTime != 100 {
		t.Fatalf("démo: %+v ok=%v", s, ok)
	}
	setDemoLive(nil)
}
