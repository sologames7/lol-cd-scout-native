package main

import (
	"encoding/json"
	"testing"
)

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
	wantX := 1920 - hudCdsW - 16
	if w["cds"].X != wantX || w["cds"].Y != 16 {
		t.Fatalf("cds permanent=%+v want x=%d", w["cds"], wantX)
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
	w := hudDefaultWidgets(1920, 1080)
	for _, id := range []string{"spA0", "sumA0", "spA4", "sumA4", "spE0", "sumE4"} {
		g, ok := w[id]
		if !ok {
			t.Fatalf("widget manquant: %s", id)
		}
		ew, eh := hudWidgetEst(id, 1)
		if g.X < 0 || g.Y < 0 || g.X+ew > 1920 || g.Y+eh > 1080 {
			t.Fatalf("%s dépasse: %+v est=%dx%d", id, g, ew, eh)
		}
	}
	rowH, rowsTop, leftX, rightX := hudTabMetrics(1920, 1080)
	if w["spA1"].Y != rowsTop+rowH || w["spA1"].X != leftX {
		t.Fatalf("spA1 pas sur la row 1 gauche: %+v top=%d rowH=%d L=%d", w["spA1"], rowsTop, rowH, leftX)
	}
	if w["sumA0"].X != leftX+tabSpellW+tabPairGap {
		t.Fatalf("sumA0 pas collé à spA0: %+v", w["sumA0"])
	}
	if w["spE0"].X != rightX {
		t.Fatalf("spE0 pas à droite: %+v R=%d", w["spE0"], rightX)
	}
}

func TestHudBuffsDefaultBesideObjs(t *testing.T) {
	w := hudDefaultWidgets(1920, 1080)
	b, ok := w["buffs"]
	if !ok {
		t.Fatal("widget buffs manquant")
	}
	o := w["objs"]
	if b.Y != o.Y {
		t.Fatalf("buffs Y=%d objs Y=%d", b.Y, o.Y)
	}
	bw, _ := hudWidgetEst("buffs", 1)
	if b.X+bw+8 > o.X {
		t.Fatalf("buffs doit être à gauche des objs: buffs=%+v w=%d objs=%+v", b, bw, o)
	}
}

func TestHudGeomMergeV7AddsTabRows(t *testing.T) {
	old := hudGeomDisk{
		V:    6,
		Keep: true,
		Widgets: map[string]hudWidgetGeom{
			"menu": {X: 100, Y: 40, Scale: 1.1},
			"tabE": {X: 50, Y: 80, Scale: 1},
			"tabA": {X: 400, Y: 80, Scale: 1},
		},
	}
	m := hudGeomMerge(old, 1920, 1080)
	if !m.Keep || m.V != hudGeomVer {
		t.Fatalf("keep perdu: %+v", m)
	}
	if m.Widgets["menu"].X != 100 || m.Widgets["menu"].Scale != 1.1 {
		t.Fatalf("layout sauvé recalé: %+v", m.Widgets["menu"])
	}
	def := hudDefaultWidgets(1920, 1080)
	if _, ok := m.Widgets["spE0"]; !ok {
		t.Fatal("spE0 défaut manquant")
	}
	if m.Widgets["spE0"] != def["spE0"] {
		t.Fatalf("spE0 pas au défaut: %+v want %+v", m.Widgets["spE0"], def["spE0"])
	}
	if _, ok := m.Widgets["sumA4"]; !ok {
		t.Fatal("sumA4 défaut manquant")
	}
	if _, ok := m.Widgets["tabE"]; ok {
		t.Fatal("tabE ne doit plus être un widget")
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
	// Ancien défaut trop bas : la colonne Flash doit remonter.
	g := clampHudWidgetID("menu", hudWidgetGeom{X: 16, Y: 900, Scale: 1}, 1920, 1080)
	_, eh := hudWidgetEst("menu", 1)
	if g.Y+eh > 1080 || g.Y < 0 {
		t.Fatalf("menu encore hors écran: y=%d estH=%d", g.Y, eh)
	}
	if g.Y >= 900 {
		t.Fatalf("doit remonter le menu: y=%d", g.Y)
	}
}

func TestHudTeamSide(t *testing.T) {
	if hudTeamSide("ORDER") != "blue" || hudTeamSide("CHAOS") != "red" || hudTeamSide("") != "blue" {
		t.Fatal("hudTeamSide")
	}
	if hudNormSide("RED") != "red" || hudNormSide("blue") != "blue" || hudNormSide("") != "blue" {
		t.Fatal("hudNormSide")
	}
}

func TestHudGeomMergeV7ToV8KeepCopiesBoth(t *testing.T) {
	old := hudGeomDisk{
		V:    7,
		Keep: true,
		Widgets: map[string]hudWidgetGeom{
			"menu": {X: 100, Y: 40, Scale: 1.1},
		},
	}
	m := hudGeomMerge(old, 1920, 1080)
	if m.V != hudGeomVer || m.Side != "blue" {
		t.Fatalf("v%d: v=%d side=%s", hudGeomVer, m.V, m.Side)
	}
	if m.Blue == nil || m.Red == nil || !m.Blue.Keep || !m.Red.Keep {
		t.Fatal("keep des deux côtés")
	}
	if m.Blue.Widgets["menu"].X != 100 || m.Red.Widgets["menu"].X != 100 {
		t.Fatalf("copie v7: blue=%+v red=%+v", m.Blue.Widgets["menu"], m.Red.Widgets["menu"])
	}
	if _, ok := m.Blue.Widgets["spA0"]; !ok {
		t.Fatal("spA0 manquant après merge v7")
	}
}

func TestHudGeomMergeV7NoKeepRedIsMirror(t *testing.T) {
	old := hudGeomDisk{V: 7, Widgets: map[string]hudWidgetGeom{"menu": {X: 100, Y: 40, Scale: 1}}}
	m := hudGeomMerge(old, 1920, 1080)
	if m.Red == nil || m.Red.Keep {
		t.Fatal("red ne doit pas hériter keep")
	}
	defB := hudDefaultWidgetsSide("blue", 1920, 1080)
	defR := hudDefaultWidgetsSide("red", 1920, 1080)
	if m.Red.Widgets["menu"] != defR["menu"] {
		t.Fatalf("red doit être le miroir: %+v want %+v", m.Red.Widgets["menu"], defR["menu"])
	}
	if m.Widgets["menu"] != defB["menu"] || m.Widgets["cds"] != defB["cds"] {
		t.Fatalf("v7 sans keep: menu/cds recalés v9 menu=%+v cds=%+v", m.Widgets["menu"], m.Widgets["cds"])
	}
}

func TestHudMirrorWidgetsBlueRed(t *testing.T) {
	blue := hudDefaultWidgetsSide("blue", 1920, 1080)
	red := hudDefaultWidgetsSide("red", 1920, 1080)
	ew, _ := hudWidgetEst("cds", 1)
	want := hudMirrorX(blue["cds"].X, ew, 1920, 1080)
	if red["cds"].X != want {
		t.Fatalf("cds red x=%d want %d (blue x=%d w=%d)", red["cds"].X, want, blue["cds"].X, ew)
	}
	if red["cds"].X >= blue["cds"].X {
		t.Fatal("red cds doit être à gauche")
	}
	if blue["spA0"].X >= blue["spE0"].X {
		t.Fatal("blue: alliés à gauche")
	}
	if red["spA0"].X <= red["spE0"].X {
		t.Fatal("red: alliés à droite")
	}
}

func TestHudPlayAreaMirror21_9(t *testing.T) {
	blue := hudDefaultWidgetsSide("blue", 2560, 1080)
	red := hudDefaultWidgetsSide("red", 2560, 1080)
	ew, _ := hudWidgetEst("cds", 1)
	want := clampHudWidgetID("cds", hudWidgetGeom{X: hudMirrorX(blue["cds"].X, ew, 2560, 1080), Y: blue["cds"].Y, Scale: 1}, 2560, 1080)
	if red["cds"].X != want.X {
		t.Fatalf("21:9 cds red=%d want %d", red["cds"].X, want.X)
	}
}

func TestHudGeomMergeV8ToV9SwapsPermanent(t *testing.T) {
	oldMenu := hudWidgetGeom{X: 1384, Y: 16, Scale: 1}
	oldCds := hudWidgetGeom{X: 800, Y: 856, Scale: 1}
	old := hudGeomDisk{
		V: 8,
		Widgets: map[string]hudWidgetGeom{
			"menu": oldMenu,
			"cds":  oldCds,
		},
	}
	m := hudGeomMerge(old, 1920, 1080)
	def := hudDefaultWidgets(1920, 1080)
	if m.Widgets["menu"] != def["menu"] || m.Widgets["cds"] != def["cds"] {
		t.Fatalf("v8 sans keep doit échanger: menu=%+v cds=%+v", m.Widgets["menu"], m.Widgets["cds"])
	}
	kept := hudGeomMerge(hudGeomDisk{
		V: 8, Keep: true,
		Widgets: map[string]hudWidgetGeom{"menu": oldMenu, "cds": oldCds},
	}, 1920, 1080)
	if kept.Widgets["menu"] != oldMenu || kept.Widgets["cds"] != oldCds {
		t.Fatalf("keep doit conserver: menu=%+v cds=%+v", kept.Widgets["menu"], kept.Widgets["cds"])
	}
}

func TestHudGeomJSONRoundtripV8(t *testing.T) {
	src := hudGeomMerge(hudGeomDisk{V: 8, Side: "red"}, 1920, 1080)
	b, err := json.Marshal(src)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if json.Unmarshal(b, &raw) != nil {
		t.Fatal("json")
	}
	if raw["v"].(float64) != float64(hudGeomVer) || raw["side"] != "red" {
		t.Fatalf("disk: %s", b)
	}
	if _, ok := raw["blue"].(map[string]any); !ok {
		t.Fatal("blue manquant")
	}
	if _, ok := raw["red"].(map[string]any); !ok {
		t.Fatal("red manquant")
	}
	var back hudGeomDisk
	if json.Unmarshal(b, &back) != nil {
		t.Fatal("unmarshal")
	}
	m := hudGeomMerge(back, 1920, 1080)
	if m.Side != "red" || m.Widgets["menu"].X != src.Red.Widgets["menu"].X {
		t.Fatalf("roundtrip side=%s menu=%+v", m.Side, m.Widgets["menu"])
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

func TestHudSyncKeptSidesMirrorsOther(t *testing.T) {
	custom := hudWidgetGeom{X: 40, Y: 80, Scale: 1.2}
	g := hudGeomMerge(hudGeomDisk{
		V:    hudGeomVer,
		Side: "blue",
		Blue: &hudSideLayout{Keep: true, Widgets: map[string]hudWidgetGeom{"menu": custom}},
	}, 1920, 1080)
	if g.Red.Keep {
		t.Fatal("red ne doit pas être keep avant sync")
	}
	hudSyncKeptSides(&g, 1920, 1080)
	if !g.Blue.Keep || !g.Red.Keep {
		t.Fatal("les deux côtés doivent être gardés")
	}
	want := hudMirrorWidgets(g.Blue.Widgets, 1920, 1080)
	if g.Red.Widgets["menu"] != want["menu"] {
		t.Fatalf("red menu=%+v want %+v", g.Red.Widgets["menu"], want["menu"])
	}
}

func TestHudSyncKeptSidesDoesNotOverwriteKept(t *testing.T) {
	blueMenu := hudWidgetGeom{X: 40, Y: 80, Scale: 1}
	redMenu := hudWidgetGeom{X: 900, Y: 200, Scale: 1.1}
	g := hudGeomMerge(hudGeomDisk{
		V:    hudGeomVer,
		Side: "blue",
		Blue: &hudSideLayout{Keep: true, Widgets: map[string]hudWidgetGeom{"menu": blueMenu}},
		Red:  &hudSideLayout{Keep: true, Widgets: map[string]hudWidgetGeom{"menu": redMenu}},
	}, 1920, 1080)
	hudSyncKeptSides(&g, 1920, 1080)
	if g.Blue.Widgets["menu"] != blueMenu || g.Red.Widgets["menu"] != redMenu {
		t.Fatalf("overwrite: blue=%+v red=%+v", g.Blue.Widgets["menu"], g.Red.Widgets["menu"])
	}
}

func TestHudSyncKeptSidesFillsCurrentFromOther(t *testing.T) {
	custom := hudWidgetGeom{X: 40, Y: 80, Scale: 1.2}
	g := hudGeomMerge(hudGeomDisk{
		V:    hudGeomVer,
		Side: "red",
		Blue: &hudSideLayout{Keep: true, Widgets: map[string]hudWidgetGeom{"menu": custom}},
	}, 1920, 1080)
	if g.Red.Keep {
		t.Fatal("red défaut ne doit pas être keep")
	}
	hudSyncKeptSides(&g, 1920, 1080)
	if !g.Red.Keep {
		t.Fatal("red doit hériter du keep blue")
	}
	want := hudMirrorWidgets(g.Blue.Widgets, 1920, 1080)
	if g.Red.Widgets["menu"] != want["menu"] {
		t.Fatalf("red menu=%+v want %+v", g.Red.Widgets["menu"], want["menu"])
	}
}

func TestHudSyncKeptSidesNoKeepIsNoop(t *testing.T) {
	g := hudGeomMerge(hudGeomDisk{V: hudGeomVer, Side: "blue"}, 1920, 1080)
	defB := g.Blue.Widgets["menu"]
	defR := g.Red.Widgets["menu"]
	hudSyncKeptSides(&g, 1920, 1080)
	if g.Blue.Keep || g.Red.Keep {
		t.Fatal("pas de keep spontané")
	}
	if g.Blue.Widgets["menu"] != defB || g.Red.Widgets["menu"] != defR {
		t.Fatal("défauts bougés sans drag")
	}
}
