package main

const (
	hudMiniW = 360
	hudMiniH = 160
	hudMinW  = 200
	hudMinH  = 72
)

type hudHit struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
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
