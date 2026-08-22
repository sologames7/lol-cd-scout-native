package main

import (
	"encoding/json"
	"strings"
)

// Géométrie HUD : HWND plein écran + widgets indépendants.
//
// Deux layouts persistés (v10) : blue = ORDER, red = CHAOS. Le fichier
// hud-geom.json stocke blue + red ; le runtime applique le côté courant.
//
// Layout Tab League — baseline 1080p, scale = sh/1080.
// Scoreboard in-game (S14–S16) : 5 lignes par équipe, tableau centré
// dans la zone jouable 16:9 (sur 21:9 : bandes noires latérales).
// Blue : alliés à gauche, ennemis à droite. Red : miroir horizontal
// dans hudPlayArea (x' = playLeft + playW - (x - playLeft) - widgetW).
//
//	tabRowH1080    = 52  hauteur d'une row champion (icône ~48 px + filet)
//	tabHeaderH1080 = 56  bandeau or / KDA au-dessus des 5 rows
//	tabBoardW1080  = 920 largeur approx. du tableau central à 1080p 16:9
//	tabCardW       = 280 largeur réservée à gauche/droite du tableau (lignes sorts)
//	tabSpellW      = 240 largeur estimée d'une ligne P/Q/W/E/R
//	tabSumW        = 80  largeur estimée d'une paire D/F
//	tabPairGap     = 4   espace entre ligne de sorts et paire d'invocs
const (
	hudGeomVer = 10
	hudMiniW   = 360
	hudMiniH   = 160
	hudMinW    = 200
	hudMinH    = 72
	// Barre Flash optionnelle : colonne de 5 chips (portrait + ult + 2 invocs).
	hudMenuW  = 184
	hudMenuH  = 248
	hudObjsW  = 158
	hudAlertW = 420
	hudItemW  = 260
	hudItemH  = 96
	hudCdsW   = 320
	hudCdsH   = 168
	hudBuffsW = 168
	hudBuffsH = 88

	tabRows        = 5
	tabRowH1080    = 52
	tabHeaderH1080 = 56
	tabBoardW1080  = 920
	tabCardW       = 280
	tabCardGap     = 8
	tabSpellW      = 240
	tabSumW        = 80
	tabPairGap     = 4

	hudScaleMin = 0.55
	hudScaleMax = 2.2
)

var hudWidgetIDs = []string{
	"menu", "objs", "alert", "item", "cds", "buffs",
	"spE0", "spE1", "spE2", "spE3", "spE4",
	"spA0", "spA1", "spA2", "spA3", "spA4",
	"sumE0", "sumE1", "sumE2", "sumE3", "sumE4",
	"sumA0", "sumA1", "sumA2", "sumA3", "sumA4",
}

type hudHit struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type hudWidgetGeom struct {
	X     int     `json:"x"`
	Y     int     `json:"y"`
	Scale float64 `json:"scale,omitempty"`
}

type hudSideLayout struct {
	Keep    bool                     `json:"keep,omitempty"`
	Widgets map[string]hudWidgetGeom `json:"widgets,omitempty"`
}

type hudGeomDisk struct {
	V    int            `json:"v"`
	Side string         `json:"side,omitempty"` // "blue" | "red"
	Blue *hudSideLayout `json:"blue,omitempty"`
	Red  *hudSideLayout `json:"red,omitempty"`
	// Keep / Widgets : côté courant (runtime + repli v≤7 à la racine).
	Keep    bool                     `json:"keep,omitempty"`
	Widgets map[string]hudWidgetGeom `json:"widgets,omitempty"`
	// Ancien format (HWND unique) : ignoré, le canvas est plein écran.
	X int `json:"x,omitempty"`
	Y int `json:"y,omitempty"`
}

type hudDragReq struct {
	ID    string  `json:"id"`
	Mode  string  `json:"mode"`
	X     int     `json:"x"`
	Y     int     `json:"y"`
	Scale float64 `json:"scale"`
}

func (g hudWidgetGeom) scaleOr1() float64 {
	if g.Scale <= 0 {
		return 1
	}
	return g.Scale
}

func clampHudGeom(x, y, w, h, sw, sh int) (int, int, int, int) {
	if sw < 1 {
		sw = 1920
	}
	if sh < 1 {
		sh = 1080
	}
	if w < hudMinW {
		w = hudMinW
	}
	if h < hudMinH {
		h = hudMinH
	}
	if w > sw {
		w = sw
	}
	if h > sh {
		h = sh
	}
	if x+w > sw {
		x = sw - w
	}
	if y+h > sh {
		y = sh - h
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return x, y, w, h
}

func hudDefaultPos(w, h, sw, sh int) (x, y int) {
	x, y, _, _ = clampHudGeom(sw-w-16, 16, w, h, sw, sh)
	return x, y
}

func hudPlayArea(sw, sh int) (x, w int) {
	if sw < 1 {
		sw = 1920
	}
	if sh < 1 {
		sh = 1080
	}
	w = sh * 16 / 9
	if w > sw {
		w = sw
	}
	x = (sw - w) / 2
	return x, w
}

func hudTabMetrics(sw, sh int) (rowH, rowsTop, leftX, rightX int) {
	if sh < 1 {
		sh = 1080
	}
	if sw < 1 {
		sw = 1920
	}
	px, pw := hudPlayArea(sw, sh)
	rowH = tabRowH1080 * sh / 1080
	if rowH < 28 {
		rowH = 28
	}
	headerH := tabHeaderH1080 * sh / 1080
	boardH := headerH + tabRows*rowH
	boardTop := (sh - boardH) / 2
	if boardTop < 0 {
		boardTop = 0
	}
	rowsTop = boardTop + headerH
	boardW := tabBoardW1080 * pw / 1920
	if boardW < 400 {
		boardW = 400
	}
	boardLeft := px + (pw-boardW)/2
	leftX = boardLeft - tabCardW - tabCardGap
	if leftX < 0 {
		leftX = 0
	}
	rightX = boardLeft + boardW + tabCardGap
	if rightX+tabCardW > sw {
		rightX = sw - tabCardW
	}
	if rightX < 0 {
		rightX = 0
	}
	return rowH, rowsTop, leftX, rightX
}

func hudNormSide(s string) string {
	if strings.EqualFold(strings.TrimSpace(s), "red") {
		return "red"
	}
	return "blue"
}

func hudTeamSide(team string) string {
	if strings.EqualFold(strings.TrimSpace(team), "CHAOS") {
		return "red"
	}
	return "blue"
}

func cloneHudWidgets(m map[string]hudWidgetGeom) map[string]hudWidgetGeom {
	if m == nil {
		return nil
	}
	out := make(map[string]hudWidgetGeom, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func hudMirrorX(x, widgetW, sw, sh int) int {
	px, pw := hudPlayArea(sw, sh)
	return px + pw - (x - px) - widgetW
}

func hudMirrorWidgets(src map[string]hudWidgetGeom, sw, sh int) map[string]hudWidgetGeom {
	out := make(map[string]hudWidgetGeom, len(src))
	for id, g := range src {
		ew, _ := hudWidgetEst(id, g.scaleOr1())
		ng := g
		ng.X = hudMirrorX(g.X, ew, sw, sh)
		out[id] = clampHudWidgetID(id, ng, sw, sh)
	}
	return out
}

func hudDefaultWidgets(sw, sh int) map[string]hudWidgetGeom {
	return hudDefaultWidgetsSide("blue", sw, sh)
}

func hudDefaultWidgetsSide(side string, sw, sh int) map[string]hudWidgetGeom {
	blue := hudDefaultWidgetsBlue(sw, sh)
	if hudNormSide(side) == "red" {
		return hudMirrorWidgets(blue, sw, sh)
	}
	return blue
}

func hudDefaultWidgetsBlue(sw, sh int) map[string]hudWidgetGeom {
	if sw < 1 {
		sw = 1920
	}
	if sh < 1 {
		sh = 1080
	}
	pad := 16
	rowH, rowsTop, tabLeft, tabRight := hudTabMetrics(sw, sh)
	itemY := sh - hudItemH - pad
	if itemY < pad {
		itemY = pad
	}
	px, _ := hudPlayArea(sw, sh)
	menuY := (sh - hudMenuH) / 2
	if menuY < pad {
		menuY = pad
	}
	objsX := sw - hudObjsW - pad
	buffsX := objsX - 8 - hudBuffsW
	if buffsX < pad {
		buffsX = pad
	}
	out := map[string]hudWidgetGeom{
		"cds":   {X: sw - hudCdsW - pad, Y: pad, Scale: 1},
		"menu":  {X: px + pad, Y: menuY, Scale: 1},
		"objs":  {X: objsX, Y: pad + 320, Scale: 1},
		"buffs": {X: buffsX, Y: pad + 320, Scale: 1},
		"alert": {X: sw - hudAlertW - pad, Y: sh / 6, Scale: 1},
		"item":  {X: sw - hudItemW - pad, Y: itemY, Scale: 1},
	}
	sumOff := tabSpellW + tabPairGap
	for i := 0; i < tabRows; i++ {
		y := rowsTop + i*rowH
		d := string(rune('0' + i))
		out["spA"+d] = hudWidgetGeom{X: tabLeft, Y: y, Scale: 1}
		out["sumA"+d] = hudWidgetGeom{X: tabLeft + sumOff, Y: y, Scale: 1}
		out["spE"+d] = hudWidgetGeom{X: tabRight, Y: y, Scale: 1}
		out["sumE"+d] = hudWidgetGeom{X: tabRight + sumOff, Y: y, Scale: 1}
	}
	return out
}

func hudWidgetEst(id string, scale float64) (w, h int) {
	if scale <= 0 {
		scale = 1
	}
	switch id {
	case "menu":
		w, h = hudMenuW, hudMenuH
	case "objs":
		w, h = hudObjsW, 180
	case "alert":
		w, h = hudAlertW, 140
	case "item":
		w, h = hudItemW, hudItemH
	case "cds":
		w, h = hudCdsW, hudCdsH
	case "buffs":
		w, h = hudBuffsW, hudBuffsH
	default:
		if len(id) >= 3 && id[:2] == "sp" {
			w, h = tabSpellW, tabRowH1080
		} else if len(id) >= 4 && id[:3] == "sum" {
			w, h = tabSumW, tabRowH1080
		} else {
			w, h = 48, 32
		}
	}
	return int(float64(w)*scale + 0.5), int(float64(h)*scale + 0.5)
}

func clampHudWidget(g hudWidgetGeom, sw, sh int) hudWidgetGeom {
	return clampHudWidgetID("", g, sw, sh)
}

func clampHudWidgetID(id string, g hudWidgetGeom, sw, sh int) hudWidgetGeom {
	sc := g.scaleOr1()
	if sc < hudScaleMin {
		sc = hudScaleMin
	}
	if sc > hudScaleMax {
		sc = hudScaleMax
	}
	g.Scale = sc
	if sw < 1 {
		sw = 1920
	}
	if sh < 1 {
		sh = 1080
	}
	ew, eh := hudWidgetEst(id, sc)
	if g.X+ew > sw {
		g.X = sw - ew
	}
	if g.Y+eh > sh {
		g.Y = sh - eh
	}
	if g.X < 0 {
		g.X = 0
	}
	if g.Y < 0 {
		g.Y = 0
	}
	return g
}

func hudGeomMerge(g hudGeomDisk, sw, sh int) hudGeomDisk {
	side := hudNormSide(g.Side)
	blueDef := hudDefaultWidgetsSide("blue", sw, sh)
	redDef := hudDefaultWidgetsSide("red", sw, sh)
	blueIn, redIn := hudLiftSides(g, blueDef, redDef)
	out := hudGeomDisk{
		V:    hudGeomVer,
		Side: side,
		Blue: hudMergeSide(blueIn, blueDef, g.V, sw, sh),
		Red:  hudMergeSide(redIn, redDef, g.V, sw, sh),
	}
	lay := out.Blue
	if side == "red" {
		lay = out.Red
	}
	out.Keep = lay.Keep
	out.Widgets = lay.Widgets
	return out
}

func hudLiftSides(g hudGeomDisk, blueDef, redDef map[string]hudWidgetGeom) (blue, red *hudSideLayout) {
	if g.Blue != nil && g.Blue.Widgets != nil {
		blue = g.Blue
	} else if g.Widgets != nil && g.V >= 2 {
		blue = &hudSideLayout{Keep: g.Keep, Widgets: cloneHudWidgets(g.Widgets)}
	} else {
		blue = &hudSideLayout{Widgets: cloneHudWidgets(blueDef)}
	}
	if g.Red != nil && g.Red.Widgets != nil {
		red = g.Red
	} else if g.Widgets != nil && g.V >= 2 {
		if g.Keep {
			red = &hudSideLayout{Keep: true, Widgets: cloneHudWidgets(g.Widgets)}
		} else {
			red = &hudSideLayout{Widgets: cloneHudWidgets(redDef)}
		}
	} else {
		red = &hudSideLayout{Widgets: cloneHudWidgets(redDef)}
	}
	return blue, red
}

func hudMergeSide(src *hudSideLayout, def map[string]hudWidgetGeom, v, sw, sh int) *hudSideLayout {
	if src == nil {
		src = &hudSideLayout{}
	}
	out := &hudSideLayout{Keep: src.Keep, Widgets: make(map[string]hudWidgetGeom, len(def))}
	for id, d := range def {
		if cur, ok := src.Widgets[id]; ok {
			out.Widgets[id] = clampHudWidgetID(id, cur, sw, sh)
		} else {
			out.Widgets[id] = d
		}
	}
	// v<5 : anciens défauts trop étroits (menu 420). v5→v6 ajoute le widget CD.
	// v6→v7 : 20 widgets Tab. v7→v8 : layouts blue/red. Keep inchangé.
	// v8→v9 : CD actifs = permanent (haut droite), barre Flash en bas.
	// v9→v10 : barre Flash en colonne (côté gauche blue / droite red).
	// Widget "buffs" (Nash/Ancestral) : pris sur le défaut s'il manque.
	if !src.Keep && v < 5 {
		out.Widgets["menu"] = def["menu"]
		out.Widgets["item"] = def["item"]
		out.Widgets["alert"] = def["alert"]
	}
	if !src.Keep && v < 9 {
		out.Widgets["menu"] = def["menu"]
		out.Widgets["cds"] = def["cds"]
	}
	if !src.Keep && v < 10 {
		out.Widgets["menu"] = def["menu"]
	}
	return out
}

func (g hudGeomDisk) MarshalJSON() ([]byte, error) {
	type disk struct {
		V    int            `json:"v"`
		Side string         `json:"side,omitempty"`
		Blue *hudSideLayout `json:"blue,omitempty"`
		Red  *hudSideLayout `json:"red,omitempty"`
	}
	side := hudNormSide(g.Side)
	blue, red := g.Blue, g.Red
	if blue == nil && g.Widgets != nil {
		blue = &hudSideLayout{Keep: g.Keep, Widgets: g.Widgets}
	}
	if red == nil && g.Widgets != nil && g.Keep {
		red = &hudSideLayout{Keep: true, Widgets: g.Widgets}
	}
	return json.Marshal(disk{V: hudGeomVer, Side: side, Blue: blue, Red: red})
}

func hudActiveLayout(g *hudGeomDisk) *hudSideLayout {
	if g == nil {
		return &hudSideLayout{Widgets: map[string]hudWidgetGeom{}}
	}
	if hudNormSide(g.Side) == "red" {
		if g.Red == nil {
			g.Red = &hudSideLayout{Widgets: map[string]hudWidgetGeom{}}
		}
		if g.Red.Widgets == nil {
			g.Red.Widgets = map[string]hudWidgetGeom{}
		}
		return g.Red
	}
	if g.Blue == nil {
		g.Blue = &hudSideLayout{Widgets: map[string]hudWidgetGeom{}}
	}
	if g.Blue.Widgets == nil {
		g.Blue.Widgets = map[string]hudWidgetGeom{}
	}
	return g.Blue
}

func hudOtherLayout(g *hudGeomDisk) *hudSideLayout {
	if g == nil {
		return &hudSideLayout{Widgets: map[string]hudWidgetGeom{}}
	}
	orig := g.Side
	if hudNormSide(orig) == "red" {
		g.Side = "blue"
	} else {
		g.Side = "red"
	}
	lay := hudActiveLayout(g)
	g.Side = orig
	return lay
}

func hudMarkSideSaved(g *hudGeomDisk) {
	if g == nil {
		return
	}
	lay := hudActiveLayout(g)
	lay.Keep = true
	g.Keep = true
}

// Un côté gardé (après un drag) sert de source : l'autre, s'il n'a
// jamais été bougé, reprend le miroir. Comme ça blue et red survivent
// aux parties sans recliquer « Sauver dispo ».
func hudSyncKeptSides(g *hudGeomDisk, sw, sh int) {
	if g == nil {
		return
	}
	cur := hudActiveLayout(g)
	other := hudOtherLayout(g)
	if cur.Keep && !other.Keep {
		other.Keep = true
		other.Widgets = hudMirrorWidgets(cur.Widgets, sw, sh)
	} else if !cur.Keep && other.Keep {
		cur.Keep = true
		cur.Widgets = hudMirrorWidgets(other.Widgets, sw, sh)
	}
	g.Keep = cur.Keep
	g.Widgets = cur.Widgets
}

func hudHitContains(hits []hudHit, x, y int) bool {
	for _, h := range hits {
		if h.W < 1 || h.H < 1 {
			continue
		}
		if x >= h.X && y >= h.Y && x < h.X+h.W && y < h.Y+h.H {
			return true
		}
	}
	return false
}

func hudHitsClean(rs []hudHit) []hudHit {
	out := make([]hudHit, 0, len(rs))
	for _, h := range rs {
		if h.W < 2 || h.H < 2 {
			continue
		}
		out = append(out, h)
	}
	return out
}

func hudHitsEqual(a, b []hudHit) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func validHudWidgetID(id string) bool {
	for _, w := range hudWidgetIDs {
		if w == id {
			return true
		}
	}
	return false
}
