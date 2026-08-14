//go:build !windows

package main

import "errors"

const hudWindowTitle = "CD Scout HUD"

type inputEvent struct {
	Seq  uint64 `json:"seq"`
	Kind string `json:"kind"`
	Slot int    `json:"slot"`
}

func setDPIAware() {}

func hudSupported() bool { return false }

func hudOpen(string) error { return errors.New("HUD épinglé réservé à Windows") }

func hudSetPin(bool, bool) {}

func hudStatus() (bool, bool, bool) { return false, false, false }

func hudHold() bool { return false }

func hudSetHold(bool) {}

func hudSetHits([]hudHit) {}

func hudSetBounds(int, int) {}

func hudResetPos() {
	hudGeomReplace(hudGeomDisk{V: 3, Widgets: hudDefaultWidgets(1920, 1080)})
}

func hudBeginDrag() {}

func hudBeginWidgetDrag(hudDragReq) {}

func hudCloseWindow() {}

func autoOpenHudForGame(string) {}

func inputSince(uint64) (bool, uint64, []inputEvent) { return false, 0, []inputEvent{} }

func hudGeomSnapshot() hudGeomDisk {
	return hudGeomDisk{V: 3, Widgets: hudDefaultWidgets(1920, 1080)}
}

func hudGeomReplace(g hudGeomDisk) {}

func hudDragLiveCopy() (hudDragReq, bool) { return hudDragReq{}, false }
