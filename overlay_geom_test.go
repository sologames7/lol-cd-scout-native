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

func TestHudTabMetrics(t *testing.T) {
	rowH, rowsTop, leftX, rightX := hudTabMetrics(1920, 1080)
	if rowH != tabRowH1080 {
		t.Fatalf("rowH=%d want %d", rowH, tabRowH1080)
	}
	if leftX >= rightX || rowsTop < 40 || rowsTop > 500 {
		t.Fatalf("tab layout: row=%d top=%d L=%d R=%d", rowH, rowsTop, leftX, rightX)
	}
	_, _, l21, r21 := hudTabMetrics(2560, 1080)
	if l21 <= leftX {
		t.Fatalf("21:9: les cartes doivent suivre la zone 16:9 centrée (L %d vs %d)", l21, leftX)
	}
	if r21 <= rightX {
		t.Fatalf("21:9 right %d vs 16:9 %d", r21, rightX)
	}
}

func TestHudDefaultWidgets(t *testing.T) {
	w := hudDefaultWidgets(1920, 1080)
	for _, id := range hudWidgetIDs {
		if _, ok := w[id]; !ok {
			t.Fatalf("widget manquant: %s", id)
		}
	}
	wantX := 1920 - hudMenuW - 16
	if w["menu"].X != wantX || w["menu"].Y != 16 {
		t.Fatalf("menu=%+v want x=%d", w["menu"], wantX)
	}
	for id, g := range w {
		ew, eh := hudWidgetEst(id, g.scaleOr1())
		if g.X < 0 || g.Y < 0 || g.X+ew > 1920 || g.Y+eh > 1080 {
			t.Fatalf("%s hors écran: %+v est=%dx%d", id, g, ew, eh)
		}
	}
}

func TestHudGeomMergeLegacy(t *testing.T) {
	g := hudGeomMerge(hudGeomDisk{X: 10, Y: 20}, 1920, 1080)
	if g.V != hudGeomVer || g.Widgets["menu"].X == 0 && g.Widgets["menu"].Y == 0 {
		t.Fatalf("merge legacy: %+v", g)
	}
	old := hudGeomDisk{V: 2, Widgets: map[string]hudWidgetGeom{"menu": {X: 100, Y: 40, Scale: 1.2}}}
	up := hudGeomMerge(old, 1920, 1080)
	if up.Widgets["menu"].X == 100 {
		t.Fatalf("v2 doit recentrer le menu horizontal: %+v", up.Widgets["menu"])
	}
	v3 := hudGeomDisk{V: 3, Widgets: map[string]hudWidgetGeom{"menu": {X: 100, Y: 40, Scale: 1}}}
	if hudGeomMerge(v3, 1920, 1080).Widgets["menu"].X == 100 {
		t.Fatal("v3 doit recaler le menu à droite")
	}
	v4 := hudGeomDisk{V: 4, Widgets: map[string]hudWidgetGeom{"menu": {X: 1664, Y: 16, Scale: 1}}}
	if hudGeomMerge(v4, 1920, 1080).Widgets["menu"].X == 1664 {
		t.Fatal("v4 menu trop étroit doit être recalculé")
	}
	cur := hudGeomDisk{V: hudGeomVer, Widgets: map[string]hudWidgetGeom{"menu": {X: 100, Y: 40, Scale: 1.2}}}
	m := hudGeomMerge(cur, 1920, 1080)
	if m.Widgets["menu"].X != 100 || m.Widgets["menu"].Scale != 1.2 {
		t.Fatalf("keep menu v5: %+v", m.Widgets["menu"])
	}
	if _, ok := m.Widgets["objs"]; !ok {
		t.Fatal("objs défaut")
	}
}

func TestHudGeomMergeSavedLayout(t *testing.T) {
	old := hudGeomDisk{
		V:    4,
		Keep: true,
		Widgets: map[string]hudWidgetGeom{
			"menu": {X: 100, Y: 40, Scale: 1.1},
			"item": {X: 80, Y: 900, Scale: 1},
		},
	}
	m := hudGeomMerge(old, 1920, 1080)
	if !m.Keep || m.V != hudGeomVer {
		t.Fatalf("keep perdu: %+v", m)
	}
	if m.Widgets["menu"].X != 100 || m.Widgets["menu"].Scale != 1.1 {
		t.Fatalf("layout sauvé recalé: %+v", m.Widgets["menu"])
	}
	if m.Widgets["item"].X != 80 {
		t.Fatalf("item sauvé recalé: %+v", m.Widgets["item"])
	}
	if _, ok := m.Widgets["objs"]; !ok {
		t.Fatal("objs défaut manquant sur un layout partiel")
	}
}

func TestClampHudWidget(t *testing.T) {
	g := clampHudWidgetID("menu", hudWidgetGeom{X: 9000, Y: -8, Scale: 9}, 1920, 1080)
	ew, _ := hudWidgetEst("menu", hudScaleMax)
	if g.X+ew > 1920 || g.Y < 0 || g.Scale > hudScaleMax {
		t.Fatalf("%+v estW=%d", g, ew)
	}
}

func TestHudTabAlliesOnScreen(t *testing.T) {
	_, _, _, rightX := hudTabMetrics(1920, 1080)
	if rightX+tabCardW > 1920 {
		t.Fatalf("tabA dépasse: right=%d + %d", rightX, tabCardW)
	}
}

func TestHudMenuDefaultFitsScreens(t *testing.T) {
	for _, scr := range [][2]int{{1920, 1080}, {2560, 1080}, {1366, 768}, {1280, 720}} {
		sw, sh := scr[0], scr[1]
		w := hudDefaultWidgets(sw, sh)
		for id, g := range w {
			ew, eh := hudWidgetEst(id, g.scaleOr1())
			if g.X < 0 || g.Y < 0 || g.X+ew > sw || g.Y+eh > sh {
				t.Fatalf("%s hors %dx%d: %+v est=%dx%d", id, sw, sh, g, ew, eh)
			}
		}
	}
}

func TestClampOldMenu420Default(t *testing.T) {
	// Défaut v5 : sw-420-16. Les 5 chips Flash font plus large → clamp à gauche.
	oldX := 1920 - 420 - 16
	g := clampHudWidgetID("menu", hudWidgetGeom{X: oldX, Y: 16, Scale: 1}, 1920, 1080)
	ew, _ := hudWidgetEst("menu", 1)
	if g.X+ew > 1920 || g.X < 0 {
		t.Fatalf("menu encore hors écran: x=%d estW=%d", g.X, ew)
	}
	if g.X >= oldX {
		t.Fatalf("doit reculer le menu: x=%d old=%d", g.X, oldX)
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
	// Hits JS serrés : deux blocs du menu, pas un bbox widget unique.
	menu := []hudHit{{X: 1500, Y: 16, W: 280, H: 32}, {X: 1500, Y: 52, W: 220, H: 90}}
	if hudHitContains(menu, 1490, 40) || !hudHitContains(menu, 1510, 20) || !hudHitContains(menu, 1520, 80) {
		t.Fatal("hits menu serrés")
	}
	if hudHitsEqual(menu, []hudHit{{X: 1500, Y: 16, W: 280, H: 126}}) {
		t.Fatal("bbox widget ≠ union des blocs")
	}
}
