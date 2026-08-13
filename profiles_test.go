package main

import "testing"

func TestVisibleName(t *testing.T) {
	cases := []struct {
		riot, champ, want string
		hidden            bool
	}{
		{"Cait#EUW", "Caitlyn", "Cait#EUW", false},
		{"Caitlyn", "Caitlyn", "Caitlyn", true},
		{"Caitlyn#EUW", "Caitlyn", "Caitlyn", true},
		{"", "Zed", "Zed", true},
		{"game_character_displayname_Ahri", "Ahri", "Ahri", true},
		{"Hide on bush#KR1", "Ahri", "Hide on bush#KR1", false},
		{"OldName", "Garen", "OldName", false},
	}
	for _, c := range cases {
		got, hid := visibleName(c.riot, c.champ)
		if got != c.want || hid != c.hidden {
			t.Errorf("visibleName(%q,%q)=(%q,%v), attendu (%q,%v)", c.riot, c.champ, got, hid, c.want, c.hidden)
		}
	}
}

func TestRankShort(t *testing.T) {
	cases := []struct {
		tier, div string
		lp        int
		want      string
	}{
		{"DIAMOND", "II", 52, "D2"},
		{"EMERALD", "I", 0, "E1"},
		{"GOLD", "IV", 10, "G4"},
		{"MASTER", "NA", 142, "Master 142"},
		{"GRANDMASTER", "I", 80, "GM 80"},
		{"CHALLENGER", "", 0, "Chall"},
		{"UNRANKED", "I", 0, ""},
		{"", "", 0, ""},
	}
	for _, c := range cases {
		if got := rankShort(c.tier, c.div, c.lp); got != c.want {
			t.Errorf("rankShort(%q,%q,%d)=%q, attendu %q", c.tier, c.div, c.lp, got, c.want)
		}
	}
}

func TestPickSolo(t *testing.T) {
	s := &lcuRankedStats{
		QueueMap: map[string]lcuRankedEntry{
			"RANKED_FLEX_SR":  {QueueType: "RANKED_FLEX_SR", Tier: "GOLD", Division: "I"},
			"RANKED_SOLO_5x5": {QueueType: "RANKED_SOLO_5x5", Tier: "DIAMOND", Division: "II", LeaguePoints: 40},
		},
	}
	e := pickSolo(s)
	if e.Tier != "DIAMOND" || e.Division != "II" {
		t.Errorf("solo attendu D2, obtenu %+v", e)
	}
	flexOnly := &lcuRankedStats{Queues: []lcuRankedEntry{{QueueType: "RANKED_FLEX_SR", Tier: "PLATINUM", Division: "III"}}}
	if got := pickSolo(flexOnly); got.Tier != "PLATINUM" {
		t.Errorf("flex en repli attendu, obtenu %+v", got)
	}
}

func TestDeepLoLURL(t *testing.T) {
	u := deepLoLURL("EUW", "Hide on bush", "KR1")
	want := "https://www.deeplol.gg/summoner/euw/Hide%20on%20bush-KR1"
	if u != want {
		t.Errorf("url=%q, attendu %q", u, want)
	}
	if !isDeepLoLURL(u) {
		t.Errorf("isDeepLoLURL devrait accepter %q", u)
	}
	if isDeepLoLURL("https://evil.example/summoner/euw/x-y") || isDeepLoLURL("http://www.deeplol.gg/summoner/euw/a-b") {
		t.Error("isDeepLoLURL trop permissif")
	}
	if regionFromTag("EUW") != "euw" || regionFromTag("NA1") != "na" || regionFromTag("EUNE") != "eune" {
		t.Errorf("regionFromTag: EUW=%s NA1=%s EUNE=%s", regionFromTag("EUW"), regionFromTag("NA1"), regionFromTag("EUNE"))
	}
}

func TestApplyIdentityStreamer(t *testing.T) {
	storeIdent(&playerIdent{RiotID: "Secret#EUW", Champion: 51, Tier: "DIAMOND", Division: "II", Rank: "D2", DeepLoL: "https://www.deeplol.gg/summoner/euw/Secret-EUW", Ready: true})
	lp := LivePlayer{RiotID: "Caitlyn", Champion: "Caitlyn", Key: 51}
	applyIdentity(&lp)
	if !lp.Hidden || lp.Name != "Caitlyn" || lp.DeepLoL != "" {
		t.Errorf("streamer: nom/lien doivent rester cachés: %+v", lp)
	}
	if lp.Rank != "D2" || lp.Tier != "diamond" {
		t.Errorf("streamer: le rang peut rester visible: %+v", lp)
	}
	lp2 := LivePlayer{RiotID: "Cait#EUW", Champion: "Caitlyn", Key: 51}
	applyIdentity(&lp2)
	if lp2.Hidden || lp2.Name == "Caitlyn" || lp2.DeepLoL == "" || lp2.Rank != "D2" {
		t.Errorf("pseudo visible attendu: %+v", lp2)
	}
}
