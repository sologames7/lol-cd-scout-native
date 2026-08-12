package main

import (
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
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const fallbackVersion = "16.15.1"

type Spell struct {
	Spell      string `json:"spell"`
	Name       string `json:"name"`
	CD         string `json:"cd"`
	Importance int    `json:"importance"`
	Note       string `json:"note"`
}

type Override struct {
	Summary      string   `json:"summary"`
	Window       string   `json:"window"`
	Important    []Spell  `json:"important"`
	Counters     []string `json:"counters,omitempty"`
	HardMatchups []string `json:"hardMatchups,omitempty"`
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
	Source       string   `json:"source"`
}

type AllyFocus struct {
	Role string `json:"role"`
	ID   int    `json:"id"`
}

type teamMember struct {
	CellID             int    `json:"cellId"`
	ChampionID         int    `json:"championId"`
	ChampionPickIntent int    `json:"championPickIntent"`
	AssignedPosition   string `json:"assignedPosition"`
}

type Snapshot struct {
	Connected bool        `json:"connected"`
	Phase     string      `json:"phase"`
	Enemies   []int       `json:"enemies"`
	Allies    []int       `json:"allies,omitempty"`
	AllyFocus []AllyFocus `json:"allyFocus,omitempty"`
	Error     string      `json:"error,omitempty"`
	Lockfile  string      `json:"lockfile,omitempty"`
	Version   string      `json:"version,omitempty"`
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
	Spells []struct {
		Name         string `json:"name"`
		CooldownBurn string `json:"cooldownBurn"`
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
var httpClient = &http.Client{Timeout: 7 * time.Second}

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
	var versions []string
	if err := jsonGET("https://ddragon.leagueoflegends.com/api/versions.json", &versions); err == nil && len(versions) > 0 {
		dragon.version = versions[0]
	} else {
		dragon.version = fallbackVersion
		if err != nil {
			dragon.lastErr = err.Error()
		}
	}
	return dragon.version
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
	dragon.index, dragon.byKey = items, byKey
	dragon.mu.Unlock()
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
		auto[letters[i]] = Spell{Spell: letters[i], Name: s.Name, CD: cooldown(s.CooldownBurn), Importance: 5}
	}
	summary := ""
	window := ""
	source := "ddragon"
	var counters, hardMatchups []string
	if m, ok := matchups[d.ID]; ok {
		counters, hardMatchups = m.Counters, m.HardMatchups
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
		for _, s := range o.Important {
			if a, ok := auto[s.Spell]; ok {
				if s.CD == "auto" || s.CD == "" {
					s.CD = a.CD
				}
				if s.Name == "" {
					s.Name = a.Name
				}
			}
			curatedByKey[s.Spell] = s
			curatedOrder = append(curatedOrder, s)
		}
	}
	// Always expose every basic ability; keep curated extras (ex: P) and notes.
	important := []Spell{}
	for _, s := range curatedOrder {
		if !isBasicSpell(s.Spell) {
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
		Counters: counters, HardMatchups: hardMatchups, Source: source,
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
	switch strings.ToLower(strings.TrimSpace(pos)) {
	case "utility":
		return "SUPP"
	case "jungle":
		return "JGL"
	case "middle":
		return "MID"
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
	defer snapMu.Unlock()
	if time.Since(snapAt) < 500*time.Millisecond && snapAt != (time.Time{}) {
		return snapCache
	}
	snap := Snapshot{Connected: false, Phase: "Disconnected", Version: getVersion()}
	creds, err := getCreds()
	if err != nil {
		snapCache, snapAt = snap, time.Now()
		return snap
	}
	var phase string
	if _, err := lcuGET(creds, "/lol-gameflow/v1/gameflow-phase", &phase); err != nil {
		snap.Phase, snap.Error = "Connection error", err.Error()
		snapCache, snapAt = snap, time.Now()
		return snap
	}
	snap.Connected, snap.Phase, snap.Lockfile = true, phase, creds.Lockfile
	if phase == "ChampSelect" {
		var session struct {
			LocalPlayerCellID int `json:"localPlayerCellId"`
			TheirTeam         []struct {
				ChampionID         int `json:"championId"`
				ChampionPickIntent int `json:"championPickIntent"`
			} `json:"theirTeam"`
			MyTeam            []teamMember `json:"myTeam"`
		}
		status, err := lcuGET(creds, "/lol-champ-select/v1/session", &session)
		if err == nil {
			for _, p := range session.TheirTeam {
				snap.Enemies = append(snap.Enemies, champID(p.ChampionID, p.ChampionPickIntent))
			}
			for _, p := range session.MyTeam {
				snap.Allies = append(snap.Allies, champID(p.ChampionID, p.ChampionPickIntent))
			}
			snap.Enemies, snap.Allies = uniqueNonzero(snap.Enemies), uniqueNonzero(snap.Allies)
			snap.AllyFocus = allyFocusFromTeam(session.MyTeam, session.LocalPlayerCellID)
		} else if status != 404 {
			snap.Error = err.Error()
		}
	}
	snapCache, snapAt = snap, time.Now()
	return snap
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}

func apiStatus(w http.ResponseWriter, r *http.Request) { writeJSON(w, getSnapshot()) }

func apiCards(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("ids"))
	if raw == "" {
		writeJSON(w, []ChampionCard{})
		return
	}
	cards := []ChampionCard{}
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
	w.WriteHeader(http.StatusNoContent)
	go func() { time.Sleep(100 * time.Millisecond); os.Exit(0) }()
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
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	cmd.SysProcAttr = windowsHiddenProcAttr()
	_ = cmd.Start()
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, indexHTML)
	})
	mux.HandleFunc("/api/status", apiStatus)
	mux.HandleFunc("/api/live", apiLive)
	mux.HandleFunc("/api/cards", apiCards)
	mux.HandleFunc("/api/search", apiSearch)
	mux.HandleFunc("/api/card", apiCard)
	mux.HandleFunc("/api/quit", apiQuit)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return
		}
	}
	url := "http://" + ln.Addr().String() + "/"
	go openBrowser(url)
	_ = http.Serve(ln, mux)
}

//go:embed index.html
var indexHTML string
