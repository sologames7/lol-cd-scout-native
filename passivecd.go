package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Les passifs ddragon n'ont pas de cooldownBurn : le CD est soit dans le texte,
// soit (plus fiable) dans les données wiki Meraki, soit dans une fiche curated.

type passiveCDMeta struct {
	Values []float64 // secondes, 1 / 3–6 paliers / 18 niveaux
	Haste  bool      // true si l'Ability Haste réduit le CD
}

func (m passiveCDMeta) has() bool {
	for _, v := range m.Values {
		if v > 0 {
			return true
		}
	}
	return false
}

var passiveCDs = struct {
	mu        sync.Mutex
	byID      map[string]passiveCDMeta
	fails     int
	skipUntil time.Time
}{byID: map[string]passiveCDMeta{}}

func resetPassiveCDCache() {
	passiveCDs.mu.Lock()
	passiveCDs.byID = map[string]passiveCDMeta{}
	passiveCDs.fails = 0
	passiveCDs.skipUntil = time.Time{}
	passiveCDs.mu.Unlock()
}

type merakiChamp struct {
	Abilities struct {
		P []struct {
			Cooldown *struct {
				Modifiers []struct {
					Values []float64 `json:"values"`
				} `json:"modifiers"`
				AffectedByCdr bool `json:"affectedByCdr"`
			} `json:"cooldown"`
		} `json:"P"`
	} `json:"abilities"`
}

func getPassiveCD(slug string) (passiveCDMeta, bool) {
	if slug == "" {
		return passiveCDMeta{}, true
	}
	passiveCDs.mu.Lock()
	if m, ok := passiveCDs.byID[slug]; ok {
		passiveCDs.mu.Unlock()
		return m, true
	}
	if time.Now().Before(passiveCDs.skipUntil) {
		passiveCDs.mu.Unlock()
		return passiveCDMeta{}, false
	}
	passiveCDs.mu.Unlock()

	m, err := fetchMerakiPassiveCD(slug)
	passiveCDs.mu.Lock()
	defer passiveCDs.mu.Unlock()
	if err != nil {
		if strings.Contains(err.Error(), "HTTP 404") {
			passiveCDs.byID[slug] = passiveCDMeta{}
			return passiveCDMeta{}, true
		}
		passiveCDs.fails++
		if passiveCDs.fails >= 2 {
			passiveCDs.skipUntil = time.Now().Add(10 * time.Minute)
		}
		return passiveCDMeta{}, false
	}
	passiveCDs.fails = 0
	passiveCDs.byID[slug] = m
	return m, true
}

func fetchMerakiPassiveCD(slug string) (passiveCDMeta, error) {
	url := "https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/champions/" + slug + ".json"
	var payload merakiChamp
	if err := jsonGET(url, &payload); err != nil {
		return passiveCDMeta{}, err
	}
	out := passiveCDMeta{}
	if len(payload.Abilities.P) == 0 || payload.Abilities.P[0].Cooldown == nil {
		return out, nil
	}
	cd := payload.Abilities.P[0].Cooldown
	out.Haste = cd.AffectedByCdr
	if len(cd.Modifiers) > 0 {
		out.Values = cleanCDValues(cd.Modifiers[0].Values)
	}
	return out, nil
}

func prefetchPassiveCDs(slugs []string) {
	seen := map[string]bool{}
	var wg sync.WaitGroup
	for _, slug := range slugs {
		if slug == "" || seen[slug] {
			continue
		}
		seen[slug] = true
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			getPassiveCD(id)
		}(slug)
	}
	wg.Wait()
}

func cleanCDValues(vals []float64) []float64 {
	if len(vals) == 0 {
		return nil
	}
	out := make([]float64, 0, len(vals))
	max := 0.0
	for _, v := range vals {
		if v < 0 {
			v = 0
		}
		out = append(out, v)
		if v > max {
			max = v
		}
	}
	if max <= 0 {
		return nil
	}
	return out
}

var (
	passiveCDFrRe = regexp.MustCompile(`(?i)(?:d[ée]lai de )?r[ée]cup[ée]ration(?: de)?(?:\s*:)?\s*(\d+(?:[.,]\d+)?)\s*(minutes?|mins?|min|secondes?|secs?|s)?`)
	passiveCDEnRe = regexp.MustCompile(`(?i)(?:cooldown(?: of)?(?:\s*:)?\s*(\d+(?:[.,]\d+)?)\s*(minutes?|mins?|min|seconds?|secs?|s)?|(\d+(?:[.,]\d+)?)\s*(minutes?|minute|seconds?|second)\s+cooldown)`)
	cdStringRe    = regexp.MustCompile(`(?i)(\d+(?:[.,]\d+)?)\s*(minutes?|mins?|min|secondes?|secs?|s)?`)
)

func parseCDNumber(raw, unit string) float64 {
	raw = strings.ReplaceAll(raw, ",", ".")
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || v <= 0 {
		return 0
	}
	u := strings.ToLower(strings.TrimSpace(unit))
	if strings.HasPrefix(u, "min") {
		v *= 60
	}
	return v
}

func parsePassiveCD(desc string) []float64 {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return nil
	}
	if m := passiveCDFrRe.FindStringSubmatch(desc); len(m) > 1 {
		if v := parseCDNumber(m[1], m[2]); v > 0 {
			return []float64{v}
		}
	}
	if m := passiveCDEnRe.FindStringSubmatch(desc); len(m) > 1 {
		num, unit := m[1], m[2]
		if num == "" {
			num, unit = m[3], m[4]
		}
		if v := parseCDNumber(num, unit); v > 0 {
			return []float64{v}
		}
	}
	return nil
}

func parseCDString(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "—" || strings.EqualFold(s, "auto") {
		return 0
	}
	m := cdStringRe.FindStringSubmatch(s)
	if len(m) < 2 {
		return 0
	}
	return parseCDNumber(m[1], m[2])
}

func compactSec(v float64) string {
	if v >= 100 || v == float64(int(v)) {
		return strconv.Itoa(int(v + 0.5))
	}
	s := strconv.FormatFloat(v, 'f', 1, 64)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	return s
}

func formatPassiveCD(vals []float64) string {
	meta := passiveCDMeta{Values: vals}
	if !meta.has() {
		return "—"
	}
	if len(vals) >= 10 {
		return compactSec(vals[0]) + " → " + compactSec(vals[len(vals)-1]) + "s"
	}
	uniq := make([]float64, 0, len(vals))
	for _, v := range vals {
		if v <= 0 {
			continue
		}
		if len(uniq) == 0 || absF(uniq[len(uniq)-1]-v) > 0.05 {
			uniq = append(uniq, v)
		}
	}
	if len(uniq) == 0 {
		return "—"
	}
	if len(uniq) == 1 {
		return compactSec(uniq[0]) + "s"
	}
	parts := make([]string, len(uniq))
	for i, v := range uniq {
		parts[i] = compactSec(v)
	}
	return strings.Join(parts, " → ") + "s"
}

func absF(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func pickPassiveCD(vals []float64, level int) float64 {
	if len(vals) == 0 {
		return 0
	}
	if level < 1 {
		level = 1
	}
	if level > 18 {
		level = 18
	}
	if len(vals) >= 10 {
		i := level - 1
		if i >= len(vals) {
			i = len(vals) - 1
		}
		return vals[i]
	}
	if len(vals) == 1 {
		return vals[0]
	}
	bucket := (level - 1) * len(vals) / 18
	if bucket >= len(vals) {
		bucket = len(vals) - 1
	}
	return vals[bucket]
}

func curatedPassiveCD(id string) float64 {
	o, ok := overrides[id]
	if !ok {
		return 0
	}
	for _, s := range o.Important {
		if s.Spell == "P" {
			return parseCDString(s.CD)
		}
	}
	return 0
}

func passiveCDFor(id, desc string) passiveCDMeta {
	if m, ok := getPassiveCD(id); ok && m.has() {
		return m
	}
	if vals := parsePassiveCD(desc); len(vals) > 0 {
		return passiveCDMeta{Values: vals}
	}
	if v := curatedPassiveCD(id); v > 0 {
		return passiveCDMeta{Values: []float64{v}}
	}
	return passiveCDMeta{}
}

func livePassiveSpell(id, name, icon, desc string, level, ah int) LiveSpell {
	if name == "" {
		name = "Passif"
	}
	sp := LiveSpell{Spell: "P", Name: name, Icon: icon, Desc: desc, Rank: "P"}
	meta := passiveCDFor(id, desc)
	if !meta.has() {
		return sp
	}
	cd := pickPassiveCD(meta.Values, level)
	if meta.Haste {
		cd *= hasteFactor(ah)
	}
	sp.CD = cd
	if len(meta.Values) >= 10 {
		sp.Rank = strconv.Itoa(level)
	} else if len(meta.Values) > 1 {
		sp.Rank = fmt.Sprintf("nv%d", level)
		sp.Est = true
	}
	return sp
}
