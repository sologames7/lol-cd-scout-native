package main

import "testing"

func TestParsePassiveCD(t *testing.T) {
	zac := "Cette compétence a un délai de récupération de 5 minutes."
	got := parsePassiveCD(zac)
	if len(got) != 1 || got[0] != 300 {
		t.Fatalf("Zac FR: %v", got)
	}
	en := "This ability has a 240 second cooldown."
	got = parsePassiveCD(en)
	if len(got) != 1 || got[0] != 240 {
		t.Fatalf("EN cooldown: %v", got)
	}
	if parsePassiveCD("After killing 9 minions or monsters, Ahri heals.") != nil {
		t.Fatal("Ahri ne doit pas avoir de CD inventé")
	}
}

func TestParseCDString(t *testing.T) {
	if parseCDString("240s") != 240 {
		t.Fatal("240s")
	}
	if parseCDString("5 min") != 300 {
		t.Fatal("5 min")
	}
	if parseCDString("—") != 0 || parseCDString("auto") != 0 {
		t.Fatal("sentinelles")
	}
}

func TestFormatPickPassiveCD(t *testing.T) {
	if g := formatPassiveCD([]float64{240, 240, 240}); g != "240s" {
		t.Fatalf("static: %s", g)
	}
	vals := make([]float64, 18)
	for i := range vals {
		vals[i] = 22 - float64(i)*10/17
	}
	if g := formatPassiveCD(vals); g != "22 → 12s" {
		t.Fatalf("per-level: %s", g)
	}
	if g := formatPassiveCD([]float64{30, 24, 18, 12}); g != "30 → 24 → 18 → 12s" {
		t.Fatalf("ranks: %s", g)
	}
	if pickPassiveCD([]float64{240, 240, 240}, 11) != 240 {
		t.Fatal("pick static")
	}
	if pickPassiveCD([]float64{30, 24, 18, 12}, 1) != 30 {
		t.Fatal("pick early")
	}
	if pickPassiveCD([]float64{30, 24, 18, 12}, 18) != 12 {
		t.Fatal("pick late")
	}
	if pickPassiveCD(vals, 1) != 22 || pickPassiveCD(vals, 18) != 12 {
		t.Fatal("pick 18-level")
	}
}

func TestAniviaPassiveCD(t *testing.T) {
	d, err := getDetail("Anivia")
	if err != nil {
		t.Skipf("ddragon indisponible: %v", err)
	}
	card := normalize(d)
	var p Spell
	for _, s := range card.Important {
		if s.Spell == "P" {
			p = s
			break
		}
	}
	if parseCDString(p.CD) != 240 {
		t.Errorf("Anivia P CD attendu 240s, obtenu %q", p.CD)
	}
	lp := livePassiveSpell("Anivia", p.Name, p.Icon, p.Desc, 11, 0)
	if lp.CD != 240 {
		t.Errorf("Anivia live P CD=%v", lp.CD)
	}
}

func TestAhriPassiveHasNoCD(t *testing.T) {
	d, err := getDetail("Ahri")
	if err != nil {
		t.Skipf("ddragon indisponible: %v", err)
	}
	card := normalize(d)
	for _, s := range card.Important {
		if s.Spell == "P" && s.CD != "—" {
			t.Errorf("Ahri P ne doit pas avoir de CD, obtenu %q", s.CD)
		}
	}
}
