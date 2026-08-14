package main

// Stats DeepLoL (rang, WR solo, AI-score des 10 dernières games) pour le live.
// LCU donne déjà rang + victoires/défaites ; l'AI-score n'existe que chez DeepLoL.
// En partie : un GET /ingame/ingame_info (match_id LCU) pour les 10 joueurs.

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var dlHTTP = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		DialContext:           dialHTTP,
		ForceAttemptHTTP2:     false,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 8 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		MaxIdleConns:          4,
		MaxIdleConnsPerHost:   2,
	},
}

var dlCache struct {
	mu   sync.Mutex
	key  string
	ok   bool
	at   time.Time
	busy bool
}

type dlSolo struct {
	Tier         string          `json:"tier"`
	Division     json.RawMessage `json:"division"`
	LeaguePoints int             `json:"league_points"`
	Wins         int             `json:"wins"`
	Losses       int             `json:"losses"`
}

type dlParticipant struct {
	Side       string `json:"side"`
	Position   string `json:"position"`
	ChampionID int    `json:"champion_id"`
	PUUID      string `json:"puu_id"`
	RiotName   string `json:"riot_id_name"`
	RiotTag    string `json:"riot_id_tag_line"`
	Realtime   struct {
		SeasonTier struct {
			Solo dlSolo `json:"ranked_solo_5x5"`
			Flex dlSolo `json:"ranked_flex_sr"`
		} `json:"season_tier_info_dict"`
	} `json:"summoner_realtime_data"`
	Info struct {
		Summoner dlSummonerInfo `json:"summoner_info_dict"`
	} `json:"participant_info"`
}

type dlSummonerInfo struct {
	AIScoreAvg     float64  `json:"ai_score_avg"`
	AIScoreAvg15   float64  `json:"ai_score_avg_15"`
	AIScoreAvgLoss float64  `json:"ai_score_avg_loss"`
	Wins           int      `json:"wins"`
	Losses         int      `json:"losses"`
	Tag            dlTagBag `json:"tag"`
}

type dlTagBag struct {
	RunawayCount  int  `json:"runaway_count"`
	ComebackUser  bool `json:"comeback_user"`
	WinningStreak int  `json:"wining_streak"` // typo DeepLoL
	LosingStreak  int  `json:"losing_streak"`
}

// ScoutTag = pastille live DeepLoL (Carry, 3L, Comeback…).
type ScoutTag struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Tone  string `json:"tone"` // good | bad | warn
	Hint  string `json:"hint,omitempty"`
}

type dlIngame struct {
	Playing          bool            `json:"playing"`
	Support          bool            `json:"support"`
	ParticipantsList []dlParticipant `json:"participants_list"`
}

func resetDeepLoL() {
	dlCache.mu.Lock()
	dlCache.key, dlCache.ok, dlCache.at, dlCache.busy = "", false, time.Time{}, false
	dlCache.mu.Unlock()
}

func enrichDeepLoL(creds *lcuCreds, phase string) {
	switch phase {
	case "GameStart", "InProgress", "Reconnect", "WaitingForStats":
	default:
		return
	}
	if creds == nil {
		return
	}
	identStore.mu.Lock()
	key := identStore.gameKey
	identStore.mu.Unlock()
	dlCache.mu.Lock()
	if dlCache.ok && dlCache.key == key && key != "" {
		dlCache.mu.Unlock()
		return
	}
	if dlCache.busy || (!dlCache.at.IsZero() && time.Since(dlCache.at) < 8*time.Second) {
		dlCache.mu.Unlock()
		return
	}
	dlCache.busy = true
	dlCache.at = time.Now()
	dlCache.key = key
	dlCache.mu.Unlock()
	defer func() {
		dlCache.mu.Lock()
		dlCache.busy = false
		dlCache.mu.Unlock()
	}()

	matchID := gameIDFromFlow(creds)
	puuid := firstStoredPUUID()
	if matchID == 0 || puuid == "" {
		return
	}
	region := cachedRegion()
	if region == "" {
		region = ensureRegion(creds)
	}
	parts, ok := fetchDeepLoLIngame(puuid, region, matchID)
	if !ok || len(parts) == 0 {
		return
	}
	mergeDeepLoL(parts, region)
	dlCache.mu.Lock()
	dlCache.ok = true
	dlCache.mu.Unlock()
}

func gameIDFromFlow(creds *lcuCreds) int64 {
	var session struct {
		GameData struct {
			GameID json.RawMessage `json:"gameId"`
		} `json:"gameData"`
	}
	if _, err := lcuGET(creds, "/lol-gameflow/v1/session", &session); err != nil {
		return 0
	}
	return parseAnyID(session.GameData.GameID)
}

func firstStoredPUUID() string {
	identStore.mu.Lock()
	defer identStore.mu.Unlock()
	for _, p := range identStore.byPUUID {
		if p != nil && p.PUUID != "" {
			return p.PUUID
		}
	}
	return ""
}

func fetchDeepLoLIngame(puuid, region string, matchID int64) ([]dlParticipant, bool) {
	season := deeplolSeason()
	u, _ := url.Parse("https://b2c-api-cdn.deeplol.gg/ingame/ingame_info")
	q := u.Query()
	q.Set("puu_id", puuid)
	q.Set("platform_id", platformID(region))
	q.Set("season", strconv.Itoa(season))
	q.Set("match_id", strconv.FormatInt(matchID, 10))
	u.RawQuery = q.Encode()
	raw, err := dlGET(u.String())
	if err != nil {
		return nil, false
	}
	ing, err := parseDLIngame(raw)
	if err != nil || !ing.Playing || len(ing.ParticipantsList) == 0 {
		return nil, false
	}
	return ing.ParticipantsList, true
}

func parseDLIngame(raw []byte) (dlIngame, error) {
	raw = unwrapDL(raw)
	var ing dlIngame
	if err := json.Unmarshal(raw, &ing); err != nil {
		return dlIngame{}, err
	}
	return ing, nil
}

func unwrapDL(raw []byte) []byte {
	var wrap struct {
		Data json.RawMessage `json:"data"`
	}
	if json.Unmarshal(raw, &wrap) == nil && len(wrap.Data) > 2 && wrap.Data[0] == '{' {
		return wrap.Data
	}
	return raw
}

func mergeDeepLoL(parts []dlParticipant, region string) {
	ids := make([]*playerIdent, 0, len(parts))
	for _, part := range parts {
		p := lookupDLIdent(part)
		if p == nil {
			p = identFromDL(part, region)
		}
		patchIdentFromDL(p, part)
		storeIdent(p)
		ids = append(ids, p)
	}
	applyCarryTags(ids)
	for _, p := range ids {
		storeIdent(p)
	}
}

func lookupDLIdent(part dlParticipant) *playerIdent {
	if part.PUUID != "" {
		identStore.mu.Lock()
		p := identStore.byPUUID[part.PUUID]
		identStore.mu.Unlock()
		if p != nil {
			return p
		}
	}
	riot := part.RiotName
	if part.RiotName != "" && part.RiotTag != "" {
		riot = part.RiotName + "#" + part.RiotTag
	}
	return findIdent(riot, part.ChampionID)
}

func identFromDL(part dlParticipant, region string) *playerIdent {
	p := &playerIdent{PUUID: part.PUUID, Champion: part.ChampionID, Ready: true}
	if part.RiotName != "" && part.RiotTag != "" {
		p.GameName, p.TagLine = part.RiotName, part.RiotTag
		p.RiotID = part.RiotName + "#" + part.RiotTag
		if region == "" {
			region = regionFromTag(part.RiotTag)
		}
		p.DeepLoL = deepLoLURL(region, part.RiotName, part.RiotTag)
	} else if part.RiotName != "" {
		p.RiotID, p.GameName = part.RiotName, part.RiotName
	}
	return p
}

func patchIdentFromDL(p *playerIdent, part dlParticipant) {
	if p == nil {
		return
	}
	if part.PUUID != "" && p.PUUID == "" {
		p.PUUID = part.PUUID
	}
	if part.ChampionID != 0 {
		p.Champion = part.ChampionID
	}
	if p.RiotID == "" && part.RiotName != "" && part.RiotTag != "" {
		p.GameName, p.TagLine = part.RiotName, part.RiotTag
		p.RiotID = part.RiotName + "#" + part.RiotTag
	}
	solo := part.Realtime.SeasonTier.Solo
	if noneTier(solo.Tier) {
		solo = part.Realtime.SeasonTier.Flex
	}
	if p.Rank == "" && !noneTier(solo.Tier) {
		p.Tier = solo.Tier
		p.Division = divFromRaw(solo.Division)
		p.LP = solo.LeaguePoints
		p.Rank = rankShort(p.Tier, p.Division, p.LP)
	}
	if p.Wins+p.Losses == 0 {
		if g := solo.Wins + solo.Losses; g > 0 {
			p.Wins, p.Losses = solo.Wins, solo.Losses
		}
	}
	if part.Info.Summoner.AIScoreAvg > 0 {
		p.AI = int(part.Info.Summoner.AIScoreAvg)
		p.HasAI = true
	}
	p.DLGames = part.Info.Summoner.Wins + part.Info.Summoner.Losses
	p.Tags = buildDeepLoLTags(part.Info.Summoner)
}

func buildDeepLoLTags(s dlSummonerInfo) []ScoutTag {
	var out []ScoutTag
	games := s.Wins + s.Losses
	if s.Tag.WinningStreak > 1 {
		n := strconv.Itoa(s.Tag.WinningStreak)
		out = append(out, ScoutTag{ID: "ws", Label: n + "W", Tone: "good", Hint: n + " victoires d'affilée (DeepLoL)"})
	} else if s.Tag.LosingStreak > 1 {
		n := strconv.Itoa(s.Tag.LosingStreak)
		out = append(out, ScoutTag{ID: "ls", Label: n + "L", Tone: "bad", Hint: n + " défaites d'affilée (DeepLoL)"})
	}
	if games > 4 {
		if s.AIScoreAvg15 > 56 {
			out = append(out, ScoutTag{ID: "early", Label: "Early", Tone: "good", Hint: "AI-score 15 min ≥ 57 sur les 10 dernières"})
		} else if s.AIScoreAvg15 > 0 && s.AIScoreAvg15 < 44 {
			out = append(out, ScoutTag{ID: "early-", Label: "Early −", Tone: "bad", Hint: "AI-score 15 min ≤ 43 sur les 10 dernières"})
		}
	}
	if s.Losses > 3 {
		if s.AIScoreAvgLoss > 47 {
			out = append(out, ScoutTag{ID: "mental+", Label: "Mental +", Tone: "good", Hint: "Tient le niveau dans les défaites (AI ≥ 48)"})
		} else if s.AIScoreAvgLoss > 0 && s.AIScoreAvgLoss < 39 {
			out = append(out, ScoutTag{ID: "mental-", Label: "Mental −", Tone: "bad", Hint: "S'écroule dans les défaites (AI ≤ 38)"})
		}
	}
	if s.Tag.RunawayCount > 0 {
		out = append(out, ScoutTag{ID: "afk", Label: "AFK", Tone: "bad", Hint: "Dodge / AFK dans les 10 dernières games"})
	}
	if s.Tag.ComebackUser {
		out = append(out, ScoutTag{ID: "cb", Label: "Comeback", Tone: "warn", Hint: "Pas joué depuis 2 semaines"})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func applyCarryTags(ids []*playerIdent) {
	type rec struct {
		p  *playerIdent
		ai int
	}
	cands := make([]rec, 0, len(ids))
	for _, p := range ids {
		if p == nil || !p.HasAI {
			continue
		}
		cands = append(cands, rec{p, p.AI})
	}
	sort.SliceStable(cands, func(i, j int) bool { return cands[i].ai > cands[j].ai })
	for i, r := range cands {
		if i >= 2 {
			break
		}
		if r.p.DLGames <= 4 || r.p.AI <= 56 {
			continue
		}
		r.p.Tags = insertCarryTag(r.p.Tags)
	}
}

func insertCarryTag(tags []ScoutTag) []ScoutTag {
	for _, t := range tags {
		if t.ID == "carry" {
			return tags
		}
	}
	carry := ScoutTag{ID: "carry", Label: "Carry", Tone: "good", Hint: "Top 2 AI-score du lobby (≥ 57, 10 dernières)"}
	i := 0
	for i < len(tags) && (tags[i].ID == "ws" || tags[i].ID == "ls") {
		i++
	}
	out := make([]ScoutTag, 0, len(tags)+1)
	out = append(out, tags[:i]...)
	out = append(out, carry)
	out = append(out, tags[i:]...)
	return out
}

func divFromRaw(raw json.RawMessage) string {
	s := strings.TrimSpace(string(raw))
	s = strings.Trim(s, `"`)
	if s == "" || s == "null" {
		return ""
	}
	if n, err := strconv.Atoi(s); err == nil {
		switch n {
		case 1:
			return "I"
		case 2:
			return "II"
		case 3:
			return "III"
		case 4:
			return "IV"
		}
	}
	return s
}

var dlSeason struct {
	mu sync.Mutex
	n  int
	at time.Time
}

func deeplolSeason() int {
	dlSeason.mu.Lock()
	n, at := dlSeason.n, dlSeason.at
	dlSeason.mu.Unlock()
	if n > 0 && time.Since(at) < 12*time.Hour {
		return n
	}
	raw, err := dlGET("https://b2c-api-cdn.deeplol.gg/common/season-list")
	if err == nil {
		var payload struct {
			SeasonList []int `json:"season_list"`
		}
		if json.Unmarshal(raw, &payload) == nil && len(payload.SeasonList) > 0 {
			n = payload.SeasonList[len(payload.SeasonList)-1]
			dlSeason.mu.Lock()
			dlSeason.n, dlSeason.at = n, time.Now()
			dlSeason.mu.Unlock()
			return n
		}
	}
	if n > 0 {
		return n
	}
	return 27
}

func platformID(region string) string {
	r := strings.ToLower(strings.TrimSpace(region))
	switch r {
	case "euw", "euw1":
		return "EUW1"
	case "eune", "eun", "eun1":
		return "EUN1"
	case "na", "na1":
		return "NA1"
	case "kr":
		return "KR"
	case "br", "br1":
		return "BR1"
	case "lan", "la1":
		return "LA1"
	case "las", "la2":
		return "LA2"
	case "oce", "oc1":
		return "OC1"
	case "ru", "ru1":
		return "RU"
	case "tr", "tr1":
		return "TR1"
	case "jp", "jp1":
		return "JP1"
	case "ph", "ph2":
		return "PH2"
	case "sg", "sg2":
		return "SG2"
	case "th", "th2":
		return "TH2"
	case "tw", "tw2":
		return "TW2"
	case "vn", "vn2":
		return "VN2"
	case "me", "me1":
		return "ME1"
	default:
		if r != "" {
			return strings.ToUpper(r)
		}
		return "EUW1"
	}
}

func dlGET(rawURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://www.deeplol.gg")
	req.Header.Set("Referer", "https://www.deeplol.gg/")
	res, err := dlHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, errHTTP(res.StatusCode)
	}
	return body, nil
}

type httpStatusError int

func (e httpStatusError) Error() string { return "HTTP " + strconv.Itoa(int(e)) }

func errHTTP(code int) error { return httpStatusError(code) }
