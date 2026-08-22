package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Plan de jeu : table de draft → Claude Code / Codex / clé API → texte + audio.

type GamePlan struct {
	Ready    bool          `json:"ready"`
	ID       string        `json:"id,omitempty"`
	Title    string        `json:"title,omitempty"`
	Role     string        `json:"role,omitempty"`
	Wait     string        `json:"wait,omitempty"`
	Status   string        `json:"status,omitempty"` // wait, nollm, loading, ready, error
	LLM      bool          `json:"llm"`
	Table    string        `json:"table,omitempty"`
	Me       PlanSide      `json:"me,omitempty"`
	Them     PlanSide      `json:"them,omitempty"`
	Sections []PlanSection `json:"sections,omitempty"`
	Speak    string        `json:"speak,omitempty"`
}

type PlanSide struct {
	Name     string   `json:"name,omitempty"`
	Key      int      `json:"key,omitempty"`
	Icon     string   `json:"icon,omitempty"`
	Sums     []string `json:"sums,omitempty"`
	Keystone string   `json:"keystone,omitempty"`
}

type PlanSection struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type planFighter struct {
	Key      int
	Name     string
	ID       string
	Icon     string
	Role     string
	Tags     []string
	Range    int
	Sums     []string
	Keystone string
	KeyID    int
	Window   string
}

type planRow struct {
	Side     string
	Role     string
	Name     string
	Sums     []string
	Keystone string
	Range    int
	Tags     []string
	Window   string
	Me       bool
	Lane     bool
}

type planInput struct {
	Role        string
	Patch       string
	Me          planFighter
	Them        planFighter
	EnemyJungle string
	AllyJungle  string
	ARAM        bool
	Allies      []planRow
	Enemies     []planRow
	Bans        []string
}

var lcuSummShort = map[int]string{
	1: "Cleanse", 3: "Exhaust", 4: "Flash", 6: "Ghost",
	7: "Heal", 11: "Smite", 12: "TP", 13: "Clarity",
	14: "Ignite", 21: "Barrier", 32: "Mark",
}

func apiPlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "GET", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, resolvePlan(r))
}

func resolvePlan(r *http.Request) GamePlan {
	if r.URL.Query().Get("retry") == "1" {
		clearPlanCache()
	}
	if q := r.URL.Query(); q.Get("me") != "" || q.Get("demo") != "" {
		return buildPlan(planFromQuery(q))
	}
	if in, ok := planFromLive(); ok {
		return buildPlan(in)
	}
	if in, ok := planFromChampSelect(); ok {
		return buildPlan(in)
	}
	return GamePlan{Status: "wait", Wait: "Lance une champ select : le plan sort dès que ton pick et le matchup de lane sont visibles."}
}

func planFromQuery(q map[string][]string) planInput {
	get := func(k string) string { return strings.TrimSpace(strings.Join(q[k], "")) }
	role := strings.ToUpper(get("role"))
	if role == "" {
		role = "TOP"
	}
	in := planInput{Role: role, ARAM: get("aram") == "1"}
	if get("demo") != "" && get("me") == "" {
		in.Role = "TOP"
		in.Patch = getVersion()
		in.Me = fighterFromKey(21, "TOP", []string{"Ghost", "Flash"}, 8369, "Premier coup")
		in.Them = fighterFromKey(150, "TOP", []string{"TP", "Flash"}, 8021, "Jeu de jambes")
		if in.Me.Name == "" {
			in.Me.Name, in.Me.ID, in.Me.Range, in.Me.Tags = "Miss Fortune", "MissFortune", 550, []string{"Marksman"}
		}
		if in.Them.Name == "" {
			in.Them.Name, in.Them.ID, in.Them.Range, in.Them.Tags = "Gnar", "Gnar", 400, []string{"Fighter"}
		}
		in.AllyJungle, in.EnemyJungle = "Lee Sin", "Elise"
		in.Allies = []planRow{
			rowFromFighter(in.Me, "ALLIÉ", true, true),
			{Side: "ALLIÉ", Role: "JGL", Name: "Lee Sin", Sums: []string{"Flash", "Smite"}},
			{Side: "ALLIÉ", Role: "MID", Name: "Orianna", Sums: []string{"Flash", "Ignite"}},
			{Side: "ALLIÉ", Role: "BOT", Name: "Jinx", Sums: []string{"Flash", "Heal"}},
			{Side: "ALLIÉ", Role: "SUPP", Name: "Leona", Sums: []string{"Flash", "Ignite"}},
		}
		in.Enemies = []planRow{
			rowFromFighter(in.Them, "ENNEMI", false, true),
			{Side: "ENNEMI", Role: "JGL", Name: "Elise", Sums: []string{"Flash", "Smite"}},
			{Side: "ENNEMI", Role: "MID", Name: "Syndra", Sums: []string{"Flash", "Ignite"}},
			{Side: "ENNEMI", Role: "BOT", Name: "Caitlyn", Sums: []string{"Flash", "Heal"}},
			{Side: "ENNEMI", Role: "SUPP", Name: "Nautilus", Sums: []string{"Flash", "Ignite"}},
		}
		in.Bans = []string{"Yasuo", "Irelia", "K'Sante", "Azir"}
		return in
	}
	in.Me = fighterFromKey(atoi(get("me")), role, parseSumList(get("mysums")), atoi(get("mykey")), get("myrune"))
	in.Them = fighterFromKey(atoi(get("them")), role, parseSumList(get("theirsums")), atoi(get("theirkey")), get("theirrune"))
	in.EnemyJungle = get("jgl")
	in.AllyJungle = get("allyjgl")
	return in
}

func planFromLive() (planInput, bool) {
	state := getLiveState()
	if !state.Active {
		if d, ok := demoLiveCopy(); ok && d.Active {
			state = d
		} else {
			return planInput{}, false
		}
	}
	var me LivePlayer
	for _, p := range state.Allies {
		if p.IsMe {
			me = p
			break
		}
	}
	if me.Key == 0 && len(state.Allies) > 0 {
		me = state.Allies[0]
	}
	if me.Key == 0 {
		return planInput{}, false
	}
	role := roleShort(me.Position)
	if role == "" {
		role = "MID"
	}
	them := liveLaneOpponent(state.Enemies, me.Position)
	in := planInput{
		Role:        role,
		Patch:       getVersion(),
		Me:          fighterFromLive(me, role),
		Them:        fighterFromLive(them, role),
		EnemyJungle: liveNameAt(state.Enemies, "JUNGLE"),
		AllyJungle:  liveNameAt(state.Allies, "JUNGLE"),
		ARAM:        strings.EqualFold(state.GameMode, "ARAM") || !summonersRift(state.GameMode),
		Allies:      rowsFromLive(state.Allies, "ALLIÉ", me.Position),
		Enemies:     rowsFromLive(state.Enemies, "ENNEMI", me.Position),
	}
	return in, in.Me.Key != 0
}

func liveLaneOpponent(enemies []LivePlayer, pos string) LivePlayer {
	if pos != "" && !strings.EqualFold(pos, "JUNGLE") {
		for _, p := range enemies {
			if strings.EqualFold(p.Position, pos) {
				return p
			}
		}
	}
	if len(enemies) > 0 {
		return enemies[0]
	}
	return LivePlayer{}
}

func liveNameAt(list []LivePlayer, pos string) string {
	for _, p := range list {
		if strings.EqualFold(p.Position, pos) && p.Champion != "" {
			return p.Champion
		}
	}
	return ""
}

func fighterFromLive(p LivePlayer, role string) planFighter {
	f := planFighter{
		Key:      p.Key,
		Name:     p.Champion,
		Role:     role,
		Sums:     shortsFromLiveSums(p.Summoners),
		Keystone: keystoneShort(p.Runes.Keystone.ID, p.Runes.Keystone.Name),
		KeyID:    p.Runes.Keystone.ID,
	}
	fillChamp(&f)
	return f
}

func rowFromFighter(f planFighter, side string, me, lane bool) planRow {
	return planRow{
		Side: side, Role: f.Role, Name: f.Name, Sums: f.Sums, Keystone: f.Keystone,
		Range: f.Range, Tags: f.Tags, Window: f.Window, Me: me, Lane: lane,
	}
}

func rowsFromLive(list []LivePlayer, side, myPos string) []planRow {
	out := make([]planRow, 0, len(list))
	for _, p := range list {
		role := roleShort(p.Position)
		f := fighterFromLive(p, role)
		out = append(out, planRow{
			Side: side, Role: role, Name: f.Name, Sums: f.Sums, Keystone: f.Keystone,
			Range: f.Range, Tags: f.Tags, Window: f.Window, Me: p.IsMe,
			Lane: myPos != "" && strings.EqualFold(p.Position, myPos),
		})
	}
	return out
}

func rowsFromCS(team []teamMember, side string, localCell int, myPos string, myKeyID int) []planRow {
	out := make([]planRow, 0, len(team))
	for _, p := range team {
		id := champID(p.ChampionID, p.ChampionPickIntent)
		if id == 0 {
			continue
		}
		role := roleShort(p.AssignedPosition)
		keyID := 0
		if p.CellID == localCell {
			keyID = myKeyID
		}
		f := fighterFromKey(id, role, lcuSums(p.Spell1ID, p.Spell2ID), keyID, "")
		out = append(out, planRow{
			Side: side, Role: role, Name: f.Name, Sums: f.Sums, Keystone: f.Keystone,
			Range: f.Range, Tags: f.Tags, Window: f.Window,
			Me:    localCell >= 0 && p.CellID == localCell,
			Lane:  myPos != "" && strings.EqualFold(p.AssignedPosition, myPos),
		})
	}
	return out
}

func champNames(ids []int) []string {
	out := []string{}
	for _, id := range ids {
		if _, name, _, _, _ := champMeta(id); name != "" {
			out = append(out, name)
		}
	}
	return out
}

func planFromChampSelect() (planInput, bool) {
	creds, err := getCreds()
	if err != nil {
		return planInput{}, false
	}
	var phase string
	if _, err := lcuGET(creds, "/lol-gameflow/v1/gameflow-phase", &phase); err != nil {
		return planInput{}, false
	}
	if phase != "ChampSelect" && phase != "GameStart" && phase != "ReadyCheck" {
		return planInput{}, false
	}
	var session struct {
		LocalPlayerCellID int          `json:"localPlayerCellId"`
		TheirTeam         []teamMember `json:"theirTeam"`
		MyTeam            []teamMember `json:"myTeam"`
		Bans              struct {
			MyTeamBans    []int `json:"myTeamBans"`
			TheirTeamBans []int `json:"theirTeamBans"`
		} `json:"bans"`
	}
	if _, err := lcuGET(creds, "/lol-champ-select/v1/session", &session); err != nil {
		return planInput{}, false
	}
	var me teamMember
	for _, p := range session.MyTeam {
		if p.CellID == session.LocalPlayerCellID {
			me = p
			break
		}
	}
	myChamp := champID(me.ChampionID, me.ChampionPickIntent)
	if myChamp == 0 {
		return planInput{}, false
	}
	role := roleShort(me.AssignedPosition)
	them, themChamp := laneOpponentCS(session.TheirTeam, me.AssignedPosition, myChamp)
	keyID := myKeystone(creds)
	in := planInput{
		Role:        role,
		Patch:       getVersion(),
		Me:          fighterFromKey(myChamp, role, lcuSums(me.Spell1ID, me.Spell2ID), keyID, ""),
		Them:        fighterFromKey(themChamp, role, lcuSums(them.Spell1ID, them.Spell2ID), 0, ""),
		EnemyJungle: champNameAt(session.TheirTeam, "jungle"),
		AllyJungle:  champNameAtSkip(session.MyTeam, "jungle", session.LocalPlayerCellID),
		Allies:      rowsFromCS(session.MyTeam, "ALLIÉ", session.LocalPlayerCellID, me.AssignedPosition, keyID),
		Enemies:     rowsFromCS(session.TheirTeam, "ENNEMI", -1, me.AssignedPosition, 0),
		Bans:        champNames(uniqueNonzero(append(append([]int{}, session.Bans.MyTeamBans...), session.Bans.TheirTeamBans...))),
	}
	if in.Role == "" {
		in.Role = guessRole(myChamp)
	}
	return in, in.Me.Key != 0
}

func laneOpponentCS(their []teamMember, myPos string, myChamp int) (teamMember, int) {
	if myPos != "" && !strings.EqualFold(myPos, "jungle") {
		for _, p := range their {
			if strings.EqualFold(p.AssignedPosition, myPos) {
				return p, champID(p.ChampionID, p.ChampionPickIntent)
			}
		}
	}
	// Blind / rôles masqués : premier ennemi dont le rôle Meraki colle.
	want := roleShort(myPos)
	if want == "" {
		want = guessRole(myChamp)
	}
	for _, p := range their {
		id := champID(p.ChampionID, p.ChampionPickIntent)
		if id == 0 {
			continue
		}
		if roleHas(id, want) {
			return p, id
		}
	}
	for _, p := range their {
		if id := champID(p.ChampionID, p.ChampionPickIntent); id != 0 {
			return p, id
		}
	}
	return teamMember{}, 0
}

func champNameAt(team []teamMember, pos string) string {
	for _, p := range team {
		if strings.EqualFold(p.AssignedPosition, pos) {
			if id := champID(p.ChampionID, p.ChampionPickIntent); id != 0 {
				_, name, _, _, _ := champMeta(id)
				return name
			}
		}
	}
	return ""
}

func champNameAtSkip(team []teamMember, pos string, skipCell int) string {
	for _, p := range team {
		if p.CellID == skipCell {
			continue
		}
		if strings.EqualFold(p.AssignedPosition, pos) {
			if id := champID(p.ChampionID, p.ChampionPickIntent); id != 0 {
				_, name, _, _, _ := champMeta(id)
				return name
			}
		}
	}
	return ""
}

func roleHas(key int, role string) bool {
	for _, r := range champRoles[key] {
		if r == role {
			return true
		}
	}
	return false
}

func guessRole(key int) string {
	if rs := champRoles[key]; len(rs) > 0 {
		return rs[0]
	}
	return "MID"
}

func myKeystone(creds *lcuCreds) int {
	var page struct {
		SelectedPerkIds []int `json:"selectedPerkIds"`
	}
	if _, err := lcuGET(creds, "/lol-perks/v1/currentpage", &page); err != nil {
		return 0
	}
	if len(page.SelectedPerkIds) == 0 {
		return 0
	}
	return page.SelectedPerkIds[0]
}

func fighterFromKey(key int, role string, sums []string, keyID int, runeName string) planFighter {
	f := planFighter{Key: key, Role: role, Sums: sums, KeyID: keyID, Keystone: keystoneShort(keyID, runeName)}
	fillChamp(&f)
	if f.Keystone == "" && runeName != "" {
		f.Keystone = runeName
	}
	return f
}

func fillChamp(f *planFighter) {
	if f.Key == 0 {
		return
	}
	id, name, icon, tags, rng := champMeta(f.Key)
	f.ID, f.Icon, f.Tags, f.Range = id, icon, tags, rng
	if f.Name == "" {
		f.Name = name
	}
	if o, ok := overrides[id]; ok {
		f.Window = o.Window
	}
}

func champMeta(key int) (id, name, icon string, tags []string, rng int) {
	_ = ensureIndex()
	dragon.mu.Lock()
	defer dragon.mu.Unlock()
	b, ok := dragon.byKey[key]
	if !ok {
		return
	}
	return b.ID, b.Name, b.Image.Full, append([]string{}, b.Tags...), int(b.Stats.AttackRange + 0.5)
}

func lcuSums(a, b int) []string {
	out := []string{}
	for _, id := range []int{a, b} {
		if s := lcuSummShort[id]; s != "" {
			out = append(out, s)
		}
	}
	return out
}

func parseSumList(s string) []string {
	out := []string{}
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if n, err := strconv.Atoi(p); err == nil {
			if short := lcuSummShort[n]; short != "" {
				out = append(out, short)
				continue
			}
		}
		out = append(out, summShort(p, p, ""))
	}
	return out
}

func shortsFromLiveSums(sums []LiveSummoner) []string {
	out := []string{}
	for _, s := range sums {
		if x := summShort(s.Slug, s.Name, s.Kind); x != "" {
			out = append(out, x)
		}
	}
	return out
}

func summShort(slug, name, kind string) string {
	blob := strings.ToLower(slug + " " + name + " " + kind)
	switch {
	case strings.Contains(blob, "hexflash"):
		return "Hexflash"
	case strings.Contains(blob, "flash"):
		return "Flash"
	case strings.Contains(blob, "haste"), strings.Contains(blob, "ghost"), strings.Contains(blob, "spectre"):
		return "Ghost"
	case strings.Contains(blob, "teleport"), strings.Contains(blob, "téléport"):
		return "TP"
	case strings.Contains(blob, "dot"), strings.Contains(blob, "ignite"), strings.Contains(blob, "brûl"), strings.Contains(blob, "brul"):
		return "Ignite"
	case strings.Contains(blob, "exhaust"), strings.Contains(blob, "fatigue"):
		return "Exhaust"
	case strings.Contains(blob, "heal"), strings.Contains(blob, "soin"):
		return "Heal"
	case strings.Contains(blob, "barrier"), strings.Contains(blob, "barrière"), strings.Contains(blob, "barriere"):
		return "Barrier"
	case strings.Contains(blob, "boost"), strings.Contains(blob, "cleanse"), strings.Contains(blob, "purge"):
		return "Cleanse"
	case strings.Contains(blob, "smite"), strings.Contains(blob, "châtiment"), strings.Contains(blob, "chatiment"):
		return "Smite"
	default:
		return strings.TrimSpace(name)
	}
}

func keystoneShort(id int, name string) string {
	switch id {
	case 8005:
		return "PTA"
	case 8008:
		return "Tempo mortel"
	case 8010:
		return "Conquérant"
	case 8021:
		return "Jeu de jambes"
	case 8112:
		return "Électrocute"
	case 8128:
		return "Dark Harvest"
	case 9923:
		return "Grêle de lames"
	case 8214:
		return "Aery"
	case 8229:
		return "Comète"
	case 8230:
		return "Phase Rush"
	case 8437:
		return "Emprise"
	case 8439:
		return "Aftershock"
	case 8465:
		return "Gardien"
	case 8351:
		return "Augure glaciaire"
	case 8360:
		return "Grimoire"
	case 8369:
		return "Premier coup"
	}
	if name != "" {
		return name
	}
	if info := runeTable()[id]; info.Name != "" {
		return info.Name
	}
	return ""
}

func cfgLine(f planFighter) string {
	bits := []string{}
	if f.Keystone != "" {
		bits = append(bits, f.Keystone)
	}
	bits = append(bits, f.Sums...)
	if len(bits) == 0 {
		return f.Name
	}
	return f.Name + " " + strings.Join(bits, " + ")
}

func joinPlus(f planFighter) string {
	bits := []string{}
	if f.Keystone != "" {
		bits = append(bits, f.Keystone)
	}
	bits = append(bits, f.Sums...)
	if len(bits) == 0 {
		return "config inconnue"
	}
	return strings.Join(bits, " + ")
}

func sideJSON(f planFighter) PlanSide {
	return PlanSide{Name: f.Name, Key: f.Key, Icon: f.Icon, Sums: f.Sums, Keystone: f.Keystone}
}

func planID(in planInput) string {
	return fmt.Sprintf("%d|%s|%d|%s|%s|%d|%d|%s|%s", in.Me.Key, in.Role, in.Them.Key, strings.Join(in.Me.Sums, ","), strings.Join(in.Them.Sums, ","), in.Me.KeyID, in.Them.KeyID, strings.Join(in.Bans, ","), in.EnemyJungle)
}

func skeletonPlan(in planInput) GamePlan {
	title := fmt.Sprintf("%s %s · %s  vs  %s", in.Me.Name, strings.ToLower(in.Role), joinPlus(in.Me), cfgLine(in.Them))
	if in.Role == "" {
		title = cfgLine(in.Me) + " vs " + cfgLine(in.Them)
	}
	return GamePlan{
		ID: planID(in), Title: title, Role: in.Role,
		Me: sideJSON(in.Me), Them: sideJSON(in.Them),
		Table: draftTable(in),
	}
}

func buildPlan(in planInput) GamePlan {
	if in.Me.Key == 0 {
		return GamePlan{Status: "wait", Wait: "En attente de ton pick…"}
	}
	if in.Them.Key == 0 {
		p := GamePlan{Status: "wait", Wait: "Ton pick est là. En attente du matchup de lane…", Role: in.Role, Me: sideJSON(in.Me)}
		p.Title = cfgLine(in.Me)
		if in.Role != "" {
			p.Title = in.Me.Name + " " + strings.ToLower(in.Role)
		}
		return p
	}
	return finishPlanLLM(skeletonPlan(in), in)
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

