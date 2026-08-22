package main

// Live Client Data API (Riot, locale, officielle) : https://127.0.0.1:2999/liveclientdata/allgamedata
// Disponible uniquement pendant une partie, en lecture seule.
// On en tire : niveaux, items (=> ability haste), morts/respawns, sorts d'invocateur,
// runes (keystone + arbre secondaire), événements (drakes / héraut / baron)
// — pour tous les joueurs de la partie.

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math"
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
	Runes rawLiveRunes `json:"runes"`
	Team  string       `json:"team"`
}

type rawLiveRune struct {
	DisplayName string `json:"displayName"`
	ID          int    `json:"id"`
}

type rawLiveRunes struct {
	Keystone          rawLiveRune `json:"keystone"`
	PrimaryRuneTree   rawLiveRune `json:"primaryRuneTree"`
	SecondaryRuneTree rawLiveRune `json:"secondaryRuneTree"`
}

type rawLiveFullRunes struct {
	Keystone          rawLiveRune   `json:"keystone"`
	PrimaryRuneTree   rawLiveRune   `json:"primaryRuneTree"`
	SecondaryRuneTree rawLiveRune   `json:"secondaryRuneTree"`
	GeneralRunes      []rawLiveRune `json:"generalRunes"`
	StatRunes         []rawLiveRune `json:"statRunes"`
}

type rawLive struct {
	ActivePlayer struct {
		RiotID       string           `json:"riotId"`
		SummonerName string           `json:"summonerName"`
		FullRunes    rawLiveFullRunes `json:"fullRunes"`
	} `json:"activePlayer"`
	AllPlayers []rawLivePlayer `json:"allPlayers"`
	Events     struct {
		Events []struct {
			EventName  string  `json:"EventName"`
			EventTime  float64 `json:"EventTime"`
			DragonType string  `json:"DragonType"`
			KillerName string  `json:"KillerName"`
			VictimName string  `json:"VictimName"`
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
	if itemHasteCache.version == version && itemHasteCache.data != nil {
		d := itemHasteCache.data
		itemHasteCache.mu.Unlock()
		return d
	}
	itemHasteCache.mu.Unlock()

	var payload struct {
		Data map[string]struct {
			Description string `json:"description"`
		} `json:"data"`
	}
	url := fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/item.json", version)
	if err := jsonGET(url, &payload); err != nil {
		itemHasteCache.mu.Lock()
		defer itemHasteCache.mu.Unlock()
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
	itemHasteCache.mu.Lock()
	itemHasteCache.version, itemHasteCache.data = version, table
	itemHasteCache.mu.Unlock()
	return table
}

// ---------- Runes (keystone + arbre secondaire) ----------

type runeInfo struct {
	Name string
	Icon string
}

var runeCache = struct {
	mu      sync.Mutex
	version string
	byID    map[int]runeInfo
}{}

type runeReforgedTree struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Slots []struct {
		Runes []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
			Icon string `json:"icon"`
		} `json:"runes"`
	} `json:"slots"`
}

func runeTable() map[int]runeInfo {
	version := getVersion()
	runeCache.mu.Lock()
	if runeCache.version == version && runeCache.byID != nil {
		d := runeCache.byID
		runeCache.mu.Unlock()
		return d
	}
	runeCache.mu.Unlock()

	var trees []runeReforgedTree
	url := fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/fr_FR/runesReforged.json", version)
	if err := jsonGET(url, &trees); err != nil || len(trees) == 0 {
		trees = nil
		url = fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/runesReforged.json", version)
		if err := jsonGET(url, &trees); err != nil || len(trees) == 0 {
			runeCache.mu.Lock()
			defer runeCache.mu.Unlock()
			if runeCache.byID != nil {
				return runeCache.byID
			}
			return map[int]runeInfo{}
		}
	}
	byID := make(map[int]runeInfo, 80)
	for _, t := range trees {
		byID[t.ID] = runeInfo{Name: t.Name, Icon: t.Icon}
		for _, slot := range t.Slots {
			for _, r := range slot.Runes {
				byID[r.ID] = runeInfo{Name: r.Name, Icon: r.Icon}
			}
		}
	}
	runeCache.mu.Lock()
	runeCache.version, runeCache.byID = version, byID
	runeCache.mu.Unlock()
	return byID
}

func liveRuneOf(raw rawLiveRune) LiveRune {
	if raw.ID == 0 {
		return LiveRune{}
	}
	info := runeTable()[raw.ID]
	name := raw.DisplayName
	if name == "" {
		name = info.Name
	}
	return LiveRune{ID: raw.ID, Name: name, Icon: info.Icon}
}

const (
	runeCosmicInsight = 8347
	runeTranscendence = 8210
	runeHexflash      = 8306
	runeStatAH        = 5007
)

var cdRuneIDs = map[int]bool{
	runeCosmicInsight: true,
	runeTranscendence: true,
	runeHexflash:      true,
}

func runeIDSet(runes []rawLiveRune) map[int]bool {
	out := make(map[int]bool, len(runes))
	for _, r := range runes {
		if r.ID != 0 {
			out[r.ID] = true
		}
	}
	return out
}

func transcendenceAH(level int) int {
	n := 0
	if level >= 5 {
		n += 5
	}
	if level >= 8 {
		n += 5
	}
	return n
}

func statShardAH(level int) int {
	if level < 1 {
		return 0
	}
	if level > 18 {
		level = 18
	}
	return int(math.Round(8.0 * float64(level) / 18.0))
}

func liveRunesOf(p rawLivePlayer, full *rawLiveFullRunes) (LiveRunes, bool, bool, bool) {
	src := p.Runes
	ids := map[int]bool{}
	known := false
	if full != nil && (full.Keystone.ID != 0 || len(full.GeneralRunes) > 0) {
		known = true
		src = rawLiveRunes{
			Keystone:          full.Keystone,
			PrimaryRuneTree:   full.PrimaryRuneTree,
			SecondaryRuneTree: full.SecondaryRuneTree,
		}
		ids = runeIDSet(full.GeneralRunes)
		for _, r := range full.StatRunes {
			if r.ID != 0 {
				ids[r.ID] = true
			}
		}
	}
	out := LiveRunes{
		Keystone:  liveRuneOf(src.Keystone),
		Primary:   liveRuneOf(src.PrimaryRuneTree),
		Secondary: liveRuneOf(src.SecondaryRuneTree),
	}
	if known {
		seen := map[int]bool{out.Keystone.ID: true}
		for _, r := range full.GeneralRunes {
			if !cdRuneIDs[r.ID] || seen[r.ID] {
				continue
			}
			seen[r.ID] = true
			if info := liveRuneOf(r); info.ID != 0 {
				out.Major = append(out.Major, info)
			}
		}
	}
	return out, known && ids[runeCosmicInsight], known && ids[runeTranscendence], known && ids[runeStatAH]
}

// ---------- Sorts d'invocateur ----------

var summonerBaseCD = map[string]float64{
	"SummonerFlash":            300,
	"SummonerHexflash":         20, // Hextech Flashtraption, après le Flash
	"SummonerTeleport":         360,
	"SummonerHaste":            210, // Ghost
	"SummonerHeal":             240,
	"SummonerBarrier":          180,
	"SummonerExhaust":          210,
	"SummonerDot":              180, // Ignite
	"SummonerBoost":            210, // Cleanse
	"SummonerSmite":            90,  // recharge d'une charge (le 15 s est l'inter-cast)
	"SummonerSmiteUnleashed":   90,
	"SummonerSmitePrimal":      90,
	"SummonerSmiteChallenging": 90,
	"SummonerSmiteChilling":    90,
	"SummonerMana":             240, // Clarity (ARAM)
	"SummonerSnowball":         80,  // Mark (ARAM)
}

var summSlugRe = regexp.MustCompile(`SummonerSpell_(\w+?)_`)

func summonerSlug(s rawLiveSummSpell) string {
	if m := summSlugRe.FindStringSubmatch(s.RawDisplayName); m != nil {
		return m[1]
	}
	return ""
}

func blobOf(s rawLiveSummSpell, slug string) string {
	return strings.ToLower(strings.TrimSpace(slug + " " + s.DisplayName + " " + s.RawDisplayName))
}

func resolveSummoner(raw rawLiveSummSpell, sh int) LiveSummoner {
	slug := summonerSlug(raw)
	blob := blobOf(raw, slug)
	name := strings.TrimSpace(raw.DisplayName)
	out := LiveSummoner{Slug: slug, Name: name, Icon: summonerIcon(slug)}

	switch {
	case strings.Contains(blob, "hexflash") || strings.Contains(blob, "hextechflash") ||
		strings.Contains(blob, "flashtraption") || strings.Contains(blob, "flashperkshex"):
		out.Slug, out.Kind = "SummonerHexflash", "hexflash"
		if out.Name == "" || strings.EqualFold(out.Name, "flash") {
			out.Name = "Hexflash"
		}
		out.Icon = "perk-images/Styles/Inspiration/HextechFlashtraption/HextechFlashtraption.png"
	case strings.Contains(blob, "smite") || strings.Contains(blob, "châtiment") || strings.Contains(blob, "chatiment"):
		out.Kind = "smite"
		switch {
		case strings.Contains(blob, "unleashed") || strings.Contains(blob, "déchaîn") || strings.Contains(blob, "dechain"):
			out.Slug = "SummonerSmiteUnleashed"
			if out.Name == "" {
				out.Name = "Smite déchaîné"
			}
		case strings.Contains(blob, "primordial") || strings.Contains(blob, "primal"):
			out.Slug = "SummonerSmitePrimal"
			if out.Name == "" {
				out.Name = "Smite primal"
			}
		case strings.Contains(blob, "challenging") || strings.Contains(blob, "ravageur") || strings.Contains(blob, "duel"):
			out.Slug = "SummonerSmiteChallenging"
			if out.Name == "" {
				out.Name = "Smite ravageur"
			}
		case strings.Contains(blob, "chilling") || strings.Contains(blob, "ganker"):
			out.Slug = "SummonerSmiteChilling"
			if out.Name == "" {
				out.Name = "Smite glacial"
			}
		default:
			out.Slug = "SummonerSmite"
			if out.Name == "" {
				out.Name = "Smite"
			}
		}
		out.Icon = "spell/SummonerSmite.png"
	case strings.Contains(blob, "flash"):
		out.Slug, out.Kind = "SummonerFlash", "flash"
		if out.Name == "" {
			out.Name = "Flash"
		}
		out.Icon = "spell/SummonerFlash.png"
	}

	if out.Slug == "" && name != "" {
		out.Slug = "SummonerUnknown"
	}
	cd, ok := summonerBaseCD[out.Slug]
	if !ok && out.Kind == "smite" {
		cd = 90
	}
	if !ok && out.Kind == "hexflash" {
		cd = 20
	}
	if cd > 0 {
		out.CD = cd * hasteFactor(sh)
	}
	if out.Icon == "" {
		out.Icon = summonerIcon(out.Slug)
	}
	return out
}

func summonerIcon(slug string) string {
	if slug == "" {
		return ""
	}
	base := slug
	if i := strings.Index(base, "_Jade"); i > 0 {
		base = base[:i]
	}
	switch {
	case strings.HasPrefix(base, "SummonerSmite"):
		return "spell/SummonerSmite.png"
	case strings.Contains(strings.ToLower(base), "hexflash"):
		return "perk-images/Styles/Inspiration/HextechFlashtraption/HextechFlashtraption.png"
	case strings.HasPrefix(base, "SummonerFlash"):
		return "spell/SummonerFlash.png"
	}
	return "spell/" + base + ".png"
}

func isFlashKind(kind, slug string) bool {
	if kind == "flash" || kind == "hexflash" {
		return true
	}
	s := strings.ToLower(slug)
	return strings.Contains(s, "flash") || strings.Contains(s, "hexflash")
}

// Durée de mort Faille (wiki) : BRW(niveau) × (1 + TIF(temps)).
var deathBRW = []float64{
	10, 10, 12, 12, 14, 16, 20, 25, 28, 32.5, 35, 37.5, 40, 42.5, 45, 47.5, 50, 52.5,
}

func deathTIF(gameTime float64) float64 {
	min := gameTime / 60
	if min < 15 {
		return 0
	}
	tif := 0.0
	if min >= 15 {
		steps := math.Ceil(2 * (math.Min(min, 30) - 15))
		tif += steps * 0.00425
	}
	if min >= 30 {
		steps := math.Ceil(2 * (math.Min(min, 45) - 30))
		tif += steps * 0.003
	}
	if min >= 45 {
		steps := math.Ceil(2 * (math.Min(min, 55) - 45))
		tif += steps * 0.0145
	}
	if tif > 0.5 {
		tif = 0.5
	}
	return tif
}

func deathDuration(level int, gameTime float64, mode string) float64 {
	if level < 1 {
		level = 1
	}
	if level > 18 {
		level = 18
	}
	if strings.EqualFold(mode, "ARAM") {
		// 16 paliers (niv. 3–18) ; on borne le niv. 1–2 au premier.
		aram := []float64{11, 13, 15, 17, 19, 21, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40}
		i := level - 3
		if i < 0 {
			i = 0
		}
		if i >= len(aram) {
			i = len(aram) - 1
		}
		return aram[i]
	}
	br := deathBRW[level-1]
	return br * (1 + deathTIF(gameTime))
}

var deathClock = struct {
	mu  sync.Mutex
	key string
	at  map[string]float64
}{at: map[string]float64{}}

func deathPlayerKey(p rawLivePlayer) string {
	id := p.RiotID
	if id == "" {
		id = p.SummonerName
	}
	if id == "" {
		id = p.RawChampionName
	}
	return id + "|" + p.RawChampionName
}

func resetDeathClock(gameKey string) {
	deathClock.mu.Lock()
	defer deathClock.mu.Unlock()
	if deathClock.key == gameKey {
		return
	}
	deathClock.key = gameKey
	deathClock.at = map[string]float64{}
}

func noteDeathAt(key string, isDead bool, remaining, gameTime float64, level int, mode string) float64 {
	deathClock.mu.Lock()
	defer deathClock.mu.Unlock()
	if deathClock.at == nil {
		deathClock.at = map[string]float64{}
	}
	if !isDead || remaining <= 0.05 {
		delete(deathClock.at, key)
		return 0
	}
	if at, ok := deathClock.at[key]; ok {
		return at
	}
	dur := deathDuration(level, gameTime, mode)
	elapsed := dur - remaining
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed > dur {
		elapsed = dur
	}
	at := gameTime - elapsed
	deathClock.at[key] = at
	return at
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
	Desc  string  `json:"desc,omitempty"`
	Cost  string  `json:"cost,omitempty"`
	Range string  `json:"range,omitempty"`
	CD    float64 `json:"cd"`   // secondes, haste incluse, 0 = inconnu
	Rank  string  `json:"rank"` // "2" exact (R) ou "≤4" plausible
	Est   bool    `json:"est"`  // true si le rang est estimé
}

type LiveSummoner struct {
	Slug string  `json:"slug"`
	Name string  `json:"name"`
	Icon string  `json:"icon,omitempty"` // chemin relatif ddragon / perk-images
	Kind string  `json:"kind,omitempty"` // flash, hexflash, smite
	CD   float64 `json:"cd"`             // secondes, haste invocateur (items + CI) incluse
}

// LiveRune : keystone ou arbre (secondaire). Icon = chemin ddragon sans version
// (perk-images/Styles/…), servi via /ddragon/img/…
type LiveRune struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Icon string `json:"icon,omitempty"`
}

type LiveRunes struct {
	Keystone  LiveRune   `json:"keystone,omitempty"`
	Primary   LiveRune   `json:"primary,omitempty"`
	Secondary LiveRune   `json:"secondary,omitempty"`
	Major     []LiveRune `json:"major,omitempty"` // CI / Transcendance / Hexflash
}

type LivePlayer struct {
	RiotID        string         `json:"riotId"`
	Name          string         `json:"name"` // pseudo affiché ; champion si streamer mode
	Hidden        bool           `json:"hidden,omitempty"`
	Rank          string         `json:"rank,omitempty"` // "D2", "Master 120"
	Tier          string         `json:"tier,omitempty"` // "diamond" pour la bordure
	WR            *int           `json:"wr,omitempty"`   // winrate solo 0–100
	Games         int            `json:"games,omitempty"`
	AI            *int           `json:"ai,omitempty"` // DeepLoL AI-score (10 last)
	Tags          []ScoutTag     `json:"tags,omitempty"`
	DeepLoL       string         `json:"deeplol,omitempty"`
	IsMe          bool           `json:"isMe,omitempty"`
	Champion      string         `json:"champion"`
	Key           int            `json:"key"`
	Icon          string         `json:"icon"`
	Position      string         `json:"position"`
	Level         int            `json:"level"`
	IsDead        bool           `json:"isDead"`
	Respawn       float64        `json:"respawn"`
	Kills         int            `json:"kills"`
	Deaths        int            `json:"deaths"`
	Assists       int            `json:"assists"`
	CS            int            `json:"cs"`
	AbilityHaste  int            `json:"abilityHaste"`
	SummHaste     int            `json:"summHaste"`
	CosmicInsight bool           `json:"cosmicInsight,omitempty"`
	CosmicKnown   bool           `json:"cosmicKnown,omitempty"` // page complète (joueur local)
	DeathAt       float64        `json:"deathAt,omitempty"`     // gameTime du moment de la mort
	Items         []int          `json:"items"`
	ItemGold      int            `json:"itemGold,omitempty"` // somme ddragon gold.total
	Spells        []LiveSpell    `json:"spells"`
	Summoners     []LiveSummoner `json:"summoners"`
	Runes         LiveRunes      `json:"runes,omitempty"`
	Buffs         []LiveObjBuff  `json:"buffs,omitempty"` // Nash / Ancestral encore actifs
}

// LiveObjBuff : Hand of Baron (180 s) ou Aspect of the Dragon (150 s).
// Perdu à la mort ; pas donné si le joueur était déjà mort au take / steal.
type LiveObjBuff struct {
	Kind  string  `json:"kind"`  // "baron" | "elder"
	Until float64 `json:"until"` // gameTime de fin
}

type LiveObjective struct {
	NextAt float64   `json:"nextAt"` // temps de jeu (s) du prochain spawn, 0 = aucun
	Label  string    `json:"label"`
	Order  int       `json:"order,omitempty"`
	Chaos  int       `json:"chaos,omitempty"`
	Note   string    `json:"note,omitempty"`  // "ÂME", "×2 restantes"…
	Kind   string    `json:"kind,omitempty"`  // infernal, mountain, ocean, cloud, hextech, chemtech, soul, elder
	Leads  []float64 `json:"leads,omitempty"` // anticipations d'alerte (s avant le spawn)
	Gone   bool      `json:"gone,omitempty"`  // objectif définitivement parti (despawn / camp fini)
	Rank   int       `json:"rank,omitempty"`  // ordre d'affichage dans la barre
}

type LiveState struct {
	Active     bool                     `json:"active"`
	Demo       bool                     `json:"demo,omitempty"`
	GameTime   float64                  `json:"gameTime"`
	GameMode   string                   `json:"gameMode"`
	Enemies    []LivePlayer             `json:"enemies"`
	Allies     []LivePlayer             `json:"allies"`
	Objectives map[string]LiveObjective `json:"objectives"`
	Gold       *GoldInfo                `json:"gold,omitempty"`
	Voices     []VoiceCue               `json:"voices,omitempty"`
	HudOpen    bool                     `json:"hudOpen,omitempty"`
	Side       string                   `json:"side,omitempty"` // "blue" (ORDER) | "red" (CHAOS)
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

var liveSide = struct {
	mu sync.Mutex
	s  string
}{s: "blue"}

func rememberLiveSide(side string) {
	side = hudNormSide(side)
	liveSide.mu.Lock()
	liveSide.s = side
	liveSide.mu.Unlock()
}

func lastLiveSide() string {
	liveSide.mu.Lock()
	defer liveSide.mu.Unlock()
	if liveSide.s == "" {
		return "blue"
	}
	return liveSide.s
}

func liveGameActive() bool {
	liveCache.mu.Lock()
	on := liveCache.state.Active
	liveCache.mu.Unlock()
	return on || demoLiveOn()
}

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

func buildLivePlayer(p rawLivePlayer, haste map[int]itemHaste, gameTime float64, mode string, full *rawLiveFullRunes) LivePlayer {
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
	runes, cosmic, trans, shardAH := liveRunesOf(p, full)
	out.Runes = runes
	out.CosmicKnown = full != nil && (full.Keystone.ID != 0 || len(full.GeneralRunes) > 0)
	if cosmic {
		sh += 18
		out.CosmicInsight = true
	}
	if trans {
		ah += transcendenceAH(p.Level)
	}
	if shardAH {
		ah += statShardAH(p.Level)
	}
	out.AbilityHaste, out.SummHaste = ah, sh
	out.DeathAt = noteDeathAt(deathPlayerKey(p), p.IsDead, p.RespawnTimer, gameTime, p.Level, mode)

	if base, ok := champBySlug(p.RawChampionName); ok {
		out.Key, _ = strconv.Atoi(base.Key)
		out.Icon = base.Image.Full
		if detail, err := getDetail(base.ID); err == nil {
			if f := detail.Passive.Image.Full; f != "" || detail.Passive.Name != "" {
				icon := ""
				if f != "" {
					icon = "passive/" + f
				}
				ps := livePassiveSpell(
					base.ID, detail.Passive.Name, icon, plainText(detail.Passive.Description), p.Level, ah,
				)
				ps.Desc = ""
				out.Spells = append(out.Spells, ps)
			}
			letters := []string{"Q", "W", "E", "R"}
			for i, s := range detail.Spells {
				if i >= len(letters) {
					break
				}
				sp := LiveSpell{
					Spell: letters[i], Name: s.Name, Icon: spellIconPath(s.Image.Full),
					Cost: burn(s.CostBurn), Range: burn(s.RangeBurn),
				}
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
		if raw.DisplayName == "" && raw.RawDisplayName == "" {
			continue
		}
		out.Summoners = append(out.Summoners, resolveSummoner(raw, sh))
	}
	applyIdentity(&out)
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
	baronBuffDur     = 180.0 // Hand of Baron
	elderBuffDur     = 150.0 // Aspect of the Dragon (perdu à la mort)
)

// Anticipations d'alerte : 1 min 15 pour se regrouper avant un objectif.
const leadTeamfight = 75.0

func liveNameMatch(p rawLivePlayer, name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	base := name
	if i := strings.Index(base, "#"); i >= 0 {
		base = strings.TrimSpace(base[:i])
	}
	return strings.EqualFold(p.SummonerName, name) || strings.EqualFold(p.SummonerName, base) ||
		strings.EqualFold(p.RiotIDGameName, name) || strings.EqualFold(p.RiotIDGameName, base) ||
		strings.EqualFold(p.RiotID, name) || strings.EqualFold(p.RiotID, base) ||
		strings.EqualFold(p.ChampionName, name) || strings.EqualFold(p.ChampionName, base)
}

func playerTeam(players []rawLivePlayer, name string) string {
	for _, p := range players {
		if liveNameMatch(p, name) {
			return p.Team
		}
	}
	return ""
}

// Replay des events : le buff part sur l'équipe du killer, seulement si le
// joueur est vivant à cet instant. Une mort (ChampionKill) le coupe ; un
// respawn ne le rend pas. p.IsDead actuel masque aussi (cadavre / poll).
func objBuffsFor(raw *rawLive, p rawLivePlayer) []LiveObjBuff {
	if raw == nil {
		return nil
	}
	if raw.GameData.MapNumber != 11 && raw.GameData.GameMode != "CLASSIC" && raw.GameData.GameMode != "PRACTICETOOL" {
		return nil
	}
	now := raw.GameData.GameTime
	mode := raw.GameData.GameMode
	alive := true
	deadUntil := 0.0
	var baronUntil, elderUntil float64
	for _, e := range raw.Events.Events {
		t := e.EventTime
		if !alive && t >= deadUntil {
			alive = true
		}
		switch {
		case e.EventName == "ChampionKill" && liveNameMatch(p, e.VictimName):
			alive = false
			deadUntil = t + deathDuration(p.Level, t, mode)
			baronUntil, elderUntil = 0, 0
		case e.EventName == "BaronKill":
			if alive && playerTeam(raw.AllPlayers, e.KillerName) == p.Team {
				baronUntil = t + baronBuffDur
			}
		case e.EventName == "DragonKill" && dragonKindFromType(e.DragonType) == "elder":
			if alive && playerTeam(raw.AllPlayers, e.KillerName) == p.Team {
				elderUntil = t + elderBuffDur
			}
		}
	}
	if p.IsDead {
		return nil
	}
	var out []LiveObjBuff
	if baronUntil > now {
		out = append(out, LiveObjBuff{Kind: "baron", Until: baronUntil})
	}
	if elderUntil > now {
		out = append(out, LiveObjBuff{Kind: "elder", Until: elderUntil})
	}
	return out
}

// Les larves n'ont pas d'événement documenté et stable : selon les versions le
// kill remonte en "HordeKill" (nom interne du camp) ou en "VoidgrubKill".
func isGrubEvent(name string) bool {
	n := strings.ToLower(name)
	return strings.Contains(n, "horde") || strings.Contains(n, "grub")
}

// DragonType du live client : Fire/Earth/Water/Air/Hextech/Chemtech/Elder
// (parfois déjà Infernal/Mountain/…).
func dragonKindFromType(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "fire", "infernal":
		return "infernal"
	case "earth", "mountain":
		return "mountain"
	case "water", "ocean":
		return "ocean"
	case "air", "cloud":
		return "cloud"
	case "hextech":
		return "hextech"
	case "chemtech":
		return "chemtech"
	case "elder", "elderdragon":
		return "elder"
	default:
		return ""
	}
}

func buildObjectives(raw *rawLive) map[string]LiveObjective {
	if raw.GameData.MapNumber != 11 && raw.GameData.GameMode != "CLASSIC" && raw.GameData.GameMode != "PRACTICETOOL" {
		return nil
	}
	t := raw.GameData.GameTime
	dragOrder, dragChaos := 0, 0
	lastDragon, lastBaron := 0.0, 0.0
	heraldDead, grubsKilled := false, 0
	elem := ""
	for _, e := range raw.Events.Events {
		switch {
		case e.EventName == "DragonKill":
			lastDragon = e.EventTime
			if k := dragonKindFromType(e.DragonType); k != "" && k != "elder" {
				elem = k
			}
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
	grubs := LiveObjective{NextAt: grubsSpawn, Label: "Larves", Leads: []float64{leadTeamfight}, Rank: 1}
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
		Leads: []float64{leadTeamfight}, Rank: 2, Kind: "dragon"}
	switch {
	case elder:
		dragon.Label, dragon.Note, dragon.Kind = "Ancestral", "fight décisif", "elder"
	case dragOrder == 3 || dragChaos == 3:
		dragon.Note, dragon.Kind = "ÂME", "soul"
	case elem != "":
		dragon.Kind = elem
	}
	obj["dragon"] = dragon

	// Héraut : pas d'alerte teamfight (objectif de siège, souvent solo).
	if !heraldDead && t < heraldDespawn {
		herald := LiveObjective{NextAt: heraldSpawn, Label: "Héraut", Leads: []float64{leadTeamfight}, Rank: 3}
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

	state := LiveState{Side: lastLiveSide()}
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
			state.Side = hudTeamSide(myTeam)
			var full *rawLiveFullRunes
			if raw.ActivePlayer.FullRunes.Keystone.ID != 0 || len(raw.ActivePlayer.FullRunes.GeneralRunes) > 0 {
				fr := raw.ActivePlayer.FullRunes
				full = &fr
			}
			haste := itemHasteTable()
			var slugs []string
			for _, p := range raw.AllPlayers {
				if base, ok := champBySlug(p.RawChampionName); ok {
					slugs = append(slugs, base.ID)
				}
			}
			prefetchPassiveCDs(slugs)
			gkey := raw.GameData.GameMode
			for _, p := range raw.AllPlayers {
				gkey += "|" + p.RawChampionName
			}
			resetDeathClock(gkey)
			for _, p := range raw.AllPlayers {
				var fr *rawLiveFullRunes
				if full != nil && ((raw.ActivePlayer.RiotID != "" && strings.EqualFold(p.RiotID, raw.ActivePlayer.RiotID)) ||
					(raw.ActivePlayer.SummonerName != "" && strings.EqualFold(p.SummonerName, raw.ActivePlayer.SummonerName))) {
					fr = full
				}
				lp := buildLivePlayer(p, haste, raw.GameData.GameTime, raw.GameData.GameMode, fr)
				lp.Buffs = objBuffsFor(&raw, p)
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
			updateVoiceCues(&state)
			prefetchObjSFX()
			if _, _, open := hudStatus(); open {
				state.HudOpen = true
			}
			if creds, err := getCreds(); err == nil {
				kickHarvest(creds, "InProgress")
			}
		}
	}
	if res != nil {
		res.Body.Close()
	}

	liveCache.mu.Lock()
	liveCache.state, liveCache.at = state, time.Now()
	liveCache.mu.Unlock()
	if state.Active {
		rememberLiveSide(state.Side)
		hudSetSide(state.Side)
	}
	return state
}

func liveGameKey(state LiveState) string {
	if !state.Active {
		return ""
	}
	parts := make([]string, 0, 1+len(state.Enemies)+len(state.Allies))
	parts = append(parts, state.GameMode)
	for _, p := range state.Enemies {
		parts = append(parts, p.Champion)
	}
	for _, p := range state.Allies {
		parts = append(parts, p.Champion)
	}
	return strings.Join(parts, "|")
}

func writeLiveJSON(w http.ResponseWriter, state LiveState) {
	type wire struct {
		LiveState
		Tracks map[string]clientTrack `json:"tracks"`
		CI     []string               `json:"ci"`
	}
	writeJSON(w, wire{LiveState: state, Tracks: hudTracksGet(), CI: hudCIGet()})
}

func apiLive(w http.ResponseWriter, r *http.Request) {
	state := getLiveState()
	if !state.Active {
		if d, ok := demoLiveCopy(); ok {
			if d.Side == "" {
				d.Side = lastLiveSide()
			}
			if _, _, open := hudStatus(); open {
				d.HudOpen = true
			}
			autoOpenHudForGame(liveGameKey(d))
			writeLiveJSON(w, d)
			return
		}
		autoOpenHudForGame("")
		writeLiveJSON(w, state)
		return
	}
	setDemoLive(nil)
	autoOpenHudForGame(liveGameKey(state))
	writeLiveJSON(w, state)
}
