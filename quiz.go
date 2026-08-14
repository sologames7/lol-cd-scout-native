package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"math/rand/v2"
)

const quizRoundTotal = 10

type quizChoice struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
}

type quizQuestion struct {
	ID      string       `json:"id"`
	Kind    string       `json:"kind"`
	Prompt  string       `json:"prompt"`
	Sub     string       `json:"sub,omitempty"`
	Icon    string       `json:"icon,omitempty"`
	Choices []quizChoice `json:"choices"`
	Round   int          `json:"round"`
	Total   int          `json:"total"`
	Score   int          `json:"score"`
	Streak  int          `json:"streak"`
	Patch   string       `json:"patch,omitempty"`
	Done    bool         `json:"done,omitempty"`
}

type quizPending struct {
	id      string
	correct string
	answer  string
	explain string
}

type quizGrade struct {
	Correct  bool   `json:"correct"`
	Pick     string `json:"pick"`
	AnswerID string `json:"answerId"`
	Answer   string `json:"answer"`
	Explain  string `json:"explain"`
	Score    int    `json:"score"`
	Streak   int    `json:"streak"`
	Round    int    `json:"round"`
	Total    int    `json:"total"`
	Done     bool   `json:"done"`
}

type quizState struct {
	mu      sync.Mutex
	asked   int
	score   int
	streak  int
	pending *quizPending
	q       *quizQuestion
	seen    map[string]bool
	done    bool
}

var quizSess quizState

type quizBuilt struct {
	key     string
	kind    string
	prompt  string
	sub     string
	icon    string
	correct quizChoice
	pool    []quizChoice
	explain string
}

type quizSpell struct {
	Champ  champDetail
	Letter string
	Name   string
	Icon   string
	Desc   string
	Cost   []float64
	Range  []float64
	CDs    []float64
}

func quizLetters() []string { return []string{"Q", "W", "E", "R"} }

func listQuizSpells(d champDetail) []quizSpell {
	letters := quizLetters()
	out := make([]quizSpell, 0, 4)
	for i, s := range d.Spells {
		if i >= len(letters) {
			break
		}
		if strings.TrimSpace(s.Name) == "" {
			continue
		}
		out = append(out, quizSpell{
			Champ:  d,
			Letter: letters[i],
			Name:   s.Name,
			Icon:   spellIconPath(s.Image.Full),
			Desc:   plainText(s.Description),
			Cost:   parseCDBurn(s.CostBurn),
			Range:  parseCDBurn(s.RangeBurn),
			CDs:    parseCDBurn(s.CooldownBurn),
		})
	}
	return out
}

func (s quizSpell) cdAt(rank int) (float64, bool) {
	if len(s.CDs) == 0 || rank < 1 || rank > len(s.CDs) {
		return 0, false
	}
	v := s.CDs[rank-1]
	if v <= 0 {
		return 0, false
	}
	return v, true
}

func (s quizSpell) label() string {
	return s.Champ.Name + " · " + s.Letter + " — " + s.Name
}

func fmtQuizCD(v float64) string { return compactSec(v) + "s" }

func fmtQuizCurve(vals []float64) string {
	s := formatPassiveCD(vals)
	if s == "—" {
		return ""
	}
	return s
}

func quizNear(a, b float64) bool { return math.Abs(a-b) < 0.051 }

func quizRankLabel(rank, n int) string {
	if n > 0 && rank == n {
		return fmt.Sprintf("rang %d (max)", rank)
	}
	return fmt.Sprintf("rang %d", rank)
}

func pickQuizRank(n int, preferScale bool, scales bool) int {
	if n <= 1 {
		return 1
	}
	if preferScale && scales && n >= 3 && rand.Float64() < 0.45 {
		return 1 + rand.IntN(n-2) + 1 // rang milieu (2..n-1)
	}
	if rand.Float64() < 0.55 {
		return 1
	}
	return n
}

func cdDistractors(correct float64, related []float64) []float64 {
	cands := append([]float64{}, related...)
	deltas := []float64{1, 2, 3, 4, 5, 6, 8, 10}
	if correct >= 50 {
		deltas = []float64{10, 20, 30, 40, 60, 80}
		cands = append(cands, 80, 90, 100, 110, 120, 140, 150, 160, 180, 200, 240, 300)
	} else {
		cands = append(cands, 4, 5, 6, 7, 8, 9, 10, 11, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30)
	}
	for _, d := range deltas {
		cands = append(cands, correct+d, correct-d)
	}
	out := make([]float64, 0, 8)
	seen := map[string]bool{compactSec(correct): true}
	for _, v := range cands {
		if v <= 0 || quizNear(v, correct) {
			continue
		}
		k := compactSec(v)
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, v)
	}
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func choicesFromCDs(correct float64, related []float64) []quizChoice {
	d := cdDistractors(correct, related)
	if len(d) < 3 {
		return nil
	}
	pool := make([]quizChoice, 0, len(d))
	for _, v := range d {
		pool = append(pool, quizChoice{Label: fmtQuizCD(v)})
	}
	return pool
}

func finishQuizChoices(correct quizChoice, pool []quizChoice, n int) []quizChoice {
	out := []quizChoice{correct}
	seen := map[string]bool{strings.TrimSpace(strings.ToLower(correct.Label)): true}
	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	for _, c := range pool {
		if len(out) >= n {
			break
		}
		k := strings.TrimSpace(strings.ToLower(c.Label))
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, c)
	}
	if len(out) < n {
		return nil
	}
	for i := range out {
		out[i].ID = fmt.Sprintf("%x", rand.Uint64())
	}
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func quizChampIcon(d champDetail) string {
	if d.Image.Full == "" {
		return ""
	}
	return "champion/" + d.Image.Full
}

func quizPassiveIcon(d champDetail) string {
	if d.Passive.Image.Full == "" {
		return ""
	}
	return "passive/" + d.Passive.Image.Full
}

func nameLeaksChamp(d champDetail, name string) bool {
	n := strings.ToLower(name)
	if strings.Contains(n, strings.ToLower(d.Name)) {
		return true
	}
	if len(d.ID) >= 4 && strings.Contains(n, strings.ToLower(d.ID)) {
		return true
	}
	return false
}

func scrubQuizText(d champDetail, s string) string {
	s = strings.ReplaceAll(s, d.Name, "ce champion")
	if d.ID != "" && d.ID != d.Name {
		s = strings.ReplaceAll(s, d.ID, "ce champion")
	}
	for _, part := range strings.Fields(d.Name) {
		clean := strings.Trim(part, "&")
		if len([]rune(clean)) >= 4 {
			s = strings.ReplaceAll(s, clean, "ce champion")
		}
	}
	s = strings.TrimSpace(s)
	if len([]rune(s)) > 280 {
		r := []rune(s)
		s = strings.TrimSpace(string(r[:280])) + "…"
	}
	return s
}

func relatedCDs(sp quizSpell, all []quizSpell) []float64 {
	out := append([]float64{}, sp.CDs...)
	for _, o := range all {
		if o.Letter == sp.Letter && o.Champ.ID == sp.Champ.ID {
			continue
		}
		if v, ok := o.cdAt(1); ok {
			out = append(out, v)
		}
	}
	return out
}

func pickPoolChamp(pool []champDetail) champDetail {
	return pool[rand.IntN(len(pool))]
}

func allPoolSpells(pool []champDetail) []quizSpell {
	out := make([]quizSpell, 0, len(pool)*4)
	for _, d := range pool {
		out = append(out, listQuizSpells(d)...)
	}
	return out
}

func genCDRank(pool []champDetail) *quizBuilt {
	if len(pool) == 0 {
		return nil
	}
	preferScale := rand.Float64() < 0.75
	for try := 0; try < 12; try++ {
		d := pickPoolChamp(pool)
		spells := listQuizSpells(d)
		cands := make([]quizSpell, 0, 4)
		for _, s := range spells {
			if len(s.CDs) == 0 {
				continue
			}
			if preferScale && distinctCDCount(s.CDs) < 2 && rand.Float64() < 0.7 {
				continue
			}
			cands = append(cands, s)
		}
		if len(cands) == 0 {
			continue
		}
		s := cands[rand.IntN(len(cands))]
		rank := pickQuizRank(len(s.CDs), true, distinctCDCount(s.CDs) >= 2)
		v, ok := s.cdAt(rank)
		if !ok {
			continue
		}
		poolCh := choicesFromCDs(v, relatedCDs(s, spells))
		if len(poolCh) < 3 {
			continue
		}
		curve := fmtQuizCurve(s.CDs)
		return &quizBuilt{
			key:     "cd|" + d.ID + "|" + s.Letter + "|" + strconv.Itoa(rank),
			kind:    "CD",
			prompt:  fmt.Sprintf("Sans haste, le CD du %s de %s au %s ?", s.Letter, d.Name, quizRankLabel(rank, len(s.CDs))),
			sub:     s.Name + " · patch " + getVersion(),
			icon:    s.Icon,
			correct: quizChoice{Label: fmtQuizCD(v)},
			pool:    poolCh,
			explain: fmt.Sprintf("%s %s (%s) : %s. Au %s = %s.", d.Name, s.Letter, s.Name, curve, quizRankLabel(rank, len(s.CDs)), fmtQuizCD(v)),
		}
	}
	return nil
}

func distinctCDCount(vals []float64) int {
	seen := map[string]bool{}
	for _, v := range vals {
		if v > 0 {
			seen[compactSec(v)] = true
		}
	}
	return len(seen)
}

func genCDCurveWho(pool []champDetail) *quizBuilt {
	spells := allPoolSpells(pool)
	cands := make([]quizSpell, 0, len(spells))
	for _, s := range spells {
		if distinctCDCount(s.CDs) < 2 {
			continue
		}
		cands = append(cands, s)
	}
	if len(cands) < 4 {
		return nil
	}
	rand.Shuffle(len(cands), func(i, j int) { cands[i], cands[j] = cands[j], cands[i] })
	var pick quizSpell
	curve := ""
	for _, s := range cands {
		c := fmtQuizCurve(s.CDs)
		if c == "" {
			continue
		}
		unique := true
		for _, o := range cands {
			if o.Champ.ID == s.Champ.ID && o.Letter == s.Letter {
				continue
			}
			if fmtQuizCurve(o.CDs) == c {
				unique = false
				break
			}
		}
		if unique {
			pick, curve = s, c
			break
		}
	}
	if curve == "" {
		return nil
	}
	wrong := make([]quizChoice, 0, 8)
	for _, s := range cands {
		if s.Champ.ID == pick.Champ.ID && s.Letter == pick.Letter {
			continue
		}
		wrong = append(wrong, quizChoice{Label: s.label(), Icon: s.Icon})
	}
	return &quizBuilt{
		key:     "curve|" + pick.Champ.ID + "|" + pick.Letter,
		kind:    "CD",
		prompt:  "Sans haste, cette courbe de CD : " + curve + ". Quel sort ?",
		sub:     "Valeurs Data Dragon, pas de haste.",
		correct: quizChoice{Label: pick.label(), Icon: pick.Icon},
		pool:    wrong,
		explain: fmt.Sprintf("%s. Courbe : %s.", pick.label(), curve),
	}
}

func genWhichSpellCD(pool []champDetail) *quizBuilt {
	for try := 0; try < 14; try++ {
		d := pickPoolChamp(pool)
		spells := listQuizSpells(d)
		if len(spells) < 4 {
			continue
		}
		type hit struct {
			s quizSpell
			v float64
		}
		by := map[string][]hit{}
		uniq := make([]hit, 0, 4)
		for _, s := range spells {
			v, ok := s.cdAt(1)
			if !ok {
				continue
			}
			k := compactSec(v)
			h := hit{s, v}
			by[k] = append(by[k], h)
		}
		for _, list := range by {
			if len(list) == 1 {
				uniq = append(uniq, list[0])
			}
		}
		if len(uniq) == 0 {
			continue
		}
		h := uniq[rand.IntN(len(uniq))]
		wrong := make([]quizChoice, 0, 4)
		for _, s := range spells {
			if s.Letter == h.s.Letter {
				continue
			}
			wrong = append(wrong, quizChoice{Label: s.Letter + " — " + s.Name, Icon: s.Icon})
		}
		return &quizBuilt{
			key:     "which|" + d.ID + "|" + h.s.Letter,
			kind:    "CD",
			prompt:  fmt.Sprintf("Parmi les sorts de %s, lequel a un CD de %s au rang 1 (sans haste) ?", d.Name, fmtQuizCD(h.v)),
			sub:     d.Title,
			icon:    quizChampIcon(d),
			correct: quizChoice{Label: h.s.Letter + " — " + h.s.Name, Icon: h.s.Icon},
			pool:    wrong,
			explain: fmt.Sprintf("%s %s (%s) : %s au rang 1. Courbe %s.", d.Name, h.s.Letter, h.s.Name, fmtQuizCD(h.v), fmtQuizCurve(h.s.CDs)),
		}
	}
	return nil
}

func genUltCompare(pool []champDetail) *quizBuilt {
	type row struct {
		d champDetail
		s quizSpell
		v float64
	}
	rows := make([]row, 0, len(pool))
	for _, d := range pool {
		for _, s := range listQuizSpells(d) {
			if s.Letter != "R" {
				continue
			}
			v, ok := s.cdAt(1)
			if ok {
				rows = append(rows, row{d, s, v})
			}
			break
		}
	}
	if len(rows) < 4 {
		return nil
	}
	rand.Shuffle(len(rows), func(i, j int) { rows[i], rows[j] = rows[j], rows[i] })
	four := rows[:4]
	longest := rand.Float64() < 0.5
	best := 0
	for i := 1; i < 4; i++ {
		if longest {
			if four[i].v > four[best].v {
				best = i
			}
		} else if four[i].v < four[best].v {
			best = i
		}
	}
	for i := 0; i < 4; i++ {
		if i != best && quizNear(four[i].v, four[best].v) {
			return nil
		}
	}
	wrong := make([]quizChoice, 0, 3)
	for i, r := range four {
		if i == best {
			continue
		}
		wrong = append(wrong, quizChoice{Label: r.d.Name, Icon: quizChampIcon(r.d)})
	}
	word := "le plus long"
	if !longest {
		word = "le plus court"
	}
	parts := make([]string, 0, 4)
	for _, r := range four {
		parts = append(parts, fmt.Sprintf("%s R %s", r.d.Name, fmtQuizCD(r.v)))
	}
	return &quizBuilt{
		key:     "ult|" + four[best].d.ID + "|" + strconv.FormatBool(longest),
		kind:    "CD",
		prompt:  "Parmi ces ultis, lequel a " + word + " CD au rang 1 (sans haste) ?",
		sub:     "Compare les R au rang 1, pas au rang max.",
		correct: quizChoice{Label: four[best].d.Name, Icon: quizChampIcon(four[best].d)},
		pool:    wrong,
		explain: strings.Join(parts, " · ") + ".",
	}
}

func genSpellName(pool []champDetail) *quizBuilt {
	spells := allPoolSpells(pool)
	if len(spells) < 8 {
		return nil
	}
	var pick quizSpell
	for try := 0; try < 16; try++ {
		s := spells[rand.IntN(len(spells))]
		if nameLeaksChamp(s.Champ, s.Name) {
			continue
		}
		pick = s
		break
	}
	if pick.Name == "" {
		return nil
	}
	wrong := make([]quizChoice, 0, 12)
	for _, s := range spells {
		if s.Name == pick.Name || (s.Champ.ID == pick.Champ.ID && s.Letter == pick.Letter) {
			continue
		}
		if s.Letter != pick.Letter && s.Champ.ID != pick.Champ.ID && rand.Float64() < 0.35 {
			continue
		}
		wrong = append(wrong, quizChoice{Label: s.Name, Icon: s.Icon})
	}
	return &quizBuilt{
		key:     "name|" + pick.Champ.ID + "|" + pick.Letter,
		kind:    "SORT",
		prompt:  fmt.Sprintf("Comment s'appelle le %s de %s ?", pick.Letter, pick.Champ.Name),
		sub:     "Nom Data Dragon (fr_FR).",
		icon:    quizChampIcon(pick.Champ),
		correct: quizChoice{Label: pick.Name, Icon: pick.Icon},
		pool:    wrong,
		explain: pick.label() + ".",
	}
}

func genSpellWho(pool []champDetail) *quizBuilt {
	spells := allPoolSpells(pool)
	if len(spells) < 8 {
		return nil
	}
	byName := map[string]int{}
	for _, s := range spells {
		byName[s.Name]++
	}
	cands := make([]quizSpell, 0, len(spells))
	for _, s := range spells {
		if byName[s.Name] != 1 || nameLeaksChamp(s.Champ, s.Name) {
			continue
		}
		cands = append(cands, s)
	}
	if len(cands) == 0 {
		return nil
	}
	pick := cands[rand.IntN(len(cands))]
	iconOnly := pick.Icon != "" && rand.Float64() < 0.5
	wrong := make([]quizChoice, 0, len(pool))
	for _, d := range pool {
		if d.ID == pick.Champ.ID {
			continue
		}
		wrong = append(wrong, quizChoice{Label: d.Name, Icon: quizChampIcon(d)})
	}
	b := &quizBuilt{
		key:     "who|" + pick.Champ.ID + "|" + pick.Letter,
		kind:    "SORT",
		correct: quizChoice{Label: pick.Champ.Name, Icon: quizChampIcon(pick.Champ)},
		pool:    wrong,
		explain: pick.label() + ".",
	}
	if iconOnly {
		b.prompt = "De quel champion est ce sort ?"
		b.sub = "Pas de lettre · icône Data Dragon."
		b.icon = pick.Icon
	} else {
		b.prompt = "« " + pick.Name + " » est le sort de :"
		b.sub = "Retrouve le champion (slot caché)."
		b.icon = pick.Icon
	}
	return b
}

func genDescWho(pool []champDetail) *quizBuilt {
	spells := allPoolSpells(pool)
	cands := make([]quizSpell, 0, len(spells))
	for _, s := range spells {
		if len([]rune(s.Desc)) < 50 {
			continue
		}
		cands = append(cands, s)
	}
	if len(cands) < 4 {
		return nil
	}
	pick := cands[rand.IntN(len(cands))]
	text := scrubQuizText(pick.Champ, pick.Desc)
	if strings.Contains(strings.ToLower(text), strings.ToLower(pick.Champ.Name)) {
		return nil
	}
	wrong := make([]quizChoice, 0, 8)
	for _, s := range cands {
		if s.Champ.ID == pick.Champ.ID && s.Letter == pick.Letter {
			continue
		}
		wrong = append(wrong, quizChoice{Label: s.label(), Icon: s.Icon})
	}
	return &quizBuilt{
		key:     "desc|" + pick.Champ.ID + "|" + pick.Letter,
		kind:    "SORT",
		prompt:  text,
		sub:     "Quel sort ? Description ddragon, nom du champion masqué.",
		correct: quizChoice{Label: pick.label(), Icon: pick.Icon},
		pool:    wrong,
		explain: pick.label() + " · CD " + fmtQuizCurve(pick.CDs) + ".",
	}
}

func genPassiveCD(pool []champDetail) *quizBuilt {
	type hit struct {
		d    champDetail
		meta passiveCDMeta
		desc string
	}
	cands := make([]hit, 0, 8)
	for _, d := range pool {
		desc := plainText(d.Passive.Description)
		meta := passiveCDFor(d.ID, desc)
		if meta.has() {
			cands = append(cands, hit{d, meta, desc})
		}
	}
	if len(cands) == 0 {
		return nil
	}
	h := cands[rand.IntN(len(cands))]
	level := 1
	ask := "Le CD du passif de " + h.d.Name
	if h.d.Passive.Name != "" {
		ask += " (« " + h.d.Passive.Name + " »)"
	}
	if len(h.meta.Values) >= 10 {
		if rand.Float64() < 0.5 {
			level = 18
		}
		ask += fmt.Sprintf(" au niveau %d", level)
	} else if len(h.meta.Values) > 1 && distinctCDCount(h.meta.Values) >= 2 {
		if rand.Float64() < 0.5 {
			level = 18
		}
		ask += fmt.Sprintf(" au niveau %d", level)
	}
	ask += " (sans haste) ?"
	v := pickPassiveCD(h.meta.Values, level)
	if v <= 0 {
		return nil
	}
	related := []float64{4, 8, 12, 20, 30, 90, 120, 180, 240, 300}
	related = append(related, h.meta.Values...)
	poolCh := choicesFromCDs(v, related)
	if len(poolCh) < 3 {
		return nil
	}
	haste := "Non réduit par l'AH."
	if h.meta.Haste {
		haste = "Réduit par l'Ability Haste."
	}
	return &quizBuilt{
		key:     "pcd|" + h.d.ID + "|" + strconv.Itoa(level),
		kind:    "PASSIF",
		prompt:  ask,
		sub:     "Les passifs n'ont pas de cooldownBurn ddragon — valeur wiki / texte.",
		icon:    quizPassiveIcon(h.d),
		correct: quizChoice{Label: fmtQuizCD(v)},
		pool:    poolCh,
		explain: fmt.Sprintf("%s P (%s) : %s. Niveau %d = %s. %s", h.d.Name, h.d.Passive.Name, fmtQuizCurve(h.meta.Values), level, fmtQuizCD(v), haste),
	}
}

func genPassiveWho(pool []champDetail) *quizBuilt {
	cands := make([]champDetail, 0, len(pool))
	byName := map[string]int{}
	for _, d := range pool {
		n := strings.TrimSpace(d.Passive.Name)
		if n == "" || nameLeaksChamp(d, n) {
			continue
		}
		byName[n]++
		cands = append(cands, d)
	}
	uniq := make([]champDetail, 0, len(cands))
	for _, d := range cands {
		if byName[d.Passive.Name] == 1 {
			uniq = append(uniq, d)
		}
	}
	if len(uniq) == 0 {
		return nil
	}
	d := uniq[rand.IntN(len(uniq))]
	wrong := make([]quizChoice, 0, len(pool))
	for _, o := range pool {
		if o.ID == d.ID {
			continue
		}
		wrong = append(wrong, quizChoice{Label: o.Name, Icon: quizChampIcon(o)})
	}
	return &quizBuilt{
		key:     "pwho|" + d.ID,
		kind:    "PASSIF",
		prompt:  "Quel champion a le passif « " + d.Passive.Name + " » ?",
		sub:     "Nom exact Data Dragon.",
		icon:    quizPassiveIcon(d),
		correct: quizChoice{Label: d.Name, Icon: quizChampIcon(d)},
		pool:    wrong,
		explain: d.Name + " — " + d.Passive.Name + ".",
	}
}

func genCostRank(pool []champDetail) *quizBuilt {
	for try := 0; try < 12; try++ {
		d := pickPoolChamp(pool)
		cands := make([]quizSpell, 0, 4)
		for _, s := range listQuizSpells(d) {
			v, ok := firstPositive(s.Cost)
			if ok && v >= 20 {
				cands = append(cands, s)
			}
		}
		if len(cands) == 0 {
			continue
		}
		s := cands[rand.IntN(len(cands))]
		rank := 1
		if distinctCDCount(s.Cost) >= 2 && len(s.Cost) > 1 && rand.Float64() < 0.4 {
			rank = len(s.Cost)
		}
		if rank > len(s.Cost) {
			rank = 1
		}
		v := s.Cost[rank-1]
		if v <= 0 {
			continue
		}
		related := append([]float64{40, 50, 60, 70, 80, 90, 100, 110, 120}, s.Cost...)
		dstr := cdDistractors(v, related)
		if len(dstr) < 3 {
			continue
		}
		poolCh := make([]quizChoice, 0, len(dstr))
		for _, x := range dstr {
			poolCh = append(poolCh, quizChoice{Label: compactSec(x)})
		}
		return &quizBuilt{
			key:     "cost|" + d.ID + "|" + s.Letter + "|" + strconv.Itoa(rank),
			kind:    "SORT",
			prompt:  fmt.Sprintf("Coût du %s de %s au %s ?", s.Letter, d.Name, quizRankLabel(rank, len(s.Cost))),
			sub:     s.Name + " · ressource Data Dragon (mana / énergie / PV…).",
			icon:    s.Icon,
			correct: quizChoice{Label: compactSec(v)},
			pool:    poolCh,
			explain: fmt.Sprintf("%s %s (%s) : %s. %s = %s.", d.Name, s.Letter, s.Name, strings.Join(compactAll(s.Cost), " → "), quizRankLabel(rank, len(s.Cost)), compactSec(v)),
		}
	}
	return nil
}

func firstPositive(vals []float64) (float64, bool) {
	for _, v := range vals {
		if v > 0 {
			return v, true
		}
	}
	return 0, false
}

func compactAll(vals []float64) []string {
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = compactSec(v)
	}
	return out
}

func wrapQuiz(b *quizBuilt, asked, score, streak int) (quizQuestion, quizPending, bool) {
	if b == nil {
		return quizQuestion{}, quizPending{}, false
	}
	choices := finishQuizChoices(b.correct, b.pool, 4)
	if len(choices) != 4 {
		return quizQuestion{}, quizPending{}, false
	}
	correctID := ""
	for _, c := range choices {
		if c.Label == b.correct.Label {
			correctID = c.ID
			break
		}
	}
	if correctID == "" {
		return quizQuestion{}, quizPending{}, false
	}
	id := fmt.Sprintf("%x", rand.Uint64())
	q := quizQuestion{
		ID: id, Kind: b.kind, Prompt: b.prompt, Sub: b.sub, Icon: b.icon,
		Choices: choices, Round: asked + 1, Total: quizRoundTotal,
		Score: score, Streak: streak, Patch: getVersion(),
	}
	p := quizPending{id: id, correct: correctID, answer: b.correct.Label, explain: b.explain}
	return q, p, true
}

func buildQuizFromPool(pool []champDetail, seen map[string]bool) (quizQuestion, quizPending, bool) {
	if len(pool) < 2 {
		return quizQuestion{}, quizPending{}, false
	}
	gens := []struct {
		w  int
		fn func([]champDetail) *quizBuilt
	}{
		{5, genCDRank},
		{4, genCDCurveWho},
		{4, genWhichSpellCD},
		{3, genUltCompare},
		{3, genPassiveCD},
		{2, genSpellName},
		{2, genSpellWho},
		{2, genDescWho},
		{2, genPassiveWho},
		{1, genCostRank},
	}
	total := 0
	for _, g := range gens {
		total += g.w
	}
	for try := 0; try < 56; try++ {
		x := rand.IntN(total)
		var fn func([]champDetail) *quizBuilt
		for _, g := range gens {
			if x < g.w {
				fn = g.fn
				break
			}
			x -= g.w
		}
		b := fn(pool)
		if b == nil || (seen != nil && seen[b.key]) {
			continue
		}
		q, p, ok := wrapQuiz(b, 0, 0, 0)
		if !ok {
			continue
		}
		if seen != nil {
			seen[b.key] = true
		}
		return q, p, true
	}
	return quizQuestion{}, quizPending{}, false
}

func cachedQuizChamps() []champDetail {
	dragon.mu.Lock()
	defer dragon.mu.Unlock()
	out := make([]champDetail, 0, len(dragon.details))
	for _, d := range dragon.details {
		if len(d.Spells) >= 4 {
			out = append(out, d)
		}
	}
	return out
}

func ensureQuizPool(n int) []champDetail {
	have := cachedQuizChamps()
	if len(have) >= n {
		return have
	}
	if ensureIndex() != nil {
		return have
	}
	dragon.mu.Lock()
	items := append([]champIndexItem(nil), dragon.index...)
	cached := map[string]bool{}
	for id := range dragon.details {
		cached[id] = true
	}
	dragon.mu.Unlock()
	rand.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
	need := n - len(have)
	if need < 8 {
		need = 8
	}
	var slugs []string
	for _, it := range items {
		if cached[it.ID] {
			continue
		}
		slugs = append(slugs, it.ID)
		if len(slugs) >= need {
			break
		}
	}
	var wg sync.WaitGroup
	for _, id := range slugs {
		wg.Add(1)
		go func(slug string) {
			defer wg.Done()
			_, _ = getDetail(slug)
		}(id)
	}
	wg.Wait()
	prefetchPassiveCDs(slugs)
	return cachedQuizChamps()
}

func nextQuizQuestionLocked() (quizQuestion, error) {
	if quizSess.done {
		return quizQuestion{Done: true, Round: quizSess.asked, Total: quizRoundTotal, Score: quizSess.score, Streak: quizSess.streak, Patch: getVersion()}, nil
	}
	if quizSess.q != nil && quizSess.pending != nil {
		q := *quizSess.q
		q.Round, q.Total, q.Score, q.Streak = quizSess.asked+1, quizRoundTotal, quizSess.score, quizSess.streak
		return q, nil
	}
	if quizSess.asked >= quizRoundTotal {
		quizSess.done = true
		return quizQuestion{Done: true, Round: quizSess.asked, Total: quizRoundTotal, Score: quizSess.score, Streak: quizSess.streak, Patch: getVersion()}, nil
	}
	pool := ensureQuizPool(16)
	if len(pool) < 4 {
		return quizQuestion{}, fmt.Errorf("Data Dragon injoignable")
	}
	if quizSess.seen == nil {
		quizSess.seen = map[string]bool{}
	}
	q, p, ok := buildQuizFromPool(pool, quizSess.seen)
	if !ok {
		pool = ensureQuizPool(40)
		q, p, ok = buildQuizFromPool(pool, quizSess.seen)
	}
	if !ok {
		return quizQuestion{}, fmt.Errorf("impossible de composer une question")
	}
	q.Round, q.Total, q.Score, q.Streak = quizSess.asked+1, quizRoundTotal, quizSess.score, quizSess.streak
	quizSess.q, quizSess.pending = &q, &p
	return q, nil
}

func resetQuizLocked() {
	quizSess.asked, quizSess.score, quizSess.streak = 0, 0, 0
	quizSess.pending, quizSess.q = nil, nil
	quizSess.seen = map[string]bool{}
	quizSess.done = false
}

func apiQuiz(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		newRound := r.URL.Query().Get("new") == "1"
		quizSess.mu.Lock()
		if newRound || (quizSess.q == nil && quizSess.asked == 0 && !quizSess.done) {
			resetQuizLocked()
		}
		q, err := nextQuizQuestionLocked()
		quizSess.mu.Unlock()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		go func() { _ = ensureQuizPool(40) }()
		writeJSON(w, q)
	case http.MethodPost:
		var body struct {
			ID   string `json:"id"`
			Pick string `json:"pick"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&body); err != nil {
			http.Error(w, "json", http.StatusBadRequest)
			return
		}
		quizSess.mu.Lock()
		defer quizSess.mu.Unlock()
		p := quizSess.pending
		if p == nil || (body.ID != "" && body.ID != p.id) {
			http.Error(w, "question expirée", http.StatusConflict)
			return
		}
		ok := body.Pick == p.correct
		if ok {
			quizSess.score++
			quizSess.streak++
		} else {
			quizSess.streak = 0
		}
		quizSess.asked++
		quizSess.pending, quizSess.q = nil, nil
		done := quizSess.asked >= quizRoundTotal
		if done {
			quizSess.done = true
		}
		writeJSON(w, quizGrade{
			Correct: ok, Pick: body.Pick, AnswerID: p.correct, Answer: p.answer,
			Explain: p.explain, Score: quizSess.score, Streak: quizSess.streak,
			Round: quizSess.asked, Total: quizRoundTotal, Done: done,
		})
	default:
		http.Error(w, "GET/POST", http.StatusMethodNotAllowed)
	}
}
