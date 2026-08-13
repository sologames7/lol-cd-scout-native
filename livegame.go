package main

// Live Client Data API (Riot, locale, officielle) : https://127.0.0.1:2999/liveclientdata/allgamedata
// Disponible uniquement pendant une partie, en lecture seule.
// On en tire : niveaux, items (=> ability haste), morts/respawns, sorts d'invocateur,
// événements (drakes / héraut / baron) — pour tous les joueurs de la partie.

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const liveURL = "https://127.0.0.1:2999/liveclientdata/allgamedata"

var liveHTTP = &http.Client{
	Timeout: 1500 * time.Millisecond,
	Transport: &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, // certificat auto-signé Riot en loopback
		MaxIdleConns:        2,
		MaxIdleConnsPerHost: 2,
		IdleConnTimeout:     4 * time.Second,
		ForceAttemptHTTP2:   false,
	},
}

// ---------- Payload brut Riot ----------

type rawLiveSummSpell struct {
	DisplayName    string `json:"displayName"`
	RawDisplayName string `json:"rawDisplayName"`
}

type rawLivePlayer struct {
	ChampionName    string `json:"championName"`
	RawChampionName string `json:"rawChampionName"`
	IsDead          bool   `json:"isDead"`
	Items           []struct {
		ItemID int `json:"itemID"`
	} `json:"items"`
	Level          int     `json:"level"`
	Position       string  `json:"position"`
	RespawnTimer   float64 `json:"respawnTimer"`
	RiotID         string  `json:"riotId"`
	RiotIDGameName string  `json:"riotIdGameName"`
	Scores         struct {
		Kills      int `json:"kills"`
		Deaths     int `json:"deaths"`
		Assists    int `json:"assists"`
		CreepScore int `json:"creepScore"`
	} `json:"scores"`
	SummonerName   string `json:"summonerName"`
	SummonerSpells struct {
		One rawLiveSummSpell `json:"summonerSpellOne"`
		Two rawLiveSummSpell `json:"summonerSpellTwo"`
	} `json:"summonerSpells"`
	Team string `json:"team"`
}

type rawLive struct {
	ActivePlayer struct {
		RiotID       string `json:"riotId"`
		SummonerName string `json:"summonerName"`
	} `json:"activePlayer"`
	AllPlayers []rawLivePlayer `json:"allPlayers"`
	Events     struct {
		Events []struct {
			EventName  string  `json:"EventName"`
			EventTime  float64 `json:"EventTime"`
			DragonType string  `json:"DragonType"`
			KillerName string  `json:"KillerName"`
		} `json:"Events"`
	} `json:"events"`
	GameData struct {
		GameMode  string  `json:"gameMode"`
		GameTime  float64 `json:"gameTime"`
		MapNumber int     `json:"mapNumber"`
	} `json:"gameData"`
}

// ---------- Haste des items (parsée depuis Data Dragon) ----------

type itemHaste struct {
	Ability  int // Ability Haste
	Ultimate int // Ultimate Haste
	Summoner int // Summoner Spell Haste
}

var itemHasteCache = struct {
	mu      sync.Mutex
	version string
	data    map[int]itemHaste
}{}

var attentionRe = regexp.MustCompile(`(?i)<attention>\s*(\d+)\s*</attention>\s*([^<]*)`)

// Les hastes "passives" sont en texte brut : "Gain 10 Summoner Spell Haste", "Gain 30 Ultimate Haste".
var plainHasteRe = regexp.MustCompile(`(?i)(\d+)\s+(summoner spell haste|ultimate haste)`)

func itemHasteTable() map[int]itemHaste {
	version := getVersion()
	itemHasteCache.mu.Lock()
	defer itemHasteCache.mu.Unlock()
	if itemHasteCache.version == version && itemHasteCache.data != nil {
		return itemHasteCache.data
	}
	var payload struct {
		Data map[string]struct {
			Description string `json:"description"`
		} `json:"data"`
	}
	url := fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/item.json", version)
	if err := jsonGET(url, &payload); err != nil {
		// On garde l'ancienne table si le refresh échoue.
		if itemHasteCache.data != nil {
			return itemHasteCache.data
		}
		return map[int]itemHaste{}
	}
	table := map[int]itemHaste{}
	for idStr, item := range payload.Data {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		h := itemHaste{}
		for _, m := range attentionRe.FindAllStringSubmatch(item.Description, -1) {
			val, _ := strconv.Atoi(m[1])
			label := strings.ToLower(strings.TrimSpace(m[2]))
			switch {
			case strings.HasPrefix(label, "ability haste"):
				h.Ability += val
			case strings.HasPrefix(label, "ultimate haste"):
				h.Ultimate += val
			case strings.HasPrefix(label, "summoner spell haste"):
				h.Summoner += val
			}
		}
		for _, m := range plainHasteRe.FindAllStringSubmatch(item.Description, -1) {
			val, _ := strconv.Atoi(m[1])
			if strings.HasPrefix(strings.ToLower(m[2]), "summoner") {
				h.Summoner += val
			} else {
				h.Ultimate += val
			}
		}
		if h != (itemHaste{}) {
			table[id] = h
		}
	}
	itemHasteCache.version, itemHasteCache.data = version, table
	return table
}

// ---------- Sorts d'invocateur ----------

var summonerBaseCD = map[string]float64{
	"SummonerFlash":    300,
	"SummonerTeleport": 360,
	"SummonerHaste":    210, // Ghost
	"SummonerHeal":     240,
	"SummonerBarrier":  180,
	"SummonerExhaust":  210,
	"SummonerDot":      180, // Ignite
	"SummonerBoost":    210, // Cleanse
	"SummonerSmite":    90,  // recharge d'une charge
	"SummonerMana":     240, // Clarity (ARAM)
	"SummonerSnowball": 80,  // Mark (ARAM)
}

var summSlugRe = regexp.MustCompile(`SummonerSpell_(\w+?)_`)

func summonerSlug(s rawLiveSummSpell) string {
	if m := summSlugRe.FindStringSubmatch(s.RawDisplayName); m != nil {
		return m[1]
	}
	return ""
}

// ---------- Estimation des rangs / cooldowns ----------

func ultRank(level int) int {
	switch {
	case level >= 16:
		return 3
	case level >= 11:
		return 2
	case level >= 6:
		return 1
	}
	return 0
}

// Rang max plausible d'un sort de base : borné par 5, par ceil(level/2)
// et par le nombre de points de base disponibles (niveau - points d'ultime).
func maxBasicRank(level int) int {
	r := (level + 1) / 2
	if r > 5 {
		r = 5
	}
	if pts := level - ultRank(level); r > pts {
		r = pts
	}
	if r < 0 {
		r = 0
	}
	return r
}

func parseCDBurn(burn string) []float64 {
	parts := strings.Split(burn, "/")
	out := make([]float64, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return nil
		}
		out = append(out, v)
	}
	return out
}

func hasteFactor(haste int) float64 { return 100.0 / (100.0 + float64(haste)) }

// ---------- Réponse /api/live ----------

type LiveSpell struct {
	Spell string  `json:"spell"`
	Name  string  `json:"name"`
	Icon  string  `json:"icon,omitempty"` // chemin relatif ddragon, ex "spell/AhriQ.png"
	CD    float64 `json:"cd"`   // secondes, haste incluse, 0 = inconnu
	Rank  string  `json:"rank"` // "2" exact (R) ou "≤4" plausible
	Est   bool    `json:"est"`  // true si le rang est estimé
}

type LiveSummoner struct {
	Slug string  `json:"slug"`
	Name string  `json:"name"`
	CD   float64 `json:"cd"` // secondes, haste invocateur (items) incluse
}

type LivePlayer struct {
	RiotID       string         `json:"riotId"`
	IsMe         bool           `json:"isMe,omitempty"`
	Champion     string         `json:"champion"`
	Key          int            `json:"key"`
	Icon         string         `json:"icon"`
	Position     string         `json:"position"`
	Level        int            `json:"level"`
	IsDead       bool           `json:"isDead"`
	Respawn      float64        `json:"respawn"`
	Kills        int            `json:"kills"`
	Deaths       int            `json:"deaths"`
	Assists      int            `json:"assists"`
	CS           int            `json:"cs"`
	AbilityHaste int            `json:"abilityHaste"`
	SummHaste    int            `json:"summHaste"`
	Items        []int          `json:"items"`
	Spells       []LiveSpell    `json:"spells"`
	Summoners    []LiveSummoner `json:"summoners"`
}

type LiveObjective struct {
	NextAt float64   `json:"nextAt"` // temps de jeu (s) du prochain spawn, 0 = aucun
	Label  string    `json:"label"`
	Order  int       `json:"order,omitempty"`
	Chaos  int       `json:"chaos,omitempty"`
	Note   string    `json:"note,omitempty"`  // "ÂME", "×2 restantes"…
	Leads  []float64 `json:"leads,omitempty"` // anticipations d'alerte (s avant le spawn)
	Gone   bool       `json:"gone,omitempty"` // objectif définitivement parti (despawn / camp fini)
	Rank   int        `json:"rank,omitempty"` // ordre d'affichage dans la barre
}

type LiveState struct {
	Active     bool                     `json:"active"`
	GameTime   float64                  `json:"gameTime"`
	GameMode   string                   `json:"gameMode"`
	Enemies    []LivePlayer             `json:"enemies"`
	Allies     []LivePlayer             `json:"allies"`
	Objectives map[string]LiveObjective `json:"objectives"`
	Gold       *GoldInfo                `json:"gold,omitempty"`
}

// ---------- Or estimé ----------
// L'API live n'expose l'or exact que du joueur actif. Pour les autres on estime :
// or de départ + or passif + CS + kills/assists. Suffisant pour la tendance.

type GoldSample struct {
	T     float64  `json:"t"`
	Ally  int      `json:"ally"`
	Enemy int      `json:"enemy"`
	Roles [][2]int `json:"roles"` // [or allié, or ennemi] par ligne de GoldInfo.Roles
}

type GoldRoleMeta struct {
	Label      string `json:"label"`
	AllyChamp  string `json:"allyChamp"`
	EnemyChamp string `json:"enemyChamp"`
}

type GoldInfo struct {
	Estimated bool           `json:"estimated"`
	Roles     []GoldRoleMeta `json:"roles"`
	History   []GoldSample   `json:"history"`
}

func goldEstimate(t float64, p LivePlayer) int {
	g := 500.0 // or de départ
	if t > 110 {
		g += (t - 110) * 2.04 // or passif (20.4 / 10 s à partir de 1:50)
	}
	g += 20.8 * float64(p.CS)
	g += 300*float64(p.Kills) + 65*float64(p.Assists)
	return int(g)
}

var posOrder = []struct{ pos, label string }{
	{"TOP", "TOP"}, {"JUNGLE", "JGL"}, {"MIDDLE", "MID"}, {"BOTTOM", "BOT"}, {"UTILITY", "SUPP"},
}

func posLabel(pos string) string {
	for _, o := range posOrder {
		if o.pos == pos {
			return o.label
		}
	}
	return ""
}

// Apparie allié/ennemi par rôle quand les positions sont complètes, sinon par index.
func pairRoles(allies, enemies []LivePlayer) ([]GoldRoleMeta, [][2]int) {
	indexByPos := func(list []LivePlayer) map[string]int {
		m := map[string]int{}
		for i, p := range list {
			if p.Position != "" {
				if _, dup := m[p.Position]; !dup {
					m[p.Position] = i
				}
			}
		}
		return m
	}
	am, em := indexByPos(allies), indexByPos(enemies)
	if len(allies) == 5 && len(enemies) == 5 && len(am) == 5 && len(em) == 5 {
		metas := make([]GoldRoleMeta, 0, 5)
		pairs := make([][2]int, 0, 5)
		ok := true
		for _, o := range posOrder {
			ai, aok := am[o.pos]
			ei, eok := em[o.pos]
			if !aok || !eok {
				ok = false
				break
			}
			metas = append(metas, GoldRoleMeta{Label: o.label, AllyChamp: allies[ai].Champion, EnemyChamp: enemies[ei].Champion})
			pairs = append(pairs, [2]int{ai, ei})
		}
		if ok {
			return metas, pairs
		}
	}
	n := len(allies)
	if len(enemies) < n {
		n = len(enemies)
	}
	if n > 5 {
		n = 5
	}
	metas := make([]GoldRoleMeta, 0, n)
	pairs := make([][2]int, 0, n)
	for i := 0; i < n; i++ {
		label := posLabel(allies[i].Position)
		if label == "" {
			label = fmt.Sprintf("N°%d", i+1)
		}
		metas = append(metas, GoldRoleMeta{Label: label, AllyChamp: allies[i].Champion, EnemyChamp: enemies[i].Champion})
		pairs = append(pairs, [2]int{i, i})
	}
	return metas, pairs
}

var goldHist = struct {
	mu      sync.Mutex
	gameKey string
	lastT   float64
	samples []GoldSample
}{}

const goldSampleEvery = 10.0 // secondes de jeu entre deux échantillons

func updateGold(state *LiveState) {
	if len(state.Allies) == 0 || len(state.Enemies) == 0 {
		return
	}
	metas, pairs := pairRoles(state.Allies, state.Enemies)
	cur := GoldSample{T: state.GameTime}
	for _, p := range state.Allies {
		cur.Ally += goldEstimate(state.GameTime, p)
	}
	for _, p := range state.Enemies {
		cur.Enemy += goldEstimate(state.GameTime, p)
	}
	for _, pr := range pairs {
		cur.Roles = append(cur.Roles, [2]int{
			goldEstimate(state.GameTime, state.Allies[pr[0]]),
			goldEstimate(state.GameTime, state.Enemies[pr[1]]),
		})
	}
	ids := make([]string, 0, 10)
	for _, p := range state.Allies {
		ids = append(ids, p.RiotID)
	}
	for _, p := range state.Enemies {
		ids = append(ids, p.RiotID)
	}
	key := strings.Join(ids, "|")

	goldHist.mu.Lock()
	defer goldHist.mu.Unlock()
	if goldHist.gameKey != key || state.GameTime+5 < goldHist.lastT {
		goldHist.samples, goldHist.lastT = nil, 0 // nouvelle partie
	}
	goldHist.gameKey = key
	if len(goldHist.samples) == 0 || state.GameTime-goldHist.lastT >= goldSampleEvery {
		goldHist.samples = append(goldHist.samples, cur)
		goldHist.lastT = state.GameTime
	}
	hist := append([]GoldSample(nil), goldHist.samples...)
	if last := hist[len(hist)-1]; cur.T > last.T {
		hist = append(hist, cur) // point courant pour que la courbe soit à jour
	}
	state.Gold = &GoldInfo{Estimated: true, Roles: metas, History: hist}
}

var liveCache = struct {
	mu    sync.Mutex
	state LiveState
	at    time.Time
}{}

func champBySlug(raw string) (champIndexItem, bool) {
	slug := raw
	if i := strings.LastIndex(raw, "_"); i >= 0 {
		slug = raw[i+1:]
	}
	if ensureIndex() != nil {
		return champIndexItem{}, false
	}
	dragon.mu.Lock()
	defer dragon.mu.Unlock()
	for _, c := range dragon.index {
		if strings.EqualFold(c.ID, slug) {
			return c, true
		}
	}
	return champIndexItem{}, false
}

func buildLivePlayer(p rawLivePlayer, haste map[int]itemHaste) LivePlayer {
	out := LivePlayer{
		RiotID:   p.RiotID,
		Champion: p.ChampionName,
		Position: p.Position,
		Level:    p.Level,
		IsDead:   p.IsDead,
		Respawn:  p.RespawnTimer,
		Kills:    p.Scores.Kills,
		Deaths:   p.Scores.Deaths,
		Assists:  p.Scores.Assists,
		CS:       p.Scores.CreepScore,
	}
	if out.RiotID == "" {
		out.RiotID = p.SummonerName
	}
	if out.RiotID == "" { // bots : ni riotId ni summonerName; le champion est unique par partie
		out.RiotID = p.RawChampionName
	}
	ah, uh, sh := 0, 0, 0
	for _, it := range p.Items {
		out.Items = append(out.Items, it.ItemID)
		if h, ok := haste[it.ItemID]; ok {
			ah += h.Ability
			uh += h.Ultimate
			sh += h.Summoner
		}
	}
	out.AbilityHaste, out.SummHaste = ah, sh

	if base, ok := champBySlug(p.RawChampionName); ok {
		out.Key, _ = strconv.Atoi(base.Key)
		out.Icon = base.Image.Full
		if detail, err := getDetail(base.ID); err == nil {
			letters := []string{"Q", "W", "E", "R"}
			for i, s := range detail.Spells {
				if i >= len(letters) {
					break
				}
				sp := LiveSpell{Spell: letters[i], Name: s.Name, Icon: spellIconPath(s.Image.Full)}
				cds := parseCDBurn(s.CooldownBurn)
				isUlt := letters[i] == "R"
				rank := 0
				if isUlt {
					rank = ultRank(p.Level)
					sp.Rank = strconv.Itoa(rank)
				} else {
					rank = maxBasicRank(p.Level)
					sp.Rank = "≤" + strconv.Itoa(rank)
					sp.Est = true
				}
				if rank > 0 && len(cds) > 0 {
					idx := rank - 1
					if idx >= len(cds) {
						idx = len(cds) - 1
					}
					h := ah
					if isUlt {
						h += uh
					}
					sp.CD = cds[idx] * hasteFactor(h)
				}
				out.Spells = append(out.Spells, sp)
			}
		}
	}

	for _, raw := range []rawLiveSummSpell{p.SummonerSpells.One, p.SummonerSpells.Two} {
		slug := summonerSlug(raw)
		if slug == "" && raw.DisplayName == "" {
			continue
		}
		cd := summonerBaseCD[slug] * hasteFactor(sh)
		out.Summoners = append(out.Summoners, LiveSummoner{Slug: slug, Name: raw.DisplayName, CD: cd})
	}
	return out
}

// Timings Failles de l'invocateur, saison 2026 (patch 26.1 : Atakhan retiré,
// Baron ramené à 20:00 · patch 26.11 : larves à 8:00 sans respawn
// · patch 26.12 : Héraut à 15:00).
const (
	dragonFirstSpawn = 300.0
	dragonRespawn    = 300.0
	elderRespawn     = 360.0
	grubsSpawn       = 480.0  // 8:00, un seul spawn par partie
	grubsDespawn     = 885.0  // 14:45 si personne ne les touche
	heraldSpawn      = 900.0  // 15:00
	heraldDespawn    = 1185.0 // 19:45
	baronFirstSpawn  = 1200.0 // 20:00
	baronRespawn     = 360.0
)

// Anticipations d'alerte : 2 min pour se regrouper avant un teamfight d'objectif,
// 70 s pour lâcher sa lane et arriver sur les larves.
const (
	leadTeamfight = 120.0
	leadGrubs     = 70.0
)

func playerTeam(players []rawLivePlayer, name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	base := name
	if i := strings.Index(base, "#"); i >= 0 {
		base = strings.TrimSpace(base[:i])
	}
	for _, p := range players {
		if strings.EqualFold(p.SummonerName, name) || strings.EqualFold(p.RiotIDGameName, base) ||
			strings.EqualFold(p.RiotID, name) || strings.EqualFold(p.SummonerName, base) {
			return p.Team
		}
	}
	return ""
}

// Les larves n'ont pas d'événement documenté et stable : selon les versions le
// kill remonte en "HordeKill" (nom interne du camp) ou en "VoidgrubKill".
func isGrubEvent(name string) bool {
	n := strings.ToLower(name)
	return strings.Contains(n, "horde") || strings.Contains(n, "grub")
}

func buildObjectives(raw *rawLive) map[string]LiveObjective {
	if raw.GameData.MapNumber != 11 && raw.GameData.GameMode != "CLASSIC" && raw.GameData.GameMode != "PRACTICETOOL" {
		return nil
	}
	t := raw.GameData.GameTime
	dragOrder, dragChaos := 0, 0
	lastDragon, lastBaron := 0.0, 0.0
	heraldDead, grubsKilled := false, 0
	for _, e := range raw.Events.Events {
		switch {
		case e.EventName == "DragonKill":
			lastDragon = e.EventTime
			switch playerTeam(raw.AllPlayers, e.KillerName) {
			case "ORDER":
				dragOrder++
			case "CHAOS":
				dragChaos++
			}
		case e.EventName == "HeraldKill":
			heraldDead = true
		case e.EventName == "BaronKill":
			lastBaron = e.EventTime
		case isGrubEvent(e.EventName):
			grubsKilled++
		}
	}
	obj := map[string]LiveObjective{}

	// Larves : un seul spawn, camp de 3, disparition à 14:45.
	grubs := LiveObjective{NextAt: grubsSpawn, Label: "Larves", Leads: []float64{leadTeamfight, leadGrubs}, Rank: 1}
	switch {
	case grubsKilled >= 3 || (t > grubsDespawn && grubsKilled == 0):
		grubs.Gone, grubs.Leads = true, nil
	case grubsKilled > 0:
		grubs.Note = fmt.Sprintf("×%d restantes", 3-grubsKilled)
		grubs.Leads = nil
	case t >= grubsSpawn:
		grubs.Note = fmt.Sprintf("part à %s", clock(grubsDespawn))
	}
	obj["grubs"] = grubs

	// Drake : le 4e d'une équipe donne l'âme, ensuite c'est l'Ancestral.
	elder := dragOrder >= 4 || dragChaos >= 4
	next := dragonFirstSpawn
	if lastDragon > 0 {
		if elder {
			next = lastDragon + elderRespawn
		} else {
			next = lastDragon + dragonRespawn
		}
	}
	dragon := LiveObjective{NextAt: next, Order: dragOrder, Chaos: dragChaos, Label: "Drake",
		Leads: []float64{leadTeamfight}, Rank: 2}
	switch {
	case elder:
		dragon.Label, dragon.Note = "Ancestral", "fight décisif"
	case dragOrder == 3 || dragChaos == 3:
		dragon.Note = "ÂME"
	}
	obj["dragon"] = dragon

	// Héraut : pas d'alerte teamfight (objectif de siège, souvent solo).
	if !heraldDead && t < heraldDespawn {
		herald := LiveObjective{NextAt: heraldSpawn, Label: "Héraut", Rank: 3}
		if t >= heraldSpawn {
			herald.Note = "part à " + clock(heraldDespawn)
		}
		obj["herald"] = herald
	}

	baronNext := baronFirstSpawn
	if lastBaron > 0 {
		baronNext = lastBaron + baronRespawn
	}
	obj["baron"] = LiveObjective{NextAt: baronNext, Label: "Baron", Leads: []float64{leadTeamfight}, Rank: 4}
	return obj
}

func clock(t float64) string {
	s := int(t + 0.5)
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}

func getLiveState() LiveState {
	liveCache.mu.Lock()
	if time.Since(liveCache.at) < 900*time.Millisecond && !liveCache.at.IsZero() {
		s := liveCache.state
		liveCache.mu.Unlock()
		return s
	}
	liveCache.mu.Unlock()

	state := LiveState{}
	req, _ := http.NewRequest(http.MethodGet, liveURL, nil)
	res, err := liveHTTP.Do(req)
	if err == nil && res.StatusCode == 200 {
		var raw rawLive
		if json.NewDecoder(res.Body).Decode(&raw) == nil && len(raw.AllPlayers) > 0 {
			state.Active = true
			state.GameTime = raw.GameData.GameTime
			state.GameMode = raw.GameData.GameMode

			myTeam := ""
			for _, p := range raw.AllPlayers {
				if (raw.ActivePlayer.RiotID != "" && strings.EqualFold(p.RiotID, raw.ActivePlayer.RiotID)) ||
					(raw.ActivePlayer.SummonerName != "" && strings.EqualFold(p.SummonerName, raw.ActivePlayer.SummonerName)) {
					myTeam = p.Team
					break
				}
			}
			if myTeam == "" {
				myTeam = "ORDER" // spectateur : ORDER affiché comme "alliés"
			}
			haste := itemHasteTable()
			for _, p := range raw.AllPlayers {
				lp := buildLivePlayer(p, haste)
				lp.IsMe = (raw.ActivePlayer.RiotID != "" && strings.EqualFold(p.RiotID, raw.ActivePlayer.RiotID)) ||
					(raw.ActivePlayer.SummonerName != "" && strings.EqualFold(p.SummonerName, raw.ActivePlayer.SummonerName))
				if p.Team == myTeam {
					state.Allies = append(state.Allies, lp)
				} else {
					state.Enemies = append(state.Enemies, lp)
				}
			}
			state.Objectives = buildObjectives(&raw)
			updateGold(&state)
		}
	}
	if res != nil {
		res.Body.Close()
	}

	liveCache.mu.Lock()
	liveCache.state, liveCache.at = state, time.Now()
	liveCache.mu.Unlock()
	return state
}

func apiLive(w http.ResponseWriter, r *http.Request) { writeJSON(w, getLiveState()) }
