//go:build ignore

// Génère synergies.go (partenaires LoLTheory, 12 plus joués parmi les bons WR)
// et roles.go (rôles Meraki, playRate > 0.35).
//
//	go run gensynergies.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type ddChamp struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type partner struct {
	Name string
	PR   float64
}

var slugOverride = map[string]string{
	"MonkeyKing": "wukong",
	"Renata":     "renataglasc",
	"Nunu":       "nunu",
}

var httpC = &http.Client{Timeout: 18 * time.Second}

func get(url string) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; CDScout/1.0)")
	req.Header.Set("Accept", "text/html,application/json")
	res, err := httpC.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d", res.StatusCode)
	}
	return b, nil
}

func main() {
	var payload struct {
		Data map[string]ddChamp `json:"data"`
	}
	raw, err := get("https://ddragon.leagueoflegends.com/cdn/16.16.1/data/en_US/champion.json")
	if err != nil {
		raw, err = get("https://ddragon.leagueoflegends.com/cdn/16.15.1/data/en_US/champion.json")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ddragon: %v\n", err)
		os.Exit(1)
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		fmt.Fprintf(os.Stderr, "ddragon json: %v\n", err)
		os.Exit(1)
	}
	champs := make([]ddChamp, 0, len(payload.Data))
	names := make([]string, 0, len(payload.Data))
	for _, c := range payload.Data {
		champs = append(champs, c)
		names = append(names, c.Name)
	}
	sort.Slice(champs, func(i, j int) bool { return champs[i].ID < champs[j].ID })
	sort.Slice(names, func(i, j int) bool { return len(names[i]) > len(names[j]) })

	type result struct {
		ID   string
		List []string
		Err  error
	}
	out := make(chan result, len(champs))
	sem := make(chan struct{}, 6)
	var wg sync.WaitGroup
	for _, c := range champs {
		wg.Add(1)
		go func(c ddChamp) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			list, err := fetchSynergies(c, names)
			out <- result{c.ID, list, err}
		}(c)
	}
	go func() { wg.Wait(); close(out) }()

	syn := map[string][]string{}
	ok, fail := 0, 0
	for r := range out {
		if r.Err != nil || len(r.List) == 0 {
			fail++
			if r.Err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", r.ID, r.Err)
			}
			continue
		}
		syn[r.ID] = r.List
		ok++
		fmt.Printf("%s: %s\n", r.ID, strings.Join(r.List, ", "))
	}
	fmt.Printf("synergies: %d ok / %d manquants\n", ok, fail)
	if err := writeSynergies(syn); err != nil {
		fmt.Fprintf(os.Stderr, "write synergies: %v\n", err)
		os.Exit(1)
	}
	if err := writeRoles(payload.Data); err != nil {
		fmt.Fprintf(os.Stderr, "write roles: %v\n", err)
		os.Exit(1)
	}
}

func fetchSynergies(c ddChamp, names []string) ([]string, error) {
	slugs := []string{strings.ToLower(c.ID)}
	if s, ok := slugOverride[c.ID]; ok {
		slugs = append([]string{s}, slugs...)
	}
	slugName := strings.ToLower(strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return -1
	}, c.Name))
	if slugName != "" && slugName != strings.ToLower(c.ID) {
		slugs = append(slugs, slugName)
	}
	var last error
	for _, slug := range slugs {
		b, err := get("https://loltheory.gg/lol/synergies/" + slug)
		if err != nil {
			last = err
			continue
		}
		html := string(b)
		if strings.Contains(html, "Best Picks with None") || strings.Contains(html, "with None") {
			last = fmt.Errorf("slug vide")
			continue
		}
		list := parseBest(html, names, c.Name)
		if len(list) > 0 {
			return list, nil
		}
		last = fmt.Errorf("aucun partenaire")
	}
	return nil, last
}

var prRe = regexp.MustCompile(`(\d+(?:\.\d+)?)%\s*PR`)

func parseBest(html string, names []string, self string) []string {
	low := html
	start := strings.Index(low, "Best Picks with")
	end := strings.Index(low, "Worst Picks with")
	if start < 0 {
		return nil
	}
	if end < start {
		end = len(low)
	}
	section := low[start:end]
	seen := map[string]float64{}
	selfN := normName(self)
	for _, name := range names {
		if normName(name) == selfN {
			continue
		}
		for _, idx := range findName(section, name) {
			rest := section[idx+len(name):]
			m := prRe.FindStringSubmatch(rest)
			if m == nil {
				continue
			}
			pr, _ := strconv.ParseFloat(m[1], 64)
			if pr > seen[name] {
				seen[name] = pr
			}
			break
		}
	}
	type hit struct {
		Name string
		PR   float64
	}
	hits := make([]hit, 0, len(seen))
	for n, pr := range seen {
		if pr < 0.8 {
			continue
		}
		hits = append(hits, hit{n, pr})
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].PR == hits[j].PR {
			return hits[i].Name < hits[j].Name
		}
		return hits[i].PR > hits[j].PR
	})
	out := []string{}
	for _, h := range hits {
		out = append(out, h.Name)
		if len(out) == 12 {
			break
		}
	}
	return out
}

func findName(s, name string) []int {
	var idx []int
	from := 0
	for {
		i := strings.Index(s[from:], name)
		if i < 0 {
			return idx
		}
		i += from
		okPrev := i == 0 || !isNameChar(rune(s[i-1]))
		end := i + len(name)
		okNext := end >= len(s) || !isNameChar(rune(s[end]))
		if okPrev && okNext {
			idx = append(idx, i)
		}
		from = i + len(name)
	}
}

func isNameChar(r rune) bool {
	return unicode.IsLetter(r) || r == '\'' || r == '’'
}

func normName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func writeSynergies(syn map[string][]string) error {
	keys := make([]string, 0, len(syn))
	for k := range syn {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf bytes.Buffer
	buf.WriteString("package main\n\n")
	buf.WriteString("// Auto-generated from LoLTheory Platinum synergies (~patch 16.16).\n")
	buf.WriteString("// 12 partenaires les plus joués parmi les paires à bon WR (min 0,8 % PR).\n")
	buf.WriteString("// Une override curated avec Synergies non vides remplace cette liste.\n\n")
	buf.WriteString("var synergies = map[string][]string{\n")
	for _, k := range keys {
		buf.WriteString("\t\"" + k + "\": {")
		for i, n := range syn[k] {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(strconv.Quote(n))
		}
		buf.WriteString("},\n")
	}
	buf.WriteString("}\n")
	return os.WriteFile("synergies.go", buf.Bytes(), 0644)
}

func writeRoles(champs map[string]ddChamp) error {
	raw, err := get("https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/championrates.json")
	if err != nil {
		return err
	}
	var rates struct {
		Data map[string]map[string]struct {
			PlayRate float64 `json:"playRate"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &rates); err != nil {
		return err
	}
	label := map[string]string{
		"TOP": "TOP", "JUNGLE": "JGL", "MIDDLE": "MID", "BOTTOM": "BOT", "UTILITY": "SUPP",
	}
	order := []string{"TOP", "JUNGLE", "MIDDLE", "BOTTOM", "UTILITY"}
	type row struct {
		Key   int
		Roles []string
	}
	var rows []row
	for keyStr, lanes := range rates.Data {
		key, _ := strconv.Atoi(keyStr)
		if key == 0 || key >= 60000 {
			continue
		}
		var rs []string
		for _, lane := range order {
			if lanes[lane].PlayRate >= 0.35 {
				rs = append(rs, label[lane])
			}
		}
		if len(rs) == 0 {
			bestLane, best := "", -1.0
			for _, lane := range order {
				if lanes[lane].PlayRate > best {
					best, bestLane = lanes[lane].PlayRate, lane
				}
			}
			if best > 0 {
				rs = []string{label[bestLane]}
			}
		}
		if len(rs) > 0 {
			rows = append(rows, row{key, rs})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Key < rows[j].Key })
	var buf bytes.Buffer
	buf.WriteString("package main\n\n")
	buf.WriteString("// Auto-generated from Meraki championrates (playRate ≥ 0,35).\n")
	buf.WriteString("// Sert à filtrer les propositions de pick par rôle en champ select.\n\n")
	buf.WriteString("var champRoles = map[int][]string{\n")
	for _, r := range rows {
		buf.WriteString(fmt.Sprintf("\t%d: {", r.Key))
		for i, role := range r.Roles {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(strconv.Quote(role))
		}
		buf.WriteString("},\n")
	}
	buf.WriteString("}\n")
	_ = champs
	return os.WriteFile("roles.go", buf.Bytes(), 0644)
}
