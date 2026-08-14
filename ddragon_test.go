package main

import "testing"

func TestDDragonPathOK(t *testing.T) {
	ok := []string{
		"16.15.1/img/champion/Ahri.png",
		"16.15.1/img/spell/AhriQ.png",
		"16.15.1/img/passive/Anivia_P.png",
		"16.15.1/img/item/3158.png",
		"16.15.1/data/fr_FR/champion.json",
		"16.15.1/img/champion/Kha%27Zix.png",
		"img/perk-images/Styles/Domination/Electrocute/Electrocute.png",
		"img/perk-images/Styles/7200_Domination.png",
	}
	for _, p := range ok {
		if !ddragonPathOK(p) {
			t.Errorf("attendu OK: %s", p)
		}
	}
	bad := []string{
		"",
		"../etc/passwd",
		"16.15.1/img/../../../etc/passwd",
		"https://evil.example/x",
		"16.15.1/img//Ahri.png",
		"foo/img/champion/Ahri.png",
		"16.15.1/other/Ahri.png",
		"img/other/foo.png",
		"img/perk-images/../secret.png",
	}
	for _, p := range bad {
		if ddragonPathOK(p) {
			t.Errorf("attendu rejeté: %s", p)
		}
	}
}

func TestGetVersionNeverEmpty(t *testing.T) {
	if v := getVersion(); v == "" {
		t.Fatal("getVersion ne doit jamais renvoyer une chaîne vide")
	}
}
