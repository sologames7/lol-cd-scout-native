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

func TestHudHitContains(t *testing.T) {
	hits := []hudHit{{X: 10, Y: 20, W: 40, H: 30}, {X: 10, Y: 60, W: 40, H: 20}}
	if !hudHitContains(hits, 12, 22) || hudHitContains(hits, 12, 55) || hudHitContains(hits, 0, 0) {
		t.Fatal("hit-test")
	}
	if len(hudHitsClean([]hudHit{{W: 1, H: 10}, {X: 2, Y: 2, W: 8, H: 8}})) != 1 {
		t.Fatal("clean")
	}
	a := []hudHit{{X: 1, Y: 2, W: 3, H: 4}}
	if !hudHitsEqual(a, []hudHit{{X: 1, Y: 2, W: 3, H: 4}}) || hudHitsEqual(a, nil) {
		t.Fatal("equal")
	}
}
