package main

import (
	"strings"
	"testing"
)

// Vérifie que la haste est bien extraite des descriptions Data Dragon (réseau requis).
func TestItemHasteTable(t *testing.T) {
	table := itemHasteTable()
	if len(table) < 20 {
		t.Fatalf("table de haste trop petite: %d items", len(table))
	}
	// 3158 = Bottes de lucidité ioniennes : ability haste + summoner spell haste.
	boots, ok := table[3158]
	if !ok || boots.Ability <= 0 || boots.Summoner <= 0 {
		t.Errorf("Bottes de lucidité mal parsées: %+v (ok=%v)", boots, ok)
	}
	t.Logf("items avec haste: %d · Lucidité: %+v", len(table), boots)
}

func TestGoldEstimateAndPairing(t *testing.T) {
	// À 20 min avec 180 CS, 3/... l'estimation doit être dans un ordre de grandeur crédible (~7-9k).
	p := LivePlayer{CS: 180, Kills: 3, Assists: 4}
	g := goldEstimate(1200, p)
	if g < 6000 || g > 10000 {
		t.Errorf("estimation d'or hors plage crédible: %d", g)
	}
	mk := func(champ, pos string) LivePlayer { return LivePlayer{Champion: champ, Position: pos, RiotID: champ} }
	allies := []LivePlayer{mk("Garen", "TOP"), mk("LeeSin", "JUNGLE"), mk("Ahri", "MIDDLE"), mk("Jinx", "BOTTOM"), mk("Lulu", "UTILITY")}
	enemies := []LivePlayer{mk("Jax", "MIDDLE"), mk("Vi", "JUNGLE"), mk("Sion", "TOP"), mk("Milio", "UTILITY"), mk("Kaisa", "BOTTOM")}
	metas, pairs := pairRoles(allies, enemies)
	if len(pairs) != 5 || metas[0].Label != "TOP" || enemies[pairs[0][1]].Champion != "Sion" {
		t.Errorf("appariement par rôle incorrect: %+v / %+v", metas, pairs)
	}
	// Positions manquantes (bots) → appariement par index.
	for i := range enemies {
		enemies[i].Position = ""
	}
	metas, pairs = pairRoles(allies, enemies)
	if len(pairs) != 5 || pairs[2] != [2]int{2, 2} {
		t.Errorf("appariement par index incorrect: %+v / %+v", metas, pairs)
	}
	// Historique : reset sur nouvelle partie + échantillonnage.
	st := &LiveState{Active: true, GameTime: 30, Allies: allies, Enemies: enemies}
	goldHist.gameKey, goldHist.samples, goldHist.lastT = "", nil, 0
	updateGold(st)
	st.GameTime = 35
	updateGold(st)
	st.GameTime = 45
	updateGold(st)
	if st.Gold == nil || len(goldHist.samples) != 2 {
		t.Fatalf("échantillonnage attendu: 2 samples, obtenu %d", len(goldHist.samples))
	}
	st.GameTime = 48 // pas de nouvel échantillon (<10 s), mais point courant ajouté
	updateGold(st)
	if len(goldHist.samples) != 2 || len(st.Gold.History) != 3 {
		t.Errorf("attendu 2 samples + point courant (3), obtenu %d samples / %d points", len(goldHist.samples), len(st.Gold.History))
	}
}

// Les icônes ddragon des sorts doivent être exposées aux cartes (réseau requis).
func TestSpellIcons(t *testing.T) {
	d, err := getDetail("Anivia")
	if err != nil {
		t.Skipf("ddragon indisponible: %v", err)
	}
	card := normalize(d)
	seen := map[string]string{}
	for _, s := range card.Important {
		seen[s.Spell] = s.Icon
	}
	for _, k := range []string{"Q", "W", "E", "R"} {
		if !strings.HasPrefix(seen[k], "spell/") {
			t.Errorf("icône de %s manquante ou invalide: %q", k, seen[k])
		}
	}
	// Anivia est curated avec une entrée passive : elle doit pointer vers img/passive.
	if p, ok := seen["P"]; ok && !strings.HasPrefix(p, "passive/") {
		t.Errorf("icône du passif invalide: %q", p)
	}
}

func TestRanks(t *testing.T) {
	cases := []struct{ level, ult, basic int }{
		{1, 0, 1}, {3, 0, 2}, {6, 1, 3}, {9, 1, 5}, {11, 2, 5}, {16, 3, 5}, {18, 3, 5},
	}
	for _, c := range cases {
		if got := ultRank(c.level); got != c.ult {
			t.Errorf("ultRank(%d)=%d, attendu %d", c.level, got, c.ult)
		}
		if got := maxBasicRank(c.level); got != c.basic {
			t.Errorf("maxBasicRank(%d)=%d, attendu %d", c.level, got, c.basic)
		}
	}
}
