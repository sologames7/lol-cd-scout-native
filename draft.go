package main

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

// Noms d'affichage EN (listes LoLalytics / LoLTheory) quand l'id Riot ne suffit pas.
var riotDisplay = map[string]string{
	"AurelionSol": "Aurelion Sol",
	"Belveth":     "Bel'Veth",
	"Chogath":     "Cho'Gath",
	"DrMundo":     "Dr. Mundo",
	"JarvanIV":    "Jarvan IV",
	"Kaisa":       "Kai'Sa",
	"Khazix":      "Kha'Zix",
	"KogMaw":      "Kog'Maw",
	"KSante":      "K'Sante",
	"Leblanc":     "LeBlanc",
	"LeeSin":      "Lee Sin",
	"MasterYi":    "Master Yi",
	"MissFortune": "Miss Fortune",
	"MonkeyKing":  "Wukong",
	"Nunu":        "Nunu & Willump",
	"RekSai":      "Rek'Sai",
	"Renata":      "Renata Glasc",
	"TahmKench":   "Tahm Kench",
	"TwistedFate": "Twisted Fate",
	"Velkoz":      "Vel'Koz",
	"XinZhao":     "Xin Zhao",
}

type DraftItem struct {
	Key   int      `json:"key"`
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Icon  string   `json:"icon"`
	Role  string   `json:"role,omitempty"`
	Score int      `json:"score"`
	Why   []string `json:"why"`
}

type DraftAdvice struct {
	Role  string      `json:"role,omitempty"`
	Phase string      `json:"phase,omitempty"`
	Picks []DraftItem `json:"picks,omitempty"`
	Bans  []DraftItem `json:"bans,omitempty"`
	Note  string      `json:"note,omitempty"`
}

type draftPicked struct {
	Key  int
	ID   string
	Name string
}

var (
	draftIdxOnce sync.Once
	counterFreq  map[string]int
)

func roleShort(pos string) string {
	switch strings.ToLower(strings.TrimSpace(pos)) {
	case "utility":
		return "SUPP"
	case "jungle":
		return "JGL"
	case "middle":
		return "MID"
	case "top":
		return "TOP"
	case "bottom":
		return "BOT"
	default:
		return ""
	}
}

func riotName(id string) string {
	if n, ok := riotDisplay[id]; ok {
		return n
	}
	var b strings.Builder
	for i, r := range id {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	if b.Len() == 0 {
		return id
	}
	return b.String()
}

func normChamp(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	for _, w := range words {
		if w == "et" || w == "and" || w == "the" || w == "of" {
			continue
		}
		b.WriteString(w)
	}
	out := b.String()
	if strings.HasPrefix(out, "nunu") {
		return "nunu"
	}
	return out
}

func ensureDraftIndex() {
	draftIdxOnce.Do(func() {
		counterFreq = map[string]int{}
		for _, m := range matchups {
			for i, name := range m.Counters {
				w := 12 - i
				if w < 1 {
					w = 1
				}
				counterFreq[normChamp(name)] += w
			}
		}
	})
}

func champLists(id string) (counters, hard, syn []string) {
	if m, ok := matchups[id]; ok {
		counters, hard = m.Counters, m.HardMatchups
	}
	if s, ok := synergies[id]; ok {
		syn = s
	}
	if o, ok := overrides[id]; ok {
		if len(o.Counters) > 0 {
			counters = o.Counters
		}
		if len(o.HardMatchups) > 0 {
			hard = o.HardMatchups
		}
		if len(o.Synergies) > 0 {
			syn = o.Synergies
		}
	}
	return
}

func listHas(list []string, name string) bool {
	n := normChamp(name)
	if n == "" {
		return false
	}
	for _, x := range list {
		if normChamp(x) == n {
			return true
		}
	}
	return false
}

func playsRole(key int, role string) bool {
	if role == "" {
		return true
	}
	rs := champRoles[key]
	if len(rs) == 0 {
		return true
	}
	for _, r := range rs {
		if r == role {
			return true
		}
	}
	return false
}

func parseIDList(s string) []int {
	out := []int{}
	for _, p := range strings.Split(s, ",") {
		n, _ := strconv.Atoi(strings.TrimSpace(p))
		if n != 0 {
			out = append(out, n)
		}
	}
	return uniqueNonzero(out)
}

func pickedFromKeys(keys []int) []draftPicked {
	_ = ensureIndex()
	ensureDraftIndex()
	out := []draftPicked{}
	dragon.mu.Lock()
	defer dragon.mu.Unlock()
	for _, k := range keys {
		if k == 0 {
			continue
		}
		c, ok := dragon.byKey[k]
		if !ok {
			continue
		}
		out = append(out, draftPicked{Key: k, ID: c.ID, Name: c.Name})
	}
	return out
}

func draftItemFromKey(key int, role string, score int, why []string) DraftItem {
	dragon.mu.Lock()
	c, ok := dragon.byKey[key]
	dragon.mu.Unlock()
	if !ok {
		return DraftItem{}
	}
	if len(why) > 3 {
		why = why[:3]
	}
	return DraftItem{
		Key: key, ID: c.ID, Name: c.Name, Icon: c.Image.Full,
		Role: role, Score: score, Why: why,
	}
}

func buildDraftAdvice(enemies, allies, bans []int, myRole string, myChamp int, phase string) DraftAdvice {
	_ = ensureIndex()
	ensureDraftIndex()
	adv := DraftAdvice{Role: myRole, Phase: phase}
	taken := map[int]bool{}
	for _, k := range append(append(append([]int{}, enemies...), allies...), bans...) {
		if k != 0 {
			taken[k] = true
		}
	}
	enemyP := pickedFromKeys(enemies)
	allyP := pickedFromKeys(allies)
	if len(enemyP) == 0 && len(allyP) == 0 {
		adv.Note = "En attente des premiers picks pour affiner."
	} else if myRole == "" {
		adv.Note = "Bans selon la compo — le rôle n'est pas encore connu."
	} else {
		adv.Note = "Pour ton " + myRole
		if phase == "ban" {
			adv.Note += " · à toi de bannir"
		} else if phase == "pick" {
			adv.Note += " · à toi de pick"
		}
	}

	type scored struct {
		key   int
		score int
		why   []string
		role  string
	}
	pushWhy := func(why []string, s string) []string {
		for _, w := range why {
			if w == s {
				return why
			}
		}
		return append(why, s)
	}

	// ---- Bans : menaces pour nos alliés, puis counters génériques.
	banHits := []scored{}
	dragon.mu.Lock()
	all := append([]champIndexItem(nil), dragon.index...)
	dragon.mu.Unlock()
	for _, c := range all {
		key, _ := strconv.Atoi(c.Key)
		if key == 0 || taken[key] {
			continue
		}
		score := 0
		var why []string
		name := riotName(c.ID)
		for _, a := range allyP {
			if a.Key == myChamp {
				continue
			}
			ct, hard, _ := champLists(a.ID)
			if listHas(ct, name) || listHas(ct, c.Name) {
				score += 6
				why = pushWhy(why, "contre "+a.Name)
			} else if listHas(hard, name) || listHas(hard, c.Name) {
				score += 3
				why = pushWhy(why, "matchup dur "+a.Name)
			}
		}
		for _, e := range enemyP {
			_, _, syn := champLists(e.ID)
			if listHas(syn, name) || listHas(syn, c.Name) {
				score += 2
				why = pushWhy(why, "synergie ennemi "+e.Name)
			}
		}
		if myRole != "" && playsRole(key, myRole) {
			score += 1
		}
		score += counterFreq[normChamp(name)] / 25
		if score <= 0 {
			continue
		}
		if len(why) == 0 {
			why = []string{"priorité de ban"}
		}
		banHits = append(banHits, scored{key, score, why, ""})
	}
	sort.Slice(banHits, func(i, j int) bool {
		if banHits[i].score == banHits[j].score {
			return banHits[i].key < banHits[j].key
		}
		return banHits[i].score > banHits[j].score
	})
	for i, h := range banHits {
		if i >= 5 {
			break
		}
		it := draftItemFromKey(h.key, h.role, h.score, h.why)
		if it.Key != 0 {
			adv.Bans = append(adv.Bans, it)
		}
	}

	// ---- Picks : counters des ennemis + synergies des alliés, filtrés au rôle.
	if myRole == "" {
		return adv
	}
	pickHits := []scored{}
	for _, c := range all {
		key, _ := strconv.Atoi(c.Key)
		if key == 0 || taken[key] || !playsRole(key, myRole) {
			continue
		}
		ct, hard, syn := champLists(c.ID)
		score := 0
		var why []string
		selfName := riotName(c.ID)
		for _, e := range enemyP {
			ect, ehard, _ := champLists(e.ID)
			if listHas(ect, selfName) || listHas(ect, c.Name) {
				score += 5
				why = pushWhy(why, "counter "+e.Name)
			} else if listHas(ehard, selfName) || listHas(ehard, c.Name) {
				score += 2
				why = pushWhy(why, "difficile pour "+e.Name)
			}
			if listHas(ct, e.Name) || listHas(ct, riotName(e.ID)) {
				score -= 5
			} else if listHas(hard, e.Name) || listHas(hard, riotName(e.ID)) {
				score -= 2
			}
		}
		for _, a := range allyP {
			if a.Key == myChamp || a.Key == key {
				continue
			}
			_, _, asyn := champLists(a.ID)
			mutual := (listHas(syn, a.Name) || listHas(syn, riotName(a.ID))) &&
				(listHas(asyn, selfName) || listHas(asyn, c.Name))
			oneWay := listHas(syn, a.Name) || listHas(syn, riotName(a.ID)) ||
				listHas(asyn, selfName) || listHas(asyn, c.Name)
			if mutual {
				score += 6
				why = pushWhy(why, "duo "+a.Name)
			} else if oneWay {
				score += 3
				why = pushWhy(why, "synergie "+a.Name)
			}
		}
		if score <= 0 {
			continue
		}
		if len(why) == 0 {
			why = []string{"bon dans ce draft"}
		}
		pickHits = append(pickHits, scored{key, score, why, myRole})
	}
	sort.Slice(pickHits, func(i, j int) bool {
		if pickHits[i].score == pickHits[j].score {
			return pickHits[i].key < pickHits[j].key
		}
		return pickHits[i].score > pickHits[j].score
	})
	for i, h := range pickHits {
		if i >= 6 {
			break
		}
		it := draftItemFromKey(h.key, h.role, h.score, h.why)
		if it.Key != 0 {
			adv.Picks = append(adv.Picks, it)
		}
	}
	return adv
}

func apiDraft(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	enemies := parseIDList(q.Get("enemies"))
	allies := parseIDList(q.Get("allies"))
	bans := parseIDList(q.Get("bans"))
	role := strings.ToUpper(strings.TrimSpace(q.Get("role")))
	my, _ := strconv.Atoi(q.Get("me"))
	phase := strings.ToLower(strings.TrimSpace(q.Get("phase")))
	if len(enemies) == 0 && len(allies) == 0 && q.Get("enemies") == "" && q.Get("allies") == "" {
		s := getSnapshot()
		enemies, allies, bans = s.Enemies, s.Allies, s.Bans
		if role == "" {
			role = s.MyRole
		}
		if my == 0 {
			my = s.MyChamp
		}
		if phase == "" && s.Draft != nil {
			phase = s.Draft.Phase
		}
	}
	writeJSON(w, buildDraftAdvice(enemies, allies, bans, role, my, phase))
}
