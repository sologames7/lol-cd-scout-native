package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// Métadonnées d'item (ddragon fr_FR) : or total + nom pour les annonces,
// et le flag Complete pour détecter un légendaire / item fini.

type itemMeta struct {
	Name     string
	Gold     int
	Complete bool
}

var itemMetaCache = struct {
	mu      sync.Mutex
	version string
	data    map[int]itemMeta
}{}

func resetItemMetaCache() {
	itemMetaCache.mu.Lock()
	itemMetaCache.version, itemMetaCache.data = "", nil
	itemMetaCache.mu.Unlock()
}

func itemMetaTable() map[int]itemMeta {
	version := getVersion()
	itemMetaCache.mu.Lock()
	if itemMetaCache.version == version && itemMetaCache.data != nil {
		d := itemMetaCache.data
		itemMetaCache.mu.Unlock()
		return d
	}
	itemMetaCache.mu.Unlock()

	var payload struct {
		Data map[string]struct {
			Name string              `json:"name"`
			Gold struct{ Total int } `json:"gold"`
			From []string            `json:"from"`
			Into []string            `json:"into"`
			Tags []string            `json:"tags"`
		} `json:"data"`
	}
	url := fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/fr_FR/item.json", version)
	if err := jsonGET(url, &payload); err != nil {
		itemMetaCache.mu.Lock()
		defer itemMetaCache.mu.Unlock()
		if itemMetaCache.data != nil {
			return itemMetaCache.data
		}
		return map[int]itemMeta{}
	}
	table := map[int]itemMeta{}
	for idStr, it := range payload.Data {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		table[id] = itemMeta{
			Name:     it.Name,
			Gold:     it.Gold.Total,
			Complete: itemIsComplete(it.Gold.Total, it.From, it.Into, it.Tags),
		}
	}
	itemMetaCache.mu.Lock()
	itemMetaCache.version, itemMetaCache.data = version, table
	itemMetaCache.mu.Unlock()
	return table
}

// Légendaire / item fini : construit (from), or conséquent, pas un composant
// ni des bottes / trinket / conso / pet jungle. Seuil 1600 pour coller Mejai.
func itemIsComplete(gold int, from, into, tags []string) bool {
	if gold < 1600 || len(from) == 0 {
		return false
	}
	for _, t := range tags {
		switch strings.ToLower(t) {
		case "trinket", "consumable", "jungle", "lane", "boots":
			return false
		}
	}
	if len(into) == 0 {
		return true
	}
	// Voie d'upgrade (Ornn, quête support) : on annonce déjà le légendaire.
	return gold >= 2000
}

func itemGoldSum(ids []int, meta map[int]itemMeta) int {
	n := 0
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if m, ok := meta[id]; ok {
			n += m.Gold
		}
	}
	return n
}

func shortItemName(name string) string {
	name = strings.TrimSpace(name)
	if i := strings.Index(name, ","); i > 3 && i < 28 {
		return strings.TrimSpace(name[:i])
	}
	return name
}
