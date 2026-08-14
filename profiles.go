package main

// Identités joueurs (pseudo, rang solo, WR, lien DeepLoL) via LCU.
// Live Client Data donne le Riot ID ; le rang / WR viennent de /lol-ranked.
// AI-score DeepLoL (10 dernières games) : deeplol.go, en partie seulement.
// Streamer mode : le live client remplace le pseudo par le champion → on n'affiche
// pas le vrai nom et on ne pose pas de lien DeepLoL (ça le révélerait).

import (
	"encoding/json"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PlayerScout struct {
	ID      int        `json:"id"`
	Name    string     `json:"name,omitempty"`
	Hidden  bool       `json:"hidden,omitempty"`
	Rank    string     `json:"rank,omitempty"`
	Tier    string     `json:"tier,omitempty"`
	WR      *int       `json:"wr,omitempty"`
	Games   int        `json:"games,omitempty"`
	AI      *int       `json:"ai,omitempty"`
	Tags    []ScoutTag `json:"tags,omitempty"`
	DeepLoL string     `json:"deeplol,omitempty"`
	RiotID  string     `json:"riotId,omitempty"`
}

type playerIdent struct {
	PUUID    string
	RiotID   string
	GameName string
	TagLine  string
	Champion int
	Tier     string
	Division string
	LP       int
	Rank     string
	Wins     int
	Losses   int
	AI       int
	HasAI    bool
	DLGames  int
	Tags     []ScoutTag
	DeepLoL  string
	Hidden   bool
	Ready    bool
}

type lcuSummoner struct {
	GameName    string `json:"gameName"`
	TagLine     string `json:"tagLine"`
	DisplayName string `json:"displayName"`
	PUUID       string `json:"puuid"`
	Unnamed     bool   `json:"unnamed"`
}

type lcuRankedEntry struct {
	QueueType    string `json:"queueType"`
	Tier         string `json:"tier"`
	Division     string `json:"division"`
	LeaguePoints int    `json:"leaguePoints"`
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
}

type lcuRankedStats struct {
	QueueMap           map[string]lcuRankedEntry `json:"queueMap"`
	Queues             []lcuRankedEntry          `json:"queues"`
	HighestRankedEntry lcuRankedEntry            `json:"highestRankedEntry"`
}

var identStore = struct {
	mu      sync.Mutex
	region  string
	gameKey string
	byPUUID map[string]*playerIdent
	byRiot  map[string]*playerIdent
	byChamp map[int]*playerIdent
	pending map[string]bool
}{
	byPUUID: map[string]*playerIdent{},
	byRiot:  map[string]*playerIdent{},
	byChamp: map[int]*playerIdent{},
	pending: map[string]bool{},
}

var harvestMu sync.Mutex
var harvestAt time.Time

func kickHarvest(creds *lcuCreds, phase string) {
	if creds == nil {
		return
	}
	harvestMu.Lock()
	if time.Since(harvestAt) < 1500*time.Millisecond {
		harvestMu.Unlock()
		return
	}
	harvestAt = time.Now()
	c := *creds
	harvestMu.Unlock()
	go harvestIdentities(&c, phase)
}

type identJob struct {
	puuid  string
	sid    int64
	champ  int
	hidden bool
	riot   string
}

func harvestIdentities(creds *lcuCreds, phase string) {
	ensureRegion(creds)
	var jobs []identJob
	switch phase {
	case "ChampSelect":
		jobs = jobsFromChampSelect(creds)
	case "GameStart", "InProgress", "Reconnect", "WaitingForStats":
		jobs = jobsFromGameflow(creds)
	default:
		return
	}
	if len(jobs) == 0 {
		return
	}
	keys := make([]string, 0, len(jobs))
	for _, j := range jobs {
		if j.puuid != "" {
			keys = append(keys, j.puuid)
		} else if j.sid != 0 {
			keys = append(keys, "id:"+strconv.FormatInt(j.sid, 10))
		}
	}
	resetGameIfNew(strings.Join(keys, "|"))

	sem := make(chan struct{}, 2)
	var wg sync.WaitGroup
	for _, j := range jobs {
		j := j
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			enrichMember(creds, j.puuid, j.sid, j.champ, j.hidden, j.riot)
			<-sem
		}()
	}
	wg.Wait()
	c := *creds
	go enrichDeepLoL(&c, phase)
}

type flexInt64 int64

func (f *flexInt64) UnmarshalJSON(b []byte) error {
	*f = flexInt64(parseAnyID(b))
	return nil
}

type csMember struct {
	CellID             int       `json:"cellId"`
	ChampionID         int       `json:"championId"`
	ChampionPickIntent int       `json:"championPickIntent"`
	AssignedPosition   string    `json:"assignedPosition"`
	SummonerID         flexInt64 `json:"summonerId"`
	PUUID              string    `json:"puuid"`
	NameVisibilityType string    `json:"nameVisibilityType"`
}

func jobsFromChampSelect(creds *lcuCreds) []identJob {
	var session struct {
		MyTeam    []csMember `json:"myTeam"`
		TheirTeam []csMember `json:"theirTeam"`
	}
	if _, err := lcuGET(creds, "/lol-champ-select/v1/session", &session); err != nil {
		return nil
	}
	out := make([]identJob, 0, 10)
	add := func(m csMember) {
		if m.PUUID == "" && m.SummonerID == 0 {
			return
		}
		out = append(out, identJob{
			puuid:  m.PUUID,
			sid:    int64(m.SummonerID),
			champ:  champID(m.ChampionID, m.ChampionPickIntent),
			hidden: strings.EqualFold(m.NameVisibilityType, "HIDDEN"),
		})
	}
	for _, m := range session.MyTeam {
		add(m)
	}
	for _, m := range session.TheirTeam {
		add(m)
	}
	return out
}

func jobsFromGameflow(creds *lcuCreds) []identJob {
	var session struct {
		GameData struct {
			TeamOne []struct {
				ChampionId   int             `json:"championId"`
				PUUID        string          `json:"puuid"`
				SummonerId   json.RawMessage `json:"summonerId"`
				SummonerName string          `json:"summonerName"`
			} `json:"teamOne"`
			TeamTwo []struct {
				ChampionId   int             `json:"championId"`
				PUUID        string          `json:"puuid"`
				SummonerId   json.RawMessage `json:"summonerId"`
				SummonerName string          `json:"summonerName"`
			} `json:"teamTwo"`
		} `json:"gameData"`
	}
	if _, err := lcuGET(creds, "/lol-gameflow/v1/session", &session); err != nil {
		return nil
	}
	out := make([]identJob, 0, 10)
	add := func(puuid string, rawID json.RawMessage, champ int, name string) {
		if puuid == "" && parseAnyID(rawID) == 0 {
			return
		}
		out = append(out, identJob{puuid: puuid, sid: parseAnyID(rawID), champ: champ, riot: name})
	}
	for _, p := range session.GameData.TeamOne {
		add(p.PUUID, p.SummonerId, p.ChampionId, p.SummonerName)
	}
	for _, p := range session.GameData.TeamTwo {
		add(p.PUUID, p.SummonerId, p.ChampionId, p.SummonerName)
	}
	return out
}

func parseAnyID(raw json.RawMessage) int64 {
	s := strings.TrimSpace(string(raw))
	s = strings.Trim(s, `"`)
	if s == "" || s == "null" {
		return 0
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int64(f)
	}
	return 0
}

func wrPct(wins, games int) int {
	if games <= 0 {
		return 0
	}
	return int(math.Round(float64(wins) * 100 / float64(games)))
}

func identWR(p *playerIdent) (*int, int) {
	if p == nil {
		return nil, 0
	}
	g := p.Wins + p.Losses
	if g <= 0 {
		return nil, 0
	}
	v := wrPct(p.Wins, g)
	return &v, g
}

func identAI(p *playerIdent) *int {
	if p == nil || !p.HasAI {
		return nil
	}
	v := p.AI
	return &v
}

func resetGameIfNew(key string) {
	if key == "" {
		return
	}
	identStore.mu.Lock()
	defer identStore.mu.Unlock()
	if identStore.gameKey == key {
		return
	}
	identStore.gameKey = key
	identStore.byChamp = map[int]*playerIdent{}
	resetDeepLoL()
}

func enrichMember(creds *lcuCreds, puuid string, summonerID int64, champID int, hidden bool, fallbackName string) {
	key := puuid
	if key == "" {
		if summonerID == 0 {
			return
		}
		key = "id:" + strconv.FormatInt(summonerID, 10)
	}
	identStore.mu.Lock()
	if p := identStore.byPUUID[puuid]; puuid != "" && p != nil && p.Ready {
		if champID != 0 {
			p.Champion = champID
			identStore.byChamp[champID] = p
		}
		identStore.mu.Unlock()
		return
	}
	if identStore.pending[key] {
		identStore.mu.Unlock()
		return
	}
	identStore.pending[key] = true
	identStore.mu.Unlock()
	defer func() {
		identStore.mu.Lock()
		delete(identStore.pending, key)
		identStore.mu.Unlock()
	}()

	var sum lcuSummoner
	if puuid != "" {
		_, _ = lcuGET(creds, "/lol-summoner/v2/summoners/puuid/"+url.PathEscape(puuid), &sum)
	}
	if sum.PUUID == "" && summonerID != 0 {
		_, _ = lcuGET(creds, "/lol-summoner/v1/summoners/"+strconv.FormatInt(summonerID, 10), &sum)
	}
	if sum.PUUID != "" {
		puuid = sum.PUUID
	}

	p := &playerIdent{PUUID: puuid, Champion: champID, Hidden: hidden || sum.Unnamed, Ready: true}
	switch {
	case sum.GameName != "" && sum.TagLine != "":
		p.GameName, p.TagLine = sum.GameName, sum.TagLine
		p.RiotID = sum.GameName + "#" + sum.TagLine
	case sum.DisplayName != "":
		p.RiotID, p.GameName = sum.DisplayName, sum.DisplayName
	case fallbackName != "":
		p.RiotID, p.GameName = fallbackName, fallbackName
	}
	if p.RiotID == "" {
		p.Hidden = true
	}

	if puuid != "" {
		var stats lcuRankedStats
		if _, err := lcuGET(creds, "/lol-ranked/v1/ranked-stats/"+url.PathEscape(puuid), &stats); err != nil {
			_, _ = lcuGET(creds, "/lol-ranked/v1/cached-ranked-stats/"+url.PathEscape(puuid), &stats)
		}
		e := pickSolo(&stats)
		p.Tier, p.Division, p.LP = e.Tier, e.Division, e.LeaguePoints
		p.Rank = rankShort(e.Tier, e.Division, e.LeaguePoints)
		p.Wins, p.Losses = e.Wins, e.Losses
	}

	if !p.Hidden {
		region := cachedRegion()
		if region == "" {
			region = ensureRegion(creds)
		}
		if p.GameName != "" && p.TagLine != "" {
			p.DeepLoL = deepLoLURL(region, p.GameName, p.TagLine)
		} else {
			p.DeepLoL = deepLoLFromRiotID(p.RiotID)
		}
	}
	storeIdent(p)
}

func storeIdent(p *playerIdent) {
	identStore.mu.Lock()
	defer identStore.mu.Unlock()
	if p.PUUID != "" {
		identStore.byPUUID[p.PUUID] = p
	}
	if p.RiotID != "" {
		identStore.byRiot[strings.ToLower(p.RiotID)] = p
		if name, _, ok := splitRiot(p.RiotID); ok {
			identStore.byRiot[strings.ToLower(name)] = p
		} else {
			identStore.byRiot[strings.ToLower(p.RiotID)] = p
		}
	}
	if p.Champion != 0 {
		identStore.byChamp[p.Champion] = p
	}
}

func findIdent(riotID string, champKey int) *playerIdent {
	identStore.mu.Lock()
	defer identStore.mu.Unlock()
	if riotID != "" {
		if p := identStore.byRiot[strings.ToLower(riotID)]; p != nil {
			return p
		}
		if name, _, ok := splitRiot(riotID); ok {
			if p := identStore.byRiot[strings.ToLower(name)]; p != nil {
				return p
			}
		}
	}
	if champKey != 0 {
		return identStore.byChamp[champKey]
	}
	return nil
}

func scoutFromChamp(id int) *PlayerScout {
	identStore.mu.Lock()
	p := identStore.byChamp[id]
	identStore.mu.Unlock()
	if p == nil {
		return nil
	}
	s := &PlayerScout{ID: id, Rank: p.Rank, Tier: strings.ToLower(strings.TrimSpace(p.Tier))}
	if noneTier(s.Tier) {
		s.Tier, s.Rank = "", ""
	}
	s.WR, s.Games = identWR(p)
	s.AI = identAI(p)
	s.Tags = p.Tags
	if p.Hidden {
		s.Hidden = true
		return s
	}
	s.Name, s.RiotID, s.DeepLoL = p.RiotID, p.RiotID, p.DeepLoL
	return s
}

func attachDraftScout(snap *Snapshot) {
	snap.EnemyScout = nil
	for _, id := range snap.Enemies {
		if s := scoutFromChamp(id); s != nil {
			snap.EnemyScout = append(snap.EnemyScout, *s)
		}
	}
	for i, a := range snap.AllyFocus {
		if s := scoutFromChamp(a.ID); s != nil {
			snap.AllyFocus[i].Name = s.Name
			snap.AllyFocus[i].Hidden = s.Hidden
			snap.AllyFocus[i].Rank = s.Rank
			snap.AllyFocus[i].Tier = s.Tier
			snap.AllyFocus[i].WR = s.WR
			snap.AllyFocus[i].Games = s.Games
			snap.AllyFocus[i].AI = s.AI
			snap.AllyFocus[i].Tags = s.Tags
			snap.AllyFocus[i].DeepLoL = s.DeepLoL
			snap.AllyFocus[i].RiotID = s.RiotID
		}
	}
}

func applyIdentity(lp *LivePlayer) {
	name, hidden := visibleName(lp.RiotID, lp.Champion)
	lp.Name, lp.Hidden = name, hidden
	lp.Rank, lp.Tier, lp.DeepLoL = "", "", ""
	lp.WR, lp.Games, lp.AI = nil, 0, nil
	lp.Tags = nil
	if !hidden {
		lp.DeepLoL = deepLoLFromRiotID(lp.RiotID)
	}
	id := findIdent(lp.RiotID, lp.Key)
	if id == nil {
		return
	}
	if !noneTier(id.Tier) {
		lp.Tier = strings.ToLower(strings.TrimSpace(id.Tier))
		lp.Rank = id.Rank
	}
	lp.WR, lp.Games = identWR(id)
	lp.AI = identAI(id)
	lp.Tags = id.Tags
	if hidden {
		return
	}
	if id.DeepLoL != "" {
		lp.DeepLoL = id.DeepLoL
	}
	if id.RiotID != "" && !id.Hidden {
		lp.Name = id.RiotID
	}
}

func ensureRegion(creds *lcuCreds) string {
	if r := cachedRegion(); r != "" {
		return r
	}
	var loc struct {
		Region    string `json:"region"`
		WebRegion string `json:"webRegion"`
	}
	if _, err := lcuGET(creds, "/riotclient/region-locale", &loc); err != nil {
		return ""
	}
	r := strings.ToLower(strings.TrimSpace(loc.WebRegion))
	if r == "" {
		r = strings.ToLower(strings.TrimSpace(loc.Region))
	}
	if r == "pbe" {
		r = ""
	}
	identStore.mu.Lock()
	identStore.region = r
	identStore.mu.Unlock()
	return r
}

func cachedRegion() string {
	identStore.mu.Lock()
	defer identStore.mu.Unlock()
	return identStore.region
}

func splitRiot(id string) (name, tag string, ok bool) {
	id = strings.TrimSpace(id)
	i := strings.LastIndex(id, "#")
	if i <= 0 || i == len(id)-1 {
		return id, "", false
	}
	return id[:i], id[i+1:], true
}

func isBotID(id string) bool {
	l := strings.ToLower(id)
	return strings.Contains(l, "character") || strings.HasPrefix(l, "game_")
}

func visibleName(riotID, champion string) (string, bool) {
	riotID = strings.TrimSpace(riotID)
	champ := strings.TrimSpace(champion)
	if riotID == "" || isBotID(riotID) {
		return champ, true
	}
	name, tag, tagged := splitRiot(riotID)
	if strings.EqualFold(name, champ) {
		return champ, true
	}
	if tagged {
		return name + "#" + tag, false
	}
	if name != "" {
		return name, false
	}
	return champ, true
}

func noneTier(t string) bool {
	t = strings.ToUpper(strings.TrimSpace(t))
	return t == "" || t == "NONE" || t == "UNRANKED"
}

func romanDiv(d string) string {
	switch strings.ToUpper(strings.TrimSpace(d)) {
	case "I", "1":
		return "1"
	case "II", "2":
		return "2"
	case "III", "3":
		return "3"
	case "IV", "4":
		return "4"
	default:
		return ""
	}
}

func rankShort(tier, div string, lp int) string {
	t := strings.ToUpper(strings.TrimSpace(tier))
	if noneTier(t) {
		return ""
	}
	letters := map[string]string{
		"IRON": "I", "BRONZE": "B", "SILVER": "S", "GOLD": "G",
		"PLATINUM": "P", "EMERALD": "E", "DIAMOND": "D",
		"MASTER": "Master", "GRANDMASTER": "GM", "CHALLENGER": "Chall",
	}
	s, ok := letters[t]
	if !ok {
		return t
	}
	if t == "MASTER" || t == "GRANDMASTER" || t == "CHALLENGER" {
		if lp > 0 {
			return s + " " + strconv.Itoa(lp)
		}
		return s
	}
	if n := romanDiv(div); n != "" {
		return s + n
	}
	return s
}

func pickSolo(s *lcuRankedStats) lcuRankedEntry {
	if s == nil {
		return lcuRankedEntry{}
	}
	get := func(q string) (lcuRankedEntry, bool) {
		if s.QueueMap != nil {
			if e, ok := s.QueueMap[q]; ok && !noneTier(e.Tier) {
				return e, true
			}
		}
		for _, e := range s.Queues {
			if e.QueueType == q && !noneTier(e.Tier) {
				return e, true
			}
		}
		return lcuRankedEntry{}, false
	}
	for _, q := range []string{"RANKED_SOLO_5x5", "RANKED_FLEX_SR"} {
		if e, ok := get(q); ok {
			return e
		}
	}
	if !noneTier(s.HighestRankedEntry.Tier) {
		return s.HighestRankedEntry
	}
	return lcuRankedEntry{}
}

func regionFromTag(tag string) string {
	t := strings.ToUpper(strings.TrimSpace(tag))
	switch {
	case strings.HasPrefix(t, "EUNE") || strings.HasPrefix(t, "EUN"):
		return "eune"
	case strings.HasPrefix(t, "EUW"):
		return "euw"
	case strings.HasPrefix(t, "NA"):
		return "na"
	case strings.HasPrefix(t, "KR"):
		return "kr"
	case strings.HasPrefix(t, "BR"):
		return "br"
	case t == "LA1" || t == "LAN":
		return "lan"
	case t == "LA2" || t == "LAS":
		return "las"
	case strings.HasPrefix(t, "OC"):
		return "oce"
	case t == "RU" || t == "RU1":
		return "ru"
	case strings.HasPrefix(t, "TR"):
		return "tr"
	case strings.HasPrefix(t, "JP"):
		return "jp"
	case strings.HasPrefix(t, "PH"):
		return "ph"
	case strings.HasPrefix(t, "SG"):
		return "sg"
	case strings.HasPrefix(t, "TH"):
		return "th"
	case strings.HasPrefix(t, "TW"):
		return "tw"
	case strings.HasPrefix(t, "VN"):
		return "vn"
	case strings.HasPrefix(t, "ME"):
		return "me"
	default:
		return ""
	}
}

func deepLoLURL(region, gameName, tag string) string {
	region = strings.ToLower(strings.TrimSpace(region))
	gameName = strings.TrimSpace(gameName)
	tag = strings.TrimSpace(tag)
	if region == "" || gameName == "" || tag == "" {
		return ""
	}
	return "https://www.deeplol.gg/summoner/" + region + "/" + url.PathEscape(gameName) + "-" + url.PathEscape(tag)
}

func deepLoLFromRiotID(riotID string) string {
	name, tag, ok := splitRiot(riotID)
	if !ok {
		return ""
	}
	region := cachedRegion()
	if region == "" {
		region = regionFromTag(tag)
	}
	return deepLoLURL(region, name, tag)
}

func isDeepLoLURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "https" {
		return false
	}
	host := strings.ToLower(u.Host)
	return host == "www.deeplol.gg" || host == "deeplol.gg"
}

func apiOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST", http.StatusMethodNotAllowed)
		return
	}
	u := strings.TrimSpace(r.URL.Query().Get("url"))
	if !isDeepLoLURL(u) {
		http.Error(w, "url", http.StatusBadRequest)
		return
	}
	go openBrowser(u)
	writeJSON(w, map[string]any{"ok": true})
}
