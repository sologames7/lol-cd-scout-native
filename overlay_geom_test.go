package main

import "testing"

func TestClampHudGeom(t *testing.T) {
	x, y, w, h := clampHudGeom(3000, 2000, 200, 100, 1920, 1080)
	if w != 200 || h != 100 {
		t.Fatalf("taille: %dx%d", w, h)
	}
	if x+w > 1920 || y+h > 1080 || x < 0 || y < 0 {
		t.Fatalf("hors écran: %d,%d %dx%d", x, y, w, h)
	}
}

func TestHudDefaultPos(t *testing.T) {
	x, y := hudDefaultPos(200, 100, 1920, 1080)
	if x != 1920-200-16 || y != 16 {
		t.Fatalf("pos=%d,%d", x, y)
	}
}
