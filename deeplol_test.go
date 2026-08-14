package main

import "testing"

func TestPlatformID(t *testing.T) {
	cases := map[string]string{
		"euw": "EUW1", "EUW": "EUW1", "eune": "EUN1", "na": "NA1",
		"kr": "KR", "lan": "LA1", "las": "LA2", "": "EUW1",
	}
	for in, want := range cases {
		if got := platformID(in); got != want {
			t.Errorf("platformID(%q)=%q, attendu %q", in, got, want)
		}
	}
}

func TestParseDLIngame(t *testing.T) {
	raw := []byte(`{
		"playing": true,
		"support": true,
		"participants_list": [{
			"side": "BLUE",
			"champion_id": 51,
			"puu_id": "abc",
			"riot_id_name": "LaFouettance",
			"riot_id_tag_line": "EUW",
			"summoner_realtime_data": {
				"season_tier_info_dict": {
					"ranked_solo_5x5": {
						"tier": "GOLD",
						"division": 2,
						"league_points": 55,
						"wins": 3,
						"losses": 2
					}
				}
			},
			"participant_info": {
				"summoner_info_dict": { "ai_score_avg": 57.8 }
			}
		}]
	}`)
	ing, err := parseDLIngame(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !ing.Playing || len(ing.ParticipantsList) != 1 {
		t.Fatalf("ingame: %+v", ing)
	}
	p := ing.ParticipantsList[0]
	if p.RiotName != "LaFouettance" || p.Realtime.SeasonTier.Solo.Wins != 3 {
		t.Errorf("participant: %+v", p)
	}
	if int(p.Info.Summoner.AIScoreAvg) != 57 {
		t.Errorf("ai trunc: %v", p.Info.Summoner.AIScoreAvg)
	}
	id := identFromDL(p, "euw")
	patchIdentFromDL(id, p)
	if id.Rank != "G2" || id.Wins != 3 || !id.HasAI || id.AI != 57 {
		t.Errorf("patch: rank=%s wins=%d ai=%d has=%v", id.Rank, id.Wins, id.AI, id.HasAI)
	}
	if id.DeepLoL != "https://www.deeplol.gg/summoner/euw/LaFouettance-EUW" {
		t.Errorf("deeplol url: %s", id.DeepLoL)
	}
}

func TestUnwrapDL(t *testing.T) {
	inner := []byte(`{"playing":true,"participants_list":[]}`)
	wrapped := []byte(`{"data":{"playing":true,"participants_list":[]}}`)
	ing, err := parseDLIngame(wrapped)
	if err != nil || !ing.Playing {
		t.Errorf("unwrap: err=%v playing=%v", err, ing.Playing)
	}
	ing2, err := parseDLIngame(inner)
	if err != nil || !ing2.Playing {
		t.Errorf("direct: err=%v playing=%v", err, ing2.Playing)
	}
}

func TestDivFromRaw(t *testing.T) {
	if divFromRaw([]byte(`2`)) != "II" || divFromRaw([]byte(`"IV"`)) != "IV" {
		t.Errorf("div 2=%q IV=%q", divFromRaw([]byte(`2`)), divFromRaw([]byte(`"IV"`)))
	}
}

func TestDeepLoLTags(t *testing.T) {
	tags := buildDeepLoLTags(dlSummonerInfo{
		Wins: 4, Losses: 6, AIScoreAvg: 50, AIScoreAvg15: 60, AIScoreAvgLoss: 30,
		Tag: dlTagBag{LosingStreak: 3, ComebackUser: true, RunawayCount: 1},
	})
	got := map[string]ScoutTag{}
	for _, tg := range tags {
		got[tg.ID] = tg
	}
	for _, id := range []string{"ls", "early", "mental-", "afk", "cb"} {
		if _, ok := got[id]; !ok {
			t.Errorf("tag %s manquant: %+v", id, tags)
		}
	}
	if got["ls"].Label != "3L" || got["ls"].Tone != "bad" {
		t.Errorf("streak: %+v", got["ls"])
	}
	none := buildDeepLoLTags(dlSummonerInfo{})
	if none != nil {
		t.Errorf("vide attendu, obtenu %+v", none)
	}
}

func TestApplyCarryTags(t *testing.T) {
	a := &playerIdent{AI: 63, HasAI: true, DLGames: 10}
	b := &playerIdent{AI: 58, HasAI: true, DLGames: 8, Tags: []ScoutTag{{ID: "ls", Label: "2L", Tone: "bad"}}}
	c := &playerIdent{AI: 57, HasAI: true, DLGames: 3}
	d := &playerIdent{AI: 40, HasAI: true, DLGames: 10}
	applyCarryTags([]*playerIdent{a, b, c, d})
	if len(a.Tags) < 1 || a.Tags[0].ID != "carry" {
		t.Errorf("a devrait être Carry: %+v", a.Tags)
	}
	if len(b.Tags) < 2 || b.Tags[0].ID != "ls" || b.Tags[1].ID != "carry" {
		t.Errorf("carry après streak: %+v", b.Tags)
	}
	for _, p := range []*playerIdent{c, d} {
		for _, tg := range p.Tags {
			if tg.ID == "carry" {
				t.Errorf("pas Carry attendu pour AI=%d games=%d", p.AI, p.DLGames)
			}
		}
	}
}
