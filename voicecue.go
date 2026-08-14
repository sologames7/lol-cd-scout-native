package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Alertes audio en live : voiceline de select du jungle aux fenêtres de gank
// habituelles, voiceline + TTS quand il revient shop avec un gros spike d'or,
// TTS quand n'importe quel champion (sauf soi) complète un légendaire.
//
// Les cues restent ~10 s dans LiveState.Voices ; le front déduplique par ID.
// Voicelines : CommunityDragon champion-choose-vo (ogg FR), proxifiées
// par /api/voice pour rester same-origin dans WebView2.

type VoiceCue struct {
	ID       string  `json:"id"`
	Kind     string  `json:"kind"` // gank | shop | item
	Champion string  `json:"champion"`
	Key      int     `json:"key"`
	Title    string  `json:"title"`
	Sub      string  `json:"sub,omitempty"`
	Speak    string  `json:"speak,omitempty"` // TTS FR
	Voice    bool    `json:"voice,omitempty"` // jouer l'ogg du champion
	Hot      bool    `json:"hot,omitempty"`
	Until    float64 `json:"-"`
}

const (
	scuttleSpawn    = 195.0 // 3:15
	voiceCueKeep    = 10.0  // secondes de jeu à renvoyer au front
	shopGoldJump    = 900   // spike d'or itemique = shop conséquent
	shopGoldAfter   = 150.0 // ignorer le fill d'inventaire au spawn
	gankAlertWindow = 4.0   // comme les objectifs
	voiceCueCap     = 24
)

type jgWindow struct {
	id    string
	at    float64
	lead  float64
	label string
}

type jgClear int

const (
	jgFullClear jgClear = iota
	jgLevel2
	jgLevel3
)

// Invade / gank niveau 2 (~2:00).
var jgLevel2Keys = map[int]struct{}{
	20: {}, 35: {}, 427: {}, // Nunu, Shaco, Ivern
}

// Clear court, gank niveau 3 (~2:45) plutôt que full clear → crabe.
var jgLevel3Keys = map[int]struct{}{
	5: {}, 24: {}, 28: {}, 29: {}, 56: {}, 59: {}, 60: {}, 62: {}, 64: {},
	76: {}, 78: {}, 80: {}, 84: {}, 91: {}, 104: {}, 105: {}, 107: {},
	121: {}, 131: {}, 200: {}, 203: {}, 233: {}, 234: {}, 245: {}, 246: {},
	254: {}, 421: {}, 517: {}, 950: {},
	// Xin, Jax, Evelynn, Twitch, Nocturne, Jarvan, Elise, Wukong, Lee Sin,
	// Nidalee, Poppy, Pantheon, Akali, Talon, Graves, Fizz, Rengar,
	// Kha'Zix, Diana, Bel'Veth, Kindred, Briar, Viego, Ekko, Qiyana,
	// Vi, Rek'Sai, Sylas, Naafiri
}

func jungleClear(key int) jgClear {
	if _, ok := jgLevel2Keys[key]; ok {
		return jgLevel2
	}
	if _, ok := jgLevel3Keys[key]; ok {
		return jgLevel3
	}
	return jgFullClear
}

func jungleWindows(key int) []jgWindow {
	first, id, label := scuttleSpawn, "scuttle", "premier crabe"
	switch jungleClear(key) {
	case jgLevel2:
		first, id, label = 120, "lv2", "gank niveau 2"
	case jgLevel3:
		first, id, label = 165, "lv3", "gank niveau 3"
	}
	wins := []jgWindow{
		{id: id, at: first, lead: 15, label: label},
		{id: "rot2", at: first + 210, lead: 18, label: "2e rotation"},
		{id: "grubs", at: grubsSpawn, lead: 25, label: "larves"},
		{id: "rot3", at: first + 420, lead: 18, label: "3e rotation"},
		{id: "herald", at: heraldSpawn, lead: 20, label: "héraut"},
	}
	if first != scuttleSpawn && absf(first-scuttleSpawn) >= 50 {
		wins = append(wins, jgWindow{id: "scuttle", at: scuttleSpawn, lead: 15, label: "premier crabe"})
	}
	return wins
}

func absf(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func summonersRift(mode string) bool {
	switch strings.ToUpper(mode) {
	case "", "CLASSIC", "PRACTICETOOL":
		return true
	default:
		return false
	}
}

func livePID(p LivePlayer) string {
	return p.RiotID + "#" + strconv.Itoa(p.Key)
}

var voiceMem = struct {
	mu        sync.Mutex
	gameKey   string
	lastT     float64
	recent    []VoiceCue
	seenItems map[string]map[int]bool
	lastGold  map[string]int
	firedGank map[string]bool
	primed    map[string]bool
}{}

func resetVoiceMem() {
	voiceMem.gameKey = ""
	voiceMem.lastT = 0
	voiceMem.recent = nil
	voiceMem.seenItems = map[string]map[int]bool{}
	voiceMem.lastGold = map[string]int{}
	voiceMem.firedGank = map[string]bool{}
	voiceMem.primed = map[string]bool{}
}

func pushCue(c VoiceCue, now float64) {
	c.Until = now + voiceCueKeep
	if len(voiceMem.recent) >= voiceCueCap {
		voiceMem.recent = voiceMem.recent[1:]
	}
	voiceMem.recent = append(voiceMem.recent, c)
}

func updateVoiceCues(state *LiveState) {
	if state == nil || !state.Active {
		return
	}
	key := liveGameKey(*state)
	meta := itemMetaTable()

	voiceMem.mu.Lock()
	defer voiceMem.mu.Unlock()
	if voiceMem.gameKey != key || state.GameTime+5 < voiceMem.lastT {
		resetVoiceMem()
		voiceMem.gameKey = key
	}
	voiceMem.lastT = state.GameTime

	keep := voiceMem.recent[:0]
	for _, c := range voiceMem.recent {
		if c.Until > state.GameTime {
			keep = append(keep, c)
		}
	}
	voiceMem.recent = keep

	scan := func(list []LivePlayer, enemy bool) {
		for i := range list {
			p := &list[i]
			p.ItemGold = itemGoldSum(p.Items, meta)
			collectPlayerCues(*p, enemy, state, meta)
		}
	}
	scan(state.Enemies, true)
	scan(state.Allies, false)

	out := make([]VoiceCue, len(voiceMem.recent))
	copy(out, voiceMem.recent)
	state.Voices = out
}

func collectPlayerCues(p LivePlayer, enemy bool, state *LiveState, meta map[int]itemMeta) {
	pid := livePID(p)
	if voiceMem.seenItems[pid] == nil {
		voiceMem.seenItems[pid] = map[int]bool{}
	}
	seen := voiceMem.seenItems[pid]
	primed := voiceMem.primed[pid]

	type doneItem struct {
		id   int
		meta itemMeta
	}
	var finished []doneItem
	for _, id := range p.Items {
		if id == 0 {
			continue
		}
		m, ok := meta[id]
		if !ok || !m.Complete || seen[id] {
			continue
		}
		seen[id] = true
		if primed && !p.IsMe {
			finished = append(finished, doneItem{id, m})
		}
	}

	prev, hadGold := voiceMem.lastGold[pid]
	goldJump := 0
	if hadGold && primed {
		goldJump = p.ItemGold - prev
	}
	voiceMem.lastGold[pid] = p.ItemGold
	voiceMem.primed[pid] = true
	if !primed {
		go prefetchChampionVO(p.Key)
	}

	isJg := strings.EqualFold(p.Position, "JUNGLE")
	now := state.GameTime

	if primed {
		for _, it := range finished {
			name := shortItemName(it.meta.Name)
			if name == "" {
				name = "un item"
			}
			pushCue(VoiceCue{
				ID:       fmt.Sprintf("item|%s|%d", pid, it.id),
				Kind:     "item",
				Champion: p.Champion,
				Key:      p.Key,
				Title:    strings.ToUpper(p.Champion) + " · ITEM",
				Sub:      name,
				Speak:    p.Champion + " a complété " + name,
				Voice:    isJg && enemy,
				Hot:      enemy,
			}, now)
		}

		if isJg && enemy && goldJump >= shopGoldJump && now >= shopGoldAfter && len(finished) == 0 {
			pushCue(VoiceCue{
				ID:       fmt.Sprintf("shop|%s|%.0f", pid, now),
				Kind:     "shop",
				Champion: p.Champion,
				Key:      p.Key,
				Title:    strings.ToUpper(p.Champion) + " · SHOP",
				Sub:      fmt.Sprintf("spike itemique +%d or — il ressort fort", goldJump),
				Speak:    p.Champion + " a shop",
				Voice:    true,
				Hot:      true,
			}, now)
		}
	}

	if !isJg || p.IsDead || !summonersRift(state.GameMode) {
		return
	}
	for _, w := range jungleWindows(p.Key) {
		fireAt := w.at - w.lead
		if now < fireAt || now >= fireAt+gankAlertWindow {
			continue
		}
		gkey := fmt.Sprintf("gank|%s|%s|%.0f", pid, w.id, w.at)
		if voiceMem.firedGank[gkey] {
			continue
		}
		voiceMem.firedGank[gkey] = true
		who := "jungle ennemi — ward et recule"
		if !enemy {
			who = "ton jungle — synchro possible"
		}
		pushCue(VoiceCue{
			ID:       gkey,
			Kind:     "gank",
			Champion: p.Champion,
			Key:      p.Key,
			Title:    strings.ToUpper(p.Champion) + " · " + strings.ToUpper(w.label),
			Sub:      who,
			Voice:    true,
			Hot:      enemy,
		}, now)
	}
}

func prefetchChampionVO(key int) {
	if key <= 0 {
		return
	}
	_, _ = championChooseVO(key)
}

var voCache = struct {
	mu   sync.Mutex
	data map[int][]byte
}{data: map[int][]byte{}}

func championChooseVO(key int) ([]byte, error) {
	if key <= 0 || key > 9999 {
		return nil, fmt.Errorf("clé invalide")
	}
	voCache.mu.Lock()
	if b, ok := voCache.data[key]; ok {
		voCache.mu.Unlock()
		return b, nil
	}
	voCache.mu.Unlock()

	urls := []string{
		fmt.Sprintf("https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/fr_fr/v1/champion-choose-vo/%d.ogg", key),
		fmt.Sprintf("https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default/v1/champion-choose-vo/%d.ogg", key),
	}
	var last error
	for _, u := range urls {
		b, err := fetchVoiceOGG(u)
		if err != nil {
			last = err
			continue
		}
		voCache.mu.Lock()
		voCache.data[key] = b
		voCache.mu.Unlock()
		return b, nil
	}
	if last == nil {
		last = fmt.Errorf("introuvable")
	}
	return nil, last
}

func fetchVoiceOGG(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "LoL-CD-Scout/0.2")
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d", res.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(res.Body, 256<<10))
	if err != nil {
		return nil, err
	}
	if len(b) < 4 || string(b[:4]) != "OggS" {
		return nil, fmt.Errorf("pas un ogg")
	}
	return b, nil
}

func apiVoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "GET", http.StatusMethodNotAllowed)
		return
	}
	key, _ := strconv.Atoi(r.URL.Query().Get("key"))
	b, err := championChooseVO(key)
	if err != nil {
		http.Error(w, "voiceline introuvable", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "audio/ogg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = io.Copy(w, bytes.NewReader(b))
}
