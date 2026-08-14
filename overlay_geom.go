package main

// Géométrie HUD : HWND plein écran + widgets indépendants.
//
// Layout Tab League — baseline 1080p, scale = sh/1080.
// Scoreboard in-game (S14–S16) : 5 lignes par équipe, tableau centré
// dans la zone jouable 16:9 (sur 21:9 : bandes noires latérales).
//
//	tabRowH1080    = 52  hauteur d'une row champion (icône ~48 px + filet)
//	tabHeaderH1080 = 56  bandeau or / KDA au-dessus des 5 rows
//	tabBoardW1080  = 920 largeur approx. du tableau central à 1080p 16:9
//	tabCardW       = 280 largeur de nos cartes latérales (icônes agrandies)
//	tabCardGap     = 8   espace entre cartes et bord du tableau
const (
	hudMiniW = 360
	hudMiniH = 160
	hudMinW  = 200
	hudMinH  = 72

	tabRows        = 5
	tabRowH1080    = 52
	tabHeaderH1080 = 56
	tabBoardW1080  = 920
	tabCardW       = 280
	tabCardGap     = 8

	hudScaleMin = 0.55
	hudScaleMax = 2.2
)

var hudWidgetIDs = []string{"menu", "objs", "alert", "item", "tabE", "tabA"}

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

type hudGeomDisk struct {
	V       int                      `json:"v"`
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
	if rightX+80 > sw {
		rightX = sw - 80
	}
	return rowH, rowsTop, leftX, rightX
}

func hudDefaultWidgets(sw, sh int) map[string]hudWidgetGeom {
	if sw < 1 {
		sw = 1920
	}
	if sh < 1 {
		sh = 1080
	}
	pad := 16
	menuW := 240
	objsW := 158
	_, rowsTop, tabLeft, tabRight := hudTabMetrics(sw, sh)
	return map[string]hudWidgetGeom{
		"menu":  {X: sw - menuW - pad, Y: pad, Scale: 1},
		"objs":  {X: sw - objsW - pad, Y: pad + 320, Scale: 1},
		"alert": {X: sw - 400 - pad, Y: sh / 6, Scale: 1},
		"item":  {X: sw - 240 - pad, Y: sh - 108, Scale: 1},
		"tabE":  {X: tabLeft, Y: rowsTop, Scale: 1},
		"tabA":  {X: tabRight, Y: rowsTop, Scale: 1},
	}
}

func clampHudWidget(g hudWidgetGeom, sw, sh int) hudWidgetGeom {
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
	if g.X > sw-48 {
		g.X = sw - 48
	}
	if g.Y > sh-32 {
		g.Y = sh - 32
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
	def := hudDefaultWidgets(sw, sh)
	if g.V < 2 || g.Widgets == nil {
		return hudGeomDisk{V: 4, Widgets: def}
	}
	out := hudGeomDisk{V: 4, Widgets: make(map[string]hudWidgetGeom, len(def))}
	for id, d := range def {
		if cur, ok := g.Widgets[id]; ok {
			out.Widgets[id] = clampHudWidget(cur, sw, sh)
		} else {
			out.Widgets[id] = d
		}
	}
	if g.V < 4 {
		out.Widgets["menu"] = def["menu"]
	}
	return out
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
