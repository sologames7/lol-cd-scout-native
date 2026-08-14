package main

import (
	"strings"
	"testing"
)

func TestItemIsCompleteAndNames(t *testing.T) {
	if itemIsComplete(300, nil, []string{"3006"}, []string{"Boots"}) {
		t.Error("bottes de base ne sont pas un item fini")
	}
	if itemIsComplete(1100, []string{"1001"}, nil, []string{"Boots"}) {
		t.Error("bottes T2 ne doivent pas être annoncées")
	}
	if itemIsComplete(1000, []string{"1036"}, []string{"3142"}, nil) {
		t.Error("Pique dentelée (composant) ne doit pas passer")
	}
	if !itemIsComplete(2700, []string{"3134", "1036"}, nil, []string{"Damage"}) {
		t.Error("Youmuu doit être un item fini")
	}
	if !itemIsComplete(1600, []string{"1082"}, nil, nil) {
		t.Error("Mejai (1600, from Dark Seal) doit passer")
	}
	if itemIsComplete(450, nil, nil, []string{"Jungle"}) {
		t.Error("pet jungle ne doit pas passer")
	}
	if itemGoldSum([]int{0, 3142, 1036}, map[int]itemMeta{3142: {Gold: 2700}, 1036: {Gold: 350}}) != 3050 {
		t.Error("somme d'or itemique incorrecte")
	}
	if got := shortItemName("Youmuu, lame spectrale"); got != "Youmuu" {
		t.Errorf("shortItemName: %q", got)
	}
	if got := shortItemName("Force de la trinité"); got != "Force de la trinité" {
		t.Errorf("shortItemName conservé: %q", got)
	}
}

func TestItemMetaTable(t *testing.T) {
	table := itemMetaTable()
	if len(table) < 50 {
		t.Skipf("ddragon item.json indisponible (%d)", len(table))
	}
	// 3158 Lucidité : bottes, pas un légendaire.
	if m := table[3158]; m.Name == "" || m.Complete || m.Gold < 800 {
		t.Errorf("Lucidité mal parsée: %+v", m)
	}
	// 3134 Pique dentelée : composant.
	if m := table[3134]; m.Complete {
		t.Errorf("Pique dentelée ne doit pas être complete: %+v", m)
	}
	// 3031 Infini / 3078 Trinité : légendaires.
	ie, tr := table[3031], table[3078]
	if !ie.Complete && !tr.Complete {
		t.Errorf("aucun légendaire reconnu (IE %+v, Trinité %+v)", ie, tr)
	}
	t.Logf("items: %d · IE complete=%v · Trinité complete=%v", len(table), ie.Complete, tr.Complete)
}

func TestJungleWindows(t *testing.T) {
	if jungleClear(64) != jgLevel3 { // Lee Sin
		t.Fatal("Lee Sin doit être un gank niv.3")
	}
	if jungleClear(141) != jgFullClear { // Kayn
		t.Fatal("Kayn doit être full clear / crabe")
	}
	if jungleClear(35) != jgLevel2 { // Shaco
		t.Fatal("Shaco doit être niv.2")
	}
	lee := jungleWindows(64)
	if len(lee) < 4 || lee[0].id != "lv3" || lee[0].at != 165 {
		t.Fatalf("fenêtre Lee Sin: %+v", lee)
	}
	kayn := jungleWindows(141)
	if kayn[0].id != "scuttle" || kayn[0].at != scuttleSpawn {
		t.Fatalf("fenêtre Kayn: %+v", kayn)
	}
	if !summonersRift("CLASSIC") || summonersRift("ARAM") {
		t.Fatal("filtre de map incorrect")
	}
}

func seedItemMeta() {
	itemMetaCache.mu.Lock()
	itemMetaCache.version = getVersion()
	itemMetaCache.data = map[int]itemMeta{
		3142: {Name: "Youmuu, lame spectrale", Gold: 2700, Complete: true},
		1036: {Name: "Épée longue", Gold: 350, Complete: false},
		1001: {Name: "Bottes de vitesse", Gold: 300, Complete: false},
		3134: {Name: "Pique dentelée", Gold: 1000, Complete: false},
	}
	itemMetaCache.mu.Unlock()
}

func TestVoiceCuesGankShopItem(t *testing.T) {
	seedItemMeta()
	voiceMem.mu.Lock()
	resetVoiceMem()
	voiceMem.mu.Unlock()
	t.Cleanup(func() {
		voiceMem.mu.Lock()
		resetVoiceMem()
		voiceMem.mu.Unlock()
		resetItemMetaCache()
	})

	kha := LivePlayer{RiotID: "Kha#EUW", Champion: "Kha'Zix", Key: 121, Position: "JUNGLE", Items: []int{1036, 1001}}
	st := &LiveState{Active: true, GameTime: 40, GameMode: "CLASSIC", Enemies: []LivePlayer{kha}}
	updateVoiceCues(st)
	if len(st.Voices) != 0 {
		t.Fatalf("premier snapshot ne doit rien annoncer: %+v", st.Voices)
	}

	// Fenêtre gank niv.3 Kha'Zix : 165-15 = 150.
	st.GameTime = 151
	st.Enemies[0].Items = []int{1036, 1001}
	updateVoiceCues(st)
	if !hasKind(st.Voices, "gank") {
		t.Fatalf("gank niv.3 attendu à 2:30, obtenu %+v", st.Voices)
	}
	updateVoiceCues(st)
	if n := countKind(st.Voices, "gank"); n != 1 {
		t.Fatalf("gank doit rester unique (renvoyé 10 s), pas se dupliquer: %d", n)
	}

	// Shop sans légendaire : +1000 (pique) + 350.
	st.GameTime = 320
	st.Enemies[0].Items = []int{1001, 3134, 1036, 1036}
	updateVoiceCues(st)
	if !hasKind(st.Voices, "shop") {
		t.Fatalf("shop jungle attendu, obtenu %+v", st.Voices)
	}

	// Jungle mort : pas de 2e rotation même dans la fenêtre (Kha rot2 ~6:00).
	st.GameTime = 165 + 210 - 18 + 1
	st.Enemies[0].IsDead = true
	updateVoiceCues(st)
	for _, c := range st.Voices {
		if strings.Contains(c.ID, "rot2") {
			t.Fatalf("pas de gank si le jungle est mort: %+v", c)
		}
	}
	st.Enemies[0].IsDead = false

	// Légendaire : Youmuu, pas de 2e cue shop.
	st.GameTime = 400
	st.Enemies[0].Items = []int{3142, 1001}
	updateVoiceCues(st)
	if !hasKind(st.Voices, "item") {
		t.Fatalf("item Youmuu attendu, obtenu %+v", st.Voices)
	}
	var itemCue VoiceCue
	for _, c := range st.Voices {
		if c.Kind == "item" {
			itemCue = c
		}
	}
	if !itemCue.Voice || !strings.Contains(itemCue.Speak, "Youmuu") {
		t.Fatalf("item jungle ennemi : VO + TTS Youmuu, obtenu %+v", itemCue)
	}

	// Soi-même : pas d'annonce.
	me := LivePlayer{RiotID: "Moi#EUW", Champion: "Ahri", Key: 103, Position: "MIDDLE", IsMe: true, Items: []int{3142}}
	st2 := &LiveState{Active: true, GameTime: 500, GameMode: "CLASSIC", Allies: []LivePlayer{me}, Enemies: []LivePlayer{kha}}
	updateVoiceCues(st2) // prime
	st2.GameTime = 510
	updateVoiceCues(st2)
	for _, c := range st2.Voices {
		if c.Kind == "item" && c.Champion == "Ahri" {
			t.Fatalf("pas d'annonce pour soi: %+v", c)
		}
	}

	// ARAM : pas de gank.
	voiceMem.mu.Lock()
	resetVoiceMem()
	voiceMem.mu.Unlock()
	aram := LivePlayer{RiotID: "Lee#EUW", Champion: "Lee Sin", Key: 64, Position: "JUNGLE", Items: []int{1001}}
	st3 := &LiveState{Active: true, GameTime: 151, GameMode: "ARAM", Enemies: []LivePlayer{aram}}
	updateVoiceCues(st3)
	st3.GameTime = 152
	updateVoiceCues(st3)
	if hasKind(st3.Voices, "gank") {
		t.Fatalf("pas de gank en ARAM: %+v", st3.Voices)
	}
}

func hasKind(list []VoiceCue, kind string) bool {
	return countKind(list, kind) > 0
}

func countKind(list []VoiceCue, kind string) int {
	n := 0
	for _, c := range list {
		if c.Kind == kind {
			n++
		}
	}
	return n
}
