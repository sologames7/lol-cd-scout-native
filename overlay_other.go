//go:build !windows

package main

import "errors"

const hudWindowTitle = "CD Scout HUD"

type inputEvent struct {
	Seq  uint64 `json:"seq"`
	Kind string `json:"kind"`
	Slot int    `json:"slot"`
}

func hudSupported() bool { return false }

func hudOpen(string) error { return errors.New("HUD épinglé réservé à Windows") }

func hudSetPin(bool, bool) {}

func hudStatus() (bool, bool, bool) { return false, false, false }

func hudHold() bool { return false }

func hudSetHold(bool) {}

type hudHit struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

func hudSetHits([]hudHit) {}

func hudSetSolid(bool) {}

func hudSetBounds(int, int) {}

func hudResetPos() {}

func hudBeginDrag() {}

func hudCloseWindow() {}

func autoOpenHudForGame(string) {}

func inputSince(uint64) (bool, uint64, []inputEvent) { return false, 0, []inputEvent{} }
