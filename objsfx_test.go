package main

import (
	"strings"
	"testing"
)

func TestWikiaOggURL(t *testing.T) {
	u := wikiaOggURL("Baron_pit_fight_start_SFX.ogg")
	if !strings.Contains(u, "/a/ab/Baron_pit_fight_start_SFX.ogg") {
		t.Errorf("hash wikia inattendu: %s", u)
	}
}

func TestFetchObjSFXDragon(t *testing.T) {
	b, err := fetchObjSFX("dragon")
	if err != nil {
		t.Skipf("wikia indisponible: %v", err)
	}
	if len(b) < 64 || string(b[:4]) != "OggS" {
		t.Fatalf("SFX dragon invalide (%d octets)", len(b))
	}
}

func TestAlertArtEmbed(t *testing.T) {
	for _, name := range []string{"grubs.png", "dragon.png", "baron.png", "herald.png", "infernal.png", "mountain.png", "ocean.png", "cloud.png", "hextech.png", "chemtech.png", "soul.png", "elder.png"} {
		b, err := alertArtFS.ReadFile("branding/alerts/" + name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(b) < 1000 {
			t.Fatalf("%s: trop petit (%d octets)", name, len(b))
		}
		png := len(b) >= 8 && string(b[:8]) == "\x89PNG\r\n\x1a\n"
		jpeg := len(b) >= 3 && b[0] == 0xff && b[1] == 0xd8
		if !png && !jpeg {
			t.Fatalf("%s: pas une image", name)
		}
	}
	if alertArtName.MatchString("foo.png") || alertArtName.MatchString("../grubs.png") {
		t.Fatal("whitelist trop large")
	}
}
