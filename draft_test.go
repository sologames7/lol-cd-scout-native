package main

import (
	"strings"
	"testing"
)

func TestNormChamp(t *testing.T) {
	pairs := [][2]string{
		{"Kai'Sa", "kaisa"},
		{"Kaisa", "kaisa"},
		{"Nunu & Willump", "nunu"},
		{"Nunu et Willump", "nunu"},
		{"Lee Sin", "leesin"},
		{"Wukong", "wukong"},
		{"Cho'Gath", "chogath"},
		{"Dr. Mundo", "drmundo"},
	}
	for _, p := range pairs {
		if got := normChamp(p[0]); got != p[1] {
			t.Errorf("normChamp(%q)=%q, want %q", p[0], got, p[1])
		}
	}
}

func TestRiotName(t *testing.T) {
	if riotName("LeeSin") != "Lee Sin" {
		t.Errorf("LeeSin: %s", riotName("LeeSin"))
	}
	if riotName("MonkeyKing") != "Wukong" {
		t.Errorf("MonkeyKing: %s", riotName("MonkeyKing"))
	}
	if riotName("Kaisa") != "Kai'Sa" {
		t.Errorf("Kaisa: %s", riotName("Kaisa"))
	}
}

func TestListHas(t *testing.T) {
	ensureDraftIndex()
	list := []string{"Lee Sin", "Kai'Sa", "Nunu & Willump"}
	if !listHas(list, "LeeSin") || !listHas(list, "Kaisa") || !listHas(list, "Nunu") {
		t.Fatal("listHas devrait matcher les variantes de noms")
	}
	if listHas(list, "Zed") {
		t.Fatal("Zed ne doit pas matcher")
	}
}

func TestPlaysRole(t *testing.T) {
	if !playsRole(222, "BOT") || playsRole(222, "MID") {
		t.Fatal("Jinx doit être BOT seulement")
	}
	if !playsRole(89, "SUPP") {
		t.Fatal("Leona SUPP")
	}
	if !playsRole(238, "MID") || !playsRole(238, "JGL") {
		t.Fatal("Zed MID et JGL")
	}
	if !playsRole(24, "TOP") || !playsRole(24, "JGL") || playsRole(24, "BOT") {
		t.Fatal("Jax TOP/JGL")
	}
}

func TestDraftAdviceSamiraNautilus(t *testing.T) {
	if err := ensureIndex(); err != nil {
		t.Skipf("ddragon indisponible: %v", err)
	}
	adv := buildDraftAdvice([]int{238}, []int{111}, nil, "BOT", 0, "pick")
	if len(adv.Picks) == 0 {
		t.Fatal("attendu des picks BOT")
	}
	foundSamira := false
	for _, p := range adv.Picks {
		if p.ID == "Samira" || p.Key == 360 {
			foundSamira = true
		}
		if p.Key == 238 || p.Key == 111 {
			t.Errorf("pick déjà dans la draft: %s", p.Name)
		}
		if p.Role != "BOT" {
			t.Errorf("pick hors rôle: %s %s", p.Name, p.Role)
		}
	}
	if !foundSamira {
		names := make([]string, 0, len(adv.Picks))
		for _, p := range adv.Picks {
			names = append(names, p.Name)
		}
		t.Errorf("Samira absente des picks BOT avec Nautilus allié: %v", names)
	}
	if len(adv.Bans) == 0 {
		t.Fatal("attendu des bans (menaces pour Nautilus)")
	}
}

func TestDraftAdviceBansAllyCounters(t *testing.T) {
	if err := ensureIndex(); err != nil {
		t.Skipf("ddragon indisponible: %v", err)
	}
	adv := buildDraftAdvice(nil, []int{103}, nil, "MID", 0, "ban")
	if len(adv.Bans) == 0 {
		t.Fatal("attendu des bans pour protéger Ahri")
	}
	ct, _, _ := champLists("Ahri")
	hit := 0
	for _, b := range adv.Bans {
		if listHas(ct, b.Name) || listHas(ct, riotName(b.ID)) {
			hit++
		}
	}
	if hit == 0 {
		n := 5
		if len(ct) < n {
			n = len(ct)
		}
		t.Errorf("aucun ban n'est un counter d'Ahri: %+v / counters=%v", adv.Bans, ct[:n])
	}
}

func TestSynergiesLoaded(t *testing.T) {
	if len(synergies) < 150 {
		t.Fatalf("synergies trop courtes: %d", len(synergies))
	}
	jinx, ok := synergies["Jinx"]
	if !ok || !listHas(jinx, "Lulu") || !listHas(jinx, "Thresh") {
		t.Errorf("Jinx doit synerger avec Lulu/Thresh: %v", jinx)
	}
	xayah, ok := synergies["Xayah"]
	if !ok || !listHas(xayah, "Rakan") {
		t.Errorf("Xayah doit synerger avec Rakan: %v", xayah)
	}
	if !strings.Contains(strings.ToLower(strings.Join(synergies["Samira"], ",")), "nautilus") {
		t.Errorf("Samira générée sans Nautilus: %v", synergies["Samira"])
	}
}
