package main

const (
	hudMiniW = 200
	hudMiniH = 100
	hudMinW  = 160
	hudMinH  = 72
)

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
