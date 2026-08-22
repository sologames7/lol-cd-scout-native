package main

import "testing"

func TestResolveSummonerFlashHexSmite(t *testing.T) {
	flash := resolveSummoner(rawLiveSummSpell{
		DisplayName:    "Flash",
		RawDisplayName: "GeneratedTip_SummonerSpell_SummonerFlash_DisplayName",
	}, 0)
	if flash.Kind != "flash" || flash.Slug != "SummonerFlash" || flash.CD != 300 {
		t.Fatalf("Flash: %+v", flash)
	}

	hex := resolveSummoner(rawLiveSummSpell{
		DisplayName:    "Flash",
		RawDisplayName: "GeneratedTip_SummonerSpell_SummonerFlash_HextechFlashtraption",
	}, 0)
	if hex.Kind != "hexflash" || hex.Slug != "SummonerHexflash" || hex.CD != 20 {
		t.Fatalf("Hexflash: %+v", hex)
	}

	smite := resolveSummoner(rawLiveSummSpell{
		DisplayName:    "Smite",
		RawDisplayName: "GeneratedTip_SummonerSpell_SummonerSmite_DisplayName",
	}, 0)
	if smite.Kind != "smite" || smite.Slug != "SummonerSmite" || smite.CD != 90 {
		t.Fatalf("Smite: %+v", smite)
	}

	unl := resolveSummoner(rawLiveSummSpell{
		DisplayName:    "Unleashed Smite",
		RawDisplayName: "GeneratedTip_SummonerSpell_SummonerSmiteUnleashed_DisplayName",
	}, 0)
	if unl.Kind != "smite" || unl.Slug != "SummonerSmiteUnleashed" {
		t.Fatalf("Unleashed: %+v", unl)
	}

	chal := resolveSummoner(rawLiveSummSpell{
		DisplayName:    "Challenging Smite",
		RawDisplayName: "GeneratedTip_SummonerSpell_SummonerSmiteChallenging_DisplayName",
	}, 0)
	if chal.Kind != "smite" || chal.Slug != "SummonerSmiteChallenging" {
		t.Fatalf("Challenging: %+v", chal)
	}

	prim := resolveSummoner(rawLiveSummSpell{
		DisplayName:    "Primordial Smite",
		RawDisplayName: "GeneratedTip_SummonerSpell_SummonerSmitePrimal_Primordial",
	}, 0)
	if prim.Kind != "smite" || prim.Slug != "SummonerSmitePrimal" {
		t.Fatalf("Primordial: %+v", prim)
	}
}

func TestNoteDeathAt(t *testing.T) {
	resetDeathClock("test-note-death")
	at := noteDeathAt("zed|Raw", true, 20, 1000, 12, "CLASSIC")
	if at <= 0 {
		t.Fatalf("mort: deathAt=%v", at)
	}
	again := noteDeathAt("zed|Raw", true, 18, 1002, 12, "CLASSIC")
	if again != at {
		t.Fatalf("même mort: %v vs %v", again, at)
	}
	alive := noteDeathAt("zed|Raw", false, 0, 1025, 12, "CLASSIC")
	if alive != 0 {
		t.Fatalf("vivant: deathAt=%v", alive)
	}
}

func TestTranscendenceAH(t *testing.T) {
	if got := transcendenceAH(8); got != 10 {
		t.Fatalf("nv8: %d", got)
	}
	if got := transcendenceAH(4); got != 0 {
		t.Fatalf("nv4: %d", got)
	}
}
