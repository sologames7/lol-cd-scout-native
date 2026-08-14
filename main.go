package main

import (
	"context"
	"crypto/tls"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const fallbackVersion = "16.15.1"

type Spell struct {
	Spell      string `json:"spell"`
	Name       string `json:"name"`
	CD         string `json:"cd"`
	Icon       string `json:"icon,omitempty"` // chemin relatif ddragon, ex "spell/AhriQ.png"
	Desc       string `json:"desc,omitempty"`
	Cost       string `json:"cost,omitempty"`
	Range      string `json:"range,omitempty"`
	Importance int    `json:"importance"`
	Note       string `json:"note"`
}

type Override struct {
	Summary      string   `json:"summary"`
	Window       string   `json:"window"`
	Important    []Spell  `json:"important"`
	Counters     []string `json:"counters,omitempty"`
	HardMatchups []string `json:"hardMatchups,omitempty"`
	Synergies    []string `json:"synergies,omitempty"`
}

type ChampionCard struct {
	ID           string   `json:"id"`
	Key          int      `json:"key"`
	Name         string   `json:"name"`
	Title        string   `json:"title"`
	Icon         string   `json:"icon"`
	Important    []Spell  `json:"important"`
	Summary      string   `json:"summary"`
	Window       string   `json:"window"`
	Counters     []string `json:"counters,omitempty"`
	HardMatchups []string `json:"hardMatchups,omitempty"`
	Synergies    []string `json:"synergies,omitempty"`
	Source       string   `json:"source"`
}

type AllyFocus struct {
	Role    string     `json:"role"`
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

type teamMember struct {
	CellID             int       `json:"cellId"`
	ChampionID         int       `json:"championId"`
	ChampionPickIntent int       `json:"championPickIntent"`
	AssignedPosition   string    `json:"assignedPosition"`
	SummonerID         flexInt64 `json:"summonerId"`
	PUUID              string    `json:"puuid"`
	NameVisibilityType string    `json:"nameVisibilityType"`
}

type Snapshot struct {
	Connected  bool          `json:"connected"`
	Phase      string        `json:"phase"`
	Enemies    []int         `json:"enemies"`
	Allies     []int         `json:"allies,omitempty"`
	AllyFocus  []AllyFocus   `json:"allyFocus,omitempty"`
	EnemyScout []PlayerScout `json:"enemyScout,omitempty"`
	MyRole     string        `json:"myRole,omitempty"`
	MyChamp    int           `json:"myChamp,omitempty"`
	Bans       []int         `json:"bans,omitempty"`
	Draft      *DraftAdvice  `json:"draft,omitempty"`
	Error      string        `json:"error,omitempty"`
	Lockfile   string        `json:"lockfile,omitempty"`
	Version    string        `json:"version,omitempty"`
	App        string        `json:"app,omitempty"`
	Update     *UpdateInfo   `json:"update,omitempty"`
}

type lcuCreds struct {
	Port     int
	Password string
	Protocol string
	Lockfile string
}

type champIndexItem struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Name  string `json:"name"`
	Title string `json:"title"`
	Image struct {
		Full string `json:"full"`
	} `json:"image"`
}

type champIndexPayload struct {
	Data map[string]champIndexItem `json:"data"`
}

type champDetail struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Name  string `json:"name"`
	Title string `json:"title"`
	Image struct {
		Full string `json:"full"`
	} `json:"image"`
	Passive struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Image       struct {
			Full string `json:"full"`
		} `json:"image"`
	} `json:"passive"`
	Spells []struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		CooldownBurn string `json:"cooldownBurn"`
		CostBurn     string `json:"costBurn"`
		RangeBurn    string `json:"rangeBurn"`
		Image        struct {
			Full string `json:"full"`
		} `json:"image"`
	} `json:"spells"`
}

type champDetailPayload struct {
	Data map[string]champDetail `json:"data"`
}

// Curated champion priorities live in curated.go.

type dragonCache struct {
	mu      sync.Mutex
	version string
	index   []champIndexItem
	byKey   map[int]champIndexItem
	details map[string]champDetail
	lastErr string
}

var dragon = dragonCache{details: map[string]champDetail{}}

// IPv4 d'abord : un DNS IPv6 blackhole fait autrement bloquer tout /api/status
// (et saturait les 6 connexions HTTP/1.1 du navigateur → Quitter ne partait plus).
var httpClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		DialContext:           dialHTTP,
		ForceAttemptHTTP2:     false,
		TLSHandshakeTimeout:   4 * time.Second,
		ResponseHeaderTimeout: 4 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		MaxIdleConns:          8,
		MaxIdleConnsPerHost:   4,
	},
}

func dialHTTP(ctx context.Context, _, addr string) (net.Conn, error) {
	d := net.Dialer{Timeout: 3 * time.Second}
	c, err := d.DialContext(ctx, "tcp4", addr)
	if err == nil {
		return c, nil
	}
	return d.DialContext(ctx, "tcp", addr)
}

// Reuse one LCU client — creating a Transport per poll leaked dozens of sockets to League.
var lcuHTTP = &http.Client{
	Timeout: 2500 * time.Millisecond,
	Transport: &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, // loopback League client certificate
		MaxIdleConns:        4,
		MaxIdleConnsPerHost: 2,
		IdleConnTimeout:     4 * time.Second,
		ForceAttemptHTTP2:   false,
	},
}

var (
	snapMu    sync.Mutex
	snapCache Snapshot
	snapAt    time.Time
)

const listenAddr = "127.0.0.1:27182"

func jsonGET(url string, out any) error {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "LoL-CD-Scout/0.2")
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(out)
}

func getVersion() string {
	dragon.mu.Lock()
	defer dragon.mu.Unlock()
	if dragon.version != "" {
		return dragon.version
	}
	return fallbackVersion
}

// refreshDragon met à jour la version CDN en fond : /api/status ne doit jamais
// attendre le réseau Riot, sinon le front reste bloqué sur « Data Dragon … »
// et le navigateur n'a plus de connexion libre pour /api/quit.
func refreshDragon() {
	var versions []string
	err := jsonGET("https://ddragon.leagueoflegends.com/api/versions.json", &versions)
	ver := fallbackVersion
	if err == nil && len(versions) > 0 {
		ver = versions[0]
	}
	dragon.mu.Lock()
	changed := dragon.version != ver
	dragon.version = ver
	if err != nil {
		dragon.lastErr = err.Error()
	} else {
		dragon.lastErr = ""
	}
	if changed {
		dragon.index, dragon.byKey, dragon.details = nil, nil, map[string]champDetail{}
		resetPassiveCDCache()
		resetItemMetaCache()
	}
	dragon.mu.Unlock()
	_ = ensureIndex()
	go runeTable()
}

func ensureIndex() error {
	dragon.mu.Lock()
	if len(dragon.index) > 0 {
		dragon.mu.Unlock()
		return nil
	}
	dragon.mu.Unlock()

	version := getVersion()
	var payload champIndexPayload
	url := fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/fr_FR/champion.json", version)
	if err := jsonGET(url, &payload); err != nil {
		url = fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion.json", version)
		if err2 := jsonGET(url, &payload); err2 != nil {
			return err2
		}
	}

	items := make([]champIndexItem, 0, len(payload.Data))
	byKey := map[int]champIndexItem{}
	for _, c := range payload.Data {
		items = append(items, c)
		if k, err := strconv.Atoi(c.Key); err == nil {
			byKey[k] = c
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	dragon.mu.Lock()
	mismatch := dragon.version != "" && dragon.version != version
	if !mismatch {
		dragon.index, dragon.byKey = items, byKey
	}
	dragon.mu.Unlock()
	if mismatch {
		return ensureIndex()
	}
	return nil
}

func getDetail(slug string) (champDetail, error) {
	dragon.mu.Lock()
	if d, ok := dragon.details[slug]; ok {
		dragon.mu.Unlock()
		return d, nil
	}
	dragon.mu.Unlock()
	version := getVersion()
	var payload champDetailPayload
	url := fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/fr_FR/champion/%s.json", version, slug)
	if err := jsonGET(url, &payload); err != nil {
		url = fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion/%s.json", version, slug)
		if err2 := jsonGET(url, &payload); err2 != nil {
			return champDetail{}, err2
		}
	}
	d, ok := payload.Data[slug]
	if !ok {
		return champDetail{}, errors.New("champion introuvable")
	}
	dragon.mu.Lock()
	dragon.details[slug] = d
	dragon.mu.Unlock()
	return d, nil
}

func cooldown(raw string) string {
	if raw == "" || raw == "0" {
		return "—"
	}
	return strings.ReplaceAll(raw, "/", " → ") + "s"
}

var brRe = regexp.MustCompile(`(?i)<br\s*/?>`)
var tagRe = regexp.MustCompile(`(?s)<[^>]*>`)
var multiSpaceRe = regexp.MustCompile(`[ \t]{2,}`)

// Les descriptions ddragon mélangent balises maison (<magicDamage>…), <br> et espaces insécables.
func plainText(s string) string {
	s = brRe.ReplaceAllString(s, "\n")
	s = tagRe.ReplaceAllString(s, "")
	s = strings.NewReplacer("&nbsp;", " ", "\u00a0", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'").Replace(s)
	s = multiSpaceRe.ReplaceAllString(s, " ")
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func burn(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "0" {
		return ""
	}
	if raw == "self" {
		return "soi"
	}
	return strings.ReplaceAll(raw, "/", " → ")
}

func spellIconPath(full string) string {
	if full == "" {
		return ""
	}
	return "spell/" + full
}

func passiveFromDetail(d champDetail) Spell {
	icon := ""
	if f := d.Passive.Image.Full; f != "" {
		icon = "passive/" + f
	}
	name := d.Passive.Name
	if name == "" {
		name = "Passif"
	}
	desc := plainText(d.Passive.Description)
	cd := formatPassiveCD(passiveCDFor(d.ID, desc).Values)
	return Spell{Spell: "P", Name: name, CD: cd, Icon: icon, Desc: desc, Importance: 4}
}

func isBasicSpell(key string) bool {
	return key == "Q" || key == "W" || key == "E" || key == "R"
}

func normalize(d champDetail) ChampionCard {
	letters := []string{"Q", "W", "E", "R"}
	auto := map[string]Spell{}
	for i, s := range d.Spells {
		if i >= len(letters) {
			break
		}
		auto[letters[i]] = Spell{
			Spell: letters[i], Name: s.Name, CD: cooldown(s.CooldownBurn),
			Icon: spellIconPath(s.Image.Full), Desc: plainText(s.Description),
			Cost: burn(s.CostBurn), Range: burn(s.RangeBurn), Importance: 5,
		}
	}
	passive := passiveFromDetail(d)
	summary := ""
	window := ""
	source := "ddragon"
	var counters, hardMatchups, syns []string
	if m, ok := matchups[d.ID]; ok {
		counters, hardMatchups = m.Counters, m.HardMatchups
	}
	if s, ok := synergies[d.ID]; ok {
		syns = s
	}
	curatedByKey := map[string]Spell{}
	var curatedOrder []Spell
	if o, ok := overrides[d.ID]; ok {
		source, summary, window = "curated", o.Summary, o.Window
		if len(o.Counters) > 0 {
			counters = o.Counters
		}
		if len(o.HardMatchups) > 0 {
			hardMatchups = o.HardMatchups
		}
		if len(o.Synergies) > 0 {
			syns = o.Synergies
		}
		for _, s := range o.Important {
			if a, ok := auto[s.Spell]; ok {
				if s.CD == "auto" || s.CD == "" {
					s.CD = a.CD
				}
				if s.Name == "" {
					s.Name = a.Name
				}
				s.Icon, s.Desc, s.Cost, s.Range = a.Icon, a.Desc, a.Cost, a.Range
			} else if s.Spell == "P" {
				if s.Name == "" {
					s.Name = passive.Name
				}
				if s.CD == "" || s.CD == "auto" || s.CD == "—" {
					s.CD = passive.CD
				}
				s.Icon, s.Desc = passive.Icon, passive.Desc
			}
			curatedByKey[s.Spell] = s
			curatedOrder = append(curatedOrder, s)
		}
	}
	// Passif d'abord, puis extras curated, puis Q/W/E/R.
	important := []Spell{}
	if s, ok := curatedByKey["P"]; ok {
		important = append(important, s)
	} else {
		important = append(important, passive)
	}
	for _, s := range curatedOrder {
		if s.Spell != "P" && !isBasicSpell(s.Spell) {
			important = append(important, s)
		}
	}
	for _, l := range letters {
		if s, ok := curatedByKey[l]; ok {
			important = append(important, s)
		} else if a, ok := auto[l]; ok {
			important = append(important, a)
		}
	}
	if source == "ddragon" {
		parts := []string{}
		for _, s := range important {
			if s.CD != "—" && len(parts) < 3 {
				parts = append(parts, s.Spell)
			}
		}
		summary = strings.Join(parts, " > ")
	}
	key, _ := strconv.Atoi(d.Key)
	return ChampionCard{
		ID: d.ID, Key: key, Name: d.Name, Title: d.Title, Icon: d.Image.Full,
		Important: important, Summary: summary, Window: window,
		Counters: counters, HardMatchups: hardMatchups, Synergies: syns, Source: source,
	}
}

func cardByKey(key int) (ChampionCard, error) {
	if err := ensureIndex(); err != nil {
		return ChampionCard{}, err
	}
	dragon.mu.Lock()
	base, ok := dragon.byKey[key]
	dragon.mu.Unlock()
	if !ok {
		return ChampionCard{}, errors.New("champion inconnu")
	}
	d, err := getDetail(base.ID)
	if err != nil {
		return ChampionCard{}, err
	}
	return normalize(d), nil
}

func searchChampions(q string) []champIndexItem {
	if ensureIndex() != nil {
		return nil
	}
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return nil
	}
	dragon.mu.Lock()
	items := append([]champIndexItem(nil), dragon.index...)
	dragon.mu.Unlock()
	out := []champIndexItem{}
	for _, c := range items {
		if strings.Contains(strings.ToLower(c.Name), q) || strings.Contains(strings.ToLower(c.ID), q) {
			out = append(out, c)
			if len(out) == 8 {
				break
			}
		}
	}
	return out
}

func commonLockfiles() []string {
	out := []string{}
	if p := os.Getenv("LEAGUE_PATH"); p != "" {
		out = append(out, filepath.Join(p, "lockfile"))
	}
	for _, drive := range []string{"C:", "D:", "E:", "F:"} {
		out = append(out,
			filepath.Join(drive+`\`, "Riot Games", "League of Legends", "lockfile"),
			filepath.Join(drive+`\`, "Program Files", "Riot Games", "League of Legends", "lockfile"),
		)
	}
	// Riot metadata often survives custom installation paths.
	if pd := os.Getenv("PROGRAMDATA"); pd != "" {
		meta := filepath.Join(pd, "Riot Games", "Metadata", "league_of_legends.live", "league_of_legends.live.product_settings.yaml")
		if b, err := os.ReadFile(meta); err == nil {
			for _, line := range strings.Split(string(b), "\n") {
				if strings.Contains(line, "product_install_full_path:") {
					p := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
					p = strings.Trim(p, "\"'")
					p = strings.ReplaceAll(p, `\\`, `\`)
					if p != "" {
						out = append([]string{filepath.Join(p, "lockfile")}, out...)
					}
				}
			}
		}
	}
	return out
}

func powershellInstallDir() string {
	script := `$p = Get-CimInstance Win32_Process -Filter "Name='LeagueClientUx.exe'" | Select-Object -First 1 -ExpandProperty CommandLine; if ($p -match '--install-directory[= ]\"?([^\"]+)') { $matches[1].Trim() }`
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = windowsHiddenProcAttr()
	b, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func findLockfile() string {
	if dir := powershellInstallDir(); dir != "" {
		p := filepath.Join(dir, "lockfile")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	for _, p := range commonLockfiles() {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func getCreds() (*lcuCreds, error) {
	lock := findLockfile()
	if lock == "" {
		return nil, os.ErrNotExist
	}
	b, err := os.ReadFile(lock)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(strings.TrimSpace(string(b)), ":")
	if len(parts) < 5 {
		return nil, errors.New("lockfile League invalide")
	}
	port, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, err
	}
	return &lcuCreds{Port: port, Password: parts[3], Protocol: parts[4], Lockfile: lock}, nil
}

func lcuGET(creds *lcuCreds, endpoint string, out any) (int, error) {
	url := fmt.Sprintf("https://127.0.0.1:%d%s", creds.Port, endpoint)
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		auth := base64.StdEncoding.EncodeToString([]byte("riot:" + creds.Password))
		req.Header.Set("Authorization", "Basic "+auth)
		req.Header.Set("Accept", "application/json")
		res, err := lcuHTTP.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(120 * time.Millisecond)
			continue
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
			return res.StatusCode, fmt.Errorf("LCU %d", res.StatusCode)
		}
		if out == nil {
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
			return res.StatusCode, nil
		}
		err = json.NewDecoder(res.Body).Decode(out)
		res.Body.Close()
		return res.StatusCode, err
	}
	return 0, lastErr
}

func uniqueNonzero(v []int) []int {
	seen := map[int]bool{}
	out := []int{}
	for _, n := range v {
		if n != 0 && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}

func champID(championID, pickIntent int) int {
	if championID != 0 {
		return championID
	}
	return pickIntent
}

func roleLabel(pos string) string {
	switch roleShort(pos) {
	case "SUPP", "JGL", "MID":
		return roleShort(pos)
	default:
		return ""
	}
}

func allyFocusFromTeam(myTeam []teamMember, localCell int) []AllyFocus {
	byRole := map[string]int{}
	for _, p := range myTeam {
		if p.CellID == localCell {
			continue
		}
		role := roleLabel(p.AssignedPosition)
		if role == "" {
			continue
		}
		if id := champID(p.ChampionID, p.ChampionPickIntent); id != 0 {
			byRole[role] = id
		}
	}
	out := []AllyFocus{}
	for _, role := range []string{"SUPP", "JGL", "MID"} {
		if id, ok := byRole[role]; ok {
			out = append(out, AllyFocus{Role: role, ID: id})
		}
	}
	return out
}

func getSnapshot() Snapshot {
	snapMu.Lock()
	if time.Since(snapAt) < 500*time.Millisecond && snapAt != (time.Time{}) {
		s := snapCache
		snapMu.Unlock()
		return s
	}
	snapMu.Unlock()

	snap := Snapshot{Connected: false, Phase: "Disconnected", Version: getVersion()}
	creds, err := getCreds()
	if err != nil {
		return storeSnapshot(snap)
	}
	var phase string
	if _, err := lcuGET(creds, "/lol-gameflow/v1/gameflow-phase", &phase); err != nil {
		snap.Phase, snap.Error = "Connection error", err.Error()
		return storeSnapshot(snap)
	}
	snap.Connected, snap.Phase, snap.Lockfile = true, phase, creds.Lockfile
	if phase == "ChampSelect" {
		var session struct {
			LocalPlayerCellID int          `json:"localPlayerCellId"`
			TheirTeam         []teamMember `json:"theirTeam"`
			MyTeam            []teamMember `json:"myTeam"`
			Bans              struct {
				MyTeamBans    []int `json:"myTeamBans"`
				TheirTeamBans []int `json:"theirTeamBans"`
			} `json:"bans"`
			Actions json.RawMessage `json:"actions"`
		}
		status, err := lcuGET(creds, "/lol-champ-select/v1/session", &session)
		if err == nil {
			for _, p := range session.TheirTeam {
				snap.Enemies = append(snap.Enemies, champID(p.ChampionID, p.ChampionPickIntent))
			}
			for _, p := range session.MyTeam {
				snap.Allies = append(snap.Allies, champID(p.ChampionID, p.ChampionPickIntent))
				if p.CellID == session.LocalPlayerCellID {
					snap.MyRole = roleShort(p.AssignedPosition)
					snap.MyChamp = champID(p.ChampionID, p.ChampionPickIntent)
				}
			}
			snap.Enemies, snap.Allies = uniqueNonzero(snap.Enemies), uniqueNonzero(snap.Allies)
			snap.AllyFocus = allyFocusFromTeam(session.MyTeam, session.LocalPlayerCellID)
			snap.Bans = uniqueNonzero(append(append([]int{}, session.Bans.MyTeamBans...), session.Bans.TheirTeamBans...))
			turn := ""
			type csAction struct {
				ActorCellID  int    `json:"actorCellId"`
				ChampionID   int    `json:"championId"`
				Completed    bool   `json:"completed"`
				IsInProgress bool   `json:"isInProgress"`
				Type         string `json:"type"`
			}
			applyActions := func(actions []csAction) {
				for _, a := range actions {
					if a.IsInProgress && a.ActorCellID == session.LocalPlayerCellID {
						t := strings.ToLower(a.Type)
						if t == "ban" || t == "pick" {
							turn = t
						}
					}
					if a.Completed && strings.EqualFold(a.Type, "ban") && a.ChampionID != 0 {
						snap.Bans = append(snap.Bans, a.ChampionID)
					}
				}
			}
			var grouped [][]csAction
			if json.Unmarshal(session.Actions, &grouped) == nil {
				for _, g := range grouped {
					applyActions(g)
				}
			} else {
				var flat []csAction
				if json.Unmarshal(session.Actions, &flat) == nil {
					applyActions(flat)
				}
			}
			snap.Bans = uniqueNonzero(snap.Bans)
			d := buildDraftAdvice(snap.Enemies, snap.Allies, snap.Bans, snap.MyRole, snap.MyChamp, turn)
			snap.Draft = &d
		} else if status != 404 {
			snap.Error = err.Error()
		}
	}
	if snap.Connected {
		kickHarvest(creds, phase)
		attachDraftScout(&snap)
	}
	return storeSnapshot(snap)
}

func storeSnapshot(snap Snapshot) Snapshot {
	snapMu.Lock()
	snapCache, snapAt = snap, time.Now()
	snapMu.Unlock()
	return snap
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}

func apiStatus(w http.ResponseWriter, r *http.Request) {
	s := getSnapshot()
	u := updateSnapshot()
	s.App, s.Update = appVersion, &u
	writeJSON(w, s)
}

func apiCards(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("ids"))
	if raw == "" {
		writeJSON(w, []ChampionCard{})
		return
	}
	cards := []ChampionCard{}
	_ = ensureIndex()
	var slugs []string
	for _, s := range strings.Split(raw, ",") {
		id, _ := strconv.Atoi(strings.TrimSpace(s))
		if id == 0 {
			continue
		}
		dragon.mu.Lock()
		if base, ok := dragon.byKey[id]; ok {
			slugs = append(slugs, base.ID)
		}
		dragon.mu.Unlock()
	}
	prefetchPassiveCDs(slugs)
	for _, s := range strings.Split(raw, ",") {
		id, _ := strconv.Atoi(strings.TrimSpace(s))
		if id == 0 {
			continue
		}
		if c, err := cardByKey(id); err == nil {
			cards = append(cards, c)
		}
		if len(cards) == 10 {
			break
		}
	}
	writeJSON(w, cards)
}

func apiSearch(w http.ResponseWriter, r *http.Request) {
	items := searchChampions(r.URL.Query().Get("q"))
	type result struct {
		ID    string `json:"id"`
		Key   int    `json:"key"`
		Name  string `json:"name"`
		Title string `json:"title"`
	}
	out := []result{}
	for _, c := range items {
		k, _ := strconv.Atoi(c.Key)
		out = append(out, result{ID: c.ID, Key: k, Name: c.Name, Title: c.Title})
	}
	writeJSON(w, out)
}

func apiQuit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "close")
	w.WriteHeader(http.StatusNoContent)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		terminateSelf()
	}()
}

var ddragonPathRe = regexp.MustCompile(`(?i)^[0-9]+(?:\.[0-9]+)+/(?:img|data)/[A-Za-z0-9._%/-]+$`)
var ddragonPerkRe = regexp.MustCompile(`(?i)^img/perk-images/[A-Za-z0-9._/-]+$`)

func ddragonPathOK(p string) bool {
	p = strings.TrimPrefix(p, "/")
	if p == "" || strings.Contains(p, "..") || strings.Contains(p, "//") {
		return false
	}
	return ddragonPathRe.MatchString(p) || ddragonPerkRe.MatchString(p)
}

// apiDDragon reverse-proxie le CDN Riot en same-origin : WebView2 (tracking
// prevention) et certains bloqueurs coupent sinon les icônes ddragon.
// Chemins versionnés (`16.15.1/img/…`) et runes sans version (`img/perk-images/…`).
func apiDDragon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "GET", http.StatusMethodNotAllowed)
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/ddragon/")
	if !ddragonPathOK(rest) {
		http.NotFound(w, r)
		return
	}
	src := "https://ddragon.leagueoflegends.com/cdn/" + rest
	req, err := http.NewRequest(r.Method, src, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	req.Header.Set("User-Agent", "LoL-CD-Scout/0.2")
	res, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer res.Body.Close()
	if ct := res.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(res.StatusCode)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = io.Copy(w, io.LimitReader(res.Body, 12<<20))
}

// hudURL renvoie l'URL de la page widget HUD (WebView2).
func hudURL() string {
	addr := listenAddr
	if ln, ok := currentListener.Load().(net.Listener); ok {
		addr = ln.Addr().String()
	}
	return "http://" + addr + "/hud"
}

func apiHUDOpen(w http.ResponseWriter, r *http.Request) {
	if !hudMayOpen() {
		http.Error(w, "HUD uniquement en partie", http.StatusForbidden)
		return
	}
	// Ne pas bloquer la requête HTTP : WebView2 Embed peut prendre plusieurs
	// secondes, et ça saturait le navigateur (Quitter / Data Dragon coincés).
	go func() { _ = hudOpen(hudURL()) }()
	writeJSON(w, map[string]any{"ok": true, "url": hudURL()})
}

func apiHUDPin(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	hudSetPin(q.Get("on") != "0", q.Get("noactivate") == "1")
	writeHUDStatus(w)
}

func apiHUDHold(w http.ResponseWriter, r *http.Request) {
	hudSetHold(r.URL.Query().Get("on") != "0")
	writeHUDStatus(w)
}

func apiHUDHits(w http.ResponseWriter, r *http.Request) {
	var rs []hudHit
	if err := json.NewDecoder(io.LimitReader(r.Body, 32<<10)).Decode(&rs); err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	hudSetHits(rs)
	writeJSON(w, map[string]any{"ok": true})
}

func apiHUDBounds(w http.ResponseWriter, r *http.Request) {
	var b struct {
		W int `json:"w"`
		H int `json:"h"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 256)).Decode(&b); err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	hudSetBounds(b.W, b.H)
	writeJSON(w, map[string]any{"ok": true})
}

func apiHUDDrag(w http.ResponseWriter, r *http.Request) {
	var req hudDragReq
	if err := json.NewDecoder(io.LimitReader(r.Body, 1024)).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.ID == "" {
		hudBeginDrag()
	} else {
		hudBeginWidgetDrag(req)
	}
	writeJSON(w, map[string]any{"ok": true})
}

func apiHUDGeom(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var g hudGeomDisk
		if err := json.NewDecoder(io.LimitReader(r.Body, 16<<10)).Decode(&g); err != nil && err != io.EOF {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		hudGeomReplace(g)
	}
	writeJSON(w, hudGeomSnapshot())
}

func apiHUDReset(w http.ResponseWriter, r *http.Request) {
	hudResetPos()
	writeJSON(w, hudGeomSnapshot())
}

func apiHUDClose(w http.ResponseWriter, r *http.Request) {
	hudCloseWindow()
	writeJSON(w, map[string]any{"ok": true})
}

func apiInput(w http.ResponseWriter, r *http.Request) {
	since, _ := strconv.ParseUint(r.URL.Query().Get("since"), 10, 64)
	tab, seq, events := inputSince(since)
	pinned, noActivate, window := hudStatus()
	out := map[string]any{
		"supported": hudSupported(), "tab": tab, "seq": seq, "events": events,
		"pinned": pinned, "noActivate": noActivate, "window": window, "hold": hudHold(),
		"tracks": hudTracksGet(),
	}
	if d, ok := hudDragLiveCopy(); ok {
		out["drag"] = d
	}
	writeJSON(w, out)
}

func apiClipboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST", http.StatusMethodNotAllowed)
		return
	}
	b, err := io.ReadAll(io.LimitReader(r.Body, 2048))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	erase, _ := strconv.Atoi(r.URL.Query().Get("erase"))
	if erase < 0 {
		erase = 0
	}
	if erase > 32 {
		erase = 32
	}
	if err := writeClipboard(strings.TrimSpace(string(b)), erase); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func writeHUDStatus(w http.ResponseWriter) {
	pinned, noActivate, window := hudStatus()
	writeJSON(w, map[string]any{"supported": hudSupported(), "pinned": pinned, "noActivate": noActivate, "window": window, "hold": hudHold(), "dev": hudDevAllowed(), "force": hudForceOn()})
}

func apiCard(w http.ResponseWriter, r *http.Request) {
	key, _ := strconv.Atoi(r.URL.Query().Get("key"))
	if key == 0 {
		http.Error(w, "bad key", 400)
		return
	}
	c, err := cardByKey(key)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	writeJSON(w, c)
}

func openBrowser(url string) {
	time.Sleep(250 * time.Millisecond)
	if !strings.Contains(url, "?") {
		url = strings.TrimRight(url, "/") + "/?t=" + strconv.FormatInt(time.Now().Unix(), 10)
	}
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	cmd.SysProcAttr = windowsHiddenProcAttr()
	_ = cmd.Start()
}

// currentListener expose l'adresse réellement écoutée (le port peut varier si
// 27182 est pris) aux fonctions qui doivent forger une URL locale.
var currentListener atomic.Value

func main() {
	setDPIAware()
	hudForceLoad()
	if applyStagedUpdate() {
		return // une instance à jour vient d'être lancée
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		io.WriteString(w, indexHTML)
	})
	mux.HandleFunc("/hud", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		io.WriteString(w, indexHTML)
	})
	mux.HandleFunc("/logo.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(logoPNG)
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(logoICO)
	})
	mux.HandleFunc("/ddragon/", apiDDragon)
	mux.HandleFunc("/api/status", apiStatus)
	mux.HandleFunc("/api/live", apiLive)
	mux.HandleFunc("/api/voice", apiVoice)
	mux.HandleFunc("/alerts/", apiAlertArt)
	mux.HandleFunc("/api/objsfx", apiObjSFX)
	mux.HandleFunc("/api/cards", apiCards)
	mux.HandleFunc("/api/draft", apiDraft)
	mux.HandleFunc("/api/search", apiSearch)
	mux.HandleFunc("/api/card", apiCard)
	mux.HandleFunc("/api/quit", apiQuit)
	mux.HandleFunc("/api/update", apiUpdate)
	mux.HandleFunc("/api/hud", func(w http.ResponseWriter, r *http.Request) { writeHUDStatus(w) })
	mux.HandleFunc("/api/hud/open", apiHUDOpen)
	mux.HandleFunc("/api/hud/devmode", apiHUDDevMode)
	mux.HandleFunc("/api/hud/pin", apiHUDPin)
	mux.HandleFunc("/api/hud/hold", apiHUDHold)
	mux.HandleFunc("/api/hud/hits", apiHUDHits)
	mux.HandleFunc("/api/hud/bounds", apiHUDBounds)
	mux.HandleFunc("/api/hud/drag", apiHUDDrag)
	mux.HandleFunc("/api/hud/geom", apiHUDGeom)
	mux.HandleFunc("/api/hud/reset", apiHUDReset)
	mux.HandleFunc("/api/hud/close", apiHUDClose)
	mux.HandleFunc("/api/tracks", apiTracks)
	mux.HandleFunc("/api/demo", apiDemo)
	mux.HandleFunc("/api/input", apiInput)
	mux.HandleFunc("/api/clipboard", apiClipboard)
	mux.HandleFunc("/api/open", apiOpen)
	mux.HandleFunc("/api/quiz", apiQuiz)
	mux.HandleFunc("/api/idea", apiIdea)
	ln, err := listenLocal()
	if err != nil {
		return
	}
	currentListener.Store(ln)
	url := "http://" + ln.Addr().String() + "/"
	go openBrowser(url)
	go refreshDragon()
	go func() { time.Sleep(5 * time.Second); checkAndStage() }()
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 3 * time.Second}
	_ = srv.Serve(ln)
}

//go:embed index.html
var indexHTML string

// Logo embarqué (header, favicon). L'icône de l'exe Windows est dans
// rsrc_windows_amd64.syso, généré depuis branding/logo.ico.
//
//go:embed branding/logo.png
var logoPNG []byte

//go:embed branding/logo.ico
var logoICO []byte
