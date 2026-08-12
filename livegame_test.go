package main

import "testing"

// Vérifie que la haste est bien extraite des descriptions Data Dragon (réseau requis).
func TestItemHasteTable(t *testing.T) {
	table := itemHasteTable()
	if len(table) < 20 {
		t.Fatalf("table de haste trop petite: %d items", len(table))
	}
	// 3158 = Bottes de lucidité ioniennes : ability haste + summoner spell haste.
	boots, ok := table[3158]
	if !ok || boots.Ability <= 0 || boots.Summoner <= 0 {
		t.Errorf("Bottes de lucidité mal parsées: %+v (ok=%v)", boots, ok)
	}
	t.Logf("items avec haste: %d · Lucidité: %+v", len(table), boots)
}

func TestRanks(t *testing.T) {
	cases := []struct{ level, ult, basic int }{
		{1, 0, 1}, {3, 0, 2}, {6, 1, 3}, {9, 1, 5}, {11, 2, 5}, {16, 3, 5}, {18, 3, 5},
	}
	for _, c := range cases {
		if got := ultRank(c.level); got != c.ult {
			t.Errorf("ultRank(%d)=%d, attendu %d", c.level, got, c.ult)
		}
		if got := maxBasicRank(c.level); got != c.basic {
			t.Errorf("maxBasicRank(%d)=%d, attendu %d", c.level, got, c.basic)
		}
	}
}
