package main

import (
	"encoding/json"
	"strconv"
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

// Timings d'objectifs du patch 26.x + fenêtres d'alerte.
func TestObjectives(t *testing.T) {
	live := func(gameTime float64, events string) *rawLive {
		payload := `{"gameData":{"gameMode":"CLASSIC","mapNumber":11,"gameTime":` +
			strconv.FormatFloat(gameTime, 'f', -1, 64) + `},"allPlayers":[{"summonerName":"Jgl","team":"ORDER"},` +
			`{"summonerName":"EnemyJgl","team":"CHAOS"}],"events":{"Events":[` + events + `]}}`
		var raw rawLive
		if err := json.Unmarshal([]byte(payload), &raw); err != nil {
			t.Fatalf("payload de test invalide: %v", err)
		}
		return &raw
	}
	kill := func(name string, at float64) string {
		return `{"EventName":"` + name + `","EventTime":` + strconv.FormatFloat(at, 'f', -1, 64) + `,"KillerName":"Jgl"}`
	}

	obj := buildObjectives(live(120, ""))
	for _, c := range []struct {
		key  string
		want float64
	}{{"grubs", 480}, {"dragon", 300}, {"herald", 900}, {"baron", 1200}} {
		if got := obj[c.key].NextAt; got != c.want {
			t.Errorf("%s: spawn %.0f, attendu %.0f", c.key, got, c.want)
		}
	}
	if l := obj["grubs"].Leads; len(l) != 2 || l[0] != leadTeamfight || l[1] != leadGrubs {
		t.Errorf("alertes larves attendues à 120 s + 70 s, obtenu %v", l)
	}
	if l := obj["baron"].Leads; len(l) != 1 || l[0] != leadTeamfight {
		t.Errorf("alerte baron attendue à 120 s, obtenu %v", l)
	}
	if obj["herald"].Leads != nil {
		t.Errorf("le héraut ne doit pas déclencher d'alerte teamfight: %v", obj["herald"].Leads)
	}

	// Camp de larves terminé : plus de compte à rebours ni d'alerte.
	obj = buildObjectives(live(520, kill("HordeKill", 500)+","+kill("HordeKill", 505)+","+kill("HordeKill", 510)))
	if !obj["grubs"].Gone || obj["grubs"].Leads != nil {
		t.Errorf("3 larves tuées: camp attendu terminé, obtenu %+v", obj["grubs"])
	}
	// Une seule larve prise : le reste du camp est encore là.
	obj = buildObjectives(live(520, kill("VoidgrubKill", 500)))
	if obj["grubs"].Gone || !strings.Contains(obj["grubs"].Note, "2") {
		t.Errorf("1 larve tuée: 2 restantes attendues, obtenu %+v", obj["grubs"])
	}
	// Personne ne les touche : elles disparaissent à 14:45.
	if o := buildObjectives(live(900, ""))["grubs"]; !o.Gone {
		t.Errorf("larves attendues disparues après 14:45: %+v", o)
	}

	// 3 drakes pour une équipe → le prochain donne l'âme ; 4 → Ancestral.
	obj = buildObjectives(live(1000, kill("DragonKill", 300)+","+kill("DragonKill", 600)+","+kill("DragonKill", 900)))
	if obj["dragon"].Note != "ÂME" || obj["dragon"].NextAt != 1200 {
		t.Errorf("drake d'âme attendu à 900+300: %+v", obj["dragon"])
	}
	obj = buildObjectives(live(1300, kill("DragonKill", 300)+","+kill("DragonKill", 600)+","+
		kill("DragonKill", 900)+","+kill("DragonKill", 1200)))
	if !strings.Contains(strings.ToLower(obj["dragon"].Label), "ancestral") || obj["dragon"].NextAt != 1200+elderRespawn {
		t.Errorf("Ancestral attendu 6 min après le 4e drake: %+v", obj["dragon"])
	}

	// Baron tué → respawn 6 min plus tard, héraut définitivement absent.
	obj = buildObjectives(live(1400, kill("BaronKill", 1300)))
	if obj["baron"].NextAt != 1300+baronRespawn {
		t.Errorf("respawn baron attendu à 1660: %.0f", obj["baron"].NextAt)
	}
	if _, ok := obj["herald"]; ok {
		t.Errorf("le héraut ne devrait plus être listé après 19:45")
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
