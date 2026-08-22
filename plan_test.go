package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDraftTableContainsMatchup(t *testing.T) {
	in := planInput{
		Role: "TOP", Patch: "16.16.1",
		Me:   planFighter{Key: 21, Name: "Miss Fortune", Sums: []string{"Ghost", "Flash"}, Keystone: "Premier coup", Range: 550, Tags: []string{"Marksman"}},
		Them: planFighter{Key: 150, Name: "Gnar", Sums: []string{"TP", "Flash"}, Keystone: "Jeu de jambes", Range: 400, Tags: []string{"Fighter"}},
		EnemyJungle: "Elise",
		Allies: []planRow{
			{Side: "ALLIÉ", Role: "TOP", Name: "Miss Fortune", Sums: []string{"Ghost", "Flash"}, Keystone: "Premier coup", Me: true, Lane: true},
			{Side: "ALLIÉ", Role: "JGL", Name: "Lee Sin"},
		},
		Enemies: []planRow{
			{Side: "ENNEMI", Role: "TOP", Name: "Gnar", Sums: []string{"TP", "Flash"}, Keystone: "Jeu de jambes", Lane: true},
			{Side: "ENNEMI", Role: "JGL", Name: "Elise"},
		},
		Bans: []string{"Yasuo"},
	}
	tab := strings.ToLower(draftTable(in))
	for _, want := range []string{"miss fortune", "gnar", "ghost", "premier coup", "jeu de jambes", "elise", "yasuo", "toi"} {
		if !strings.Contains(tab, want) {
			t.Errorf("table sans %q\n%s", want, tab)
		}
	}
}

func TestPlanWaitsWithoutMatchup(t *testing.T) {
	p := buildPlan(planInput{Role: "TOP", Me: planFighter{Key: 21, Name: "Miss Fortune"}})
	if p.Ready || p.Wait == "" {
		t.Fatalf("attendu wait matchup, obtenu %+v", p)
	}
}

func TestPlanNeedsLLMKey(t *testing.T) {
	clearPlanCache()
	loadPlanLLMCfg = func() llmCfg { return llmCfg{} }
	t.Cleanup(func() { loadPlanLLMCfg = defaultLoadPlanLLMCfg })
	p := buildPlan(planInput{
		Role: "TOP",
		Me:   planFighter{Key: 21, Name: "Miss Fortune", Sums: []string{"Ghost"}},
		Them: planFighter{Key: 150, Name: "Gnar", Sums: []string{"TP"}},
	})
	if p.Ready || p.Status != "nollm" {
		t.Fatalf("sans clé: %+v", p)
	}
	if !strings.Contains(p.Wait, "Claude") || !strings.Contains(p.Wait, "llm.json") {
		t.Errorf("wait devrait expliquer Claude Code / llm.json: %s", p.Wait)
	}
}

func TestPlanCLIWithoutKey(t *testing.T) {
	clearPlanCache()
	loadPlanLLMCfg = func() llmCfg { return llmCfg{CLI: "claude"} }
	callPlanLLM = func(cfg llmCfg, table string) (llmPlanOut, error) {
		if cfg.cliKind() != "claude" {
			t.Errorf("CLI=%q", cfg.CLI)
		}
		return llmPlanOut{
			Title:    "MF top vs Gnar",
			Sections: []PlanSection{{ID: "early", Title: "Very early", Body: "Niveau 1 Q."}},
			Speak:    "Tu joues Miss Fortune top contre Gnar.",
		}, nil
	}
	t.Cleanup(func() {
		loadPlanLLMCfg = defaultLoadPlanLLMCfg
		callPlanLLM = defaultCallPlanLLM
		clearPlanCache()
	})
	in := planInput{
		Role: "TOP",
		Me:   planFighter{Key: 21, Name: "Miss Fortune", Sums: []string{"Ghost"}},
		Them: planFighter{Key: 150, Name: "Gnar", Sums: []string{"TP"}},
	}
	p := buildPlan(in)
	if p.Status != "loading" && !p.Ready {
		t.Fatalf("CLI sans clé devrait lancer le coach, obtenu %+v", p)
	}
	for i := 0; i < 80; i++ {
		time.Sleep(5 * time.Millisecond)
		p = buildPlan(in)
		if p.Ready {
			break
		}
		if p.Status == "error" {
			t.Fatalf("erreur CLI: %s", p.Wait)
		}
	}
	if !p.Ready {
		t.Fatal("plan CLI pas prêt")
	}
}

func TestResolvePlanCLIKind(t *testing.T) {
	lookPlanCLIBin = func(name string) string {
		if name == "claude" {
			return `C:\claude.exe`
		}
		return ""
	}
	t.Cleanup(func() { lookPlanCLIBin = defaultLookPlanCLIBin })
	if got := resolvePlanCLIKind(llmCfg{}); got != "claude" {
		t.Fatalf("auto: %s", got)
	}
	if got := resolvePlanCLIKind(llmCfg{CLI: "codex"}); got != "codex" {
		t.Fatalf("forcé: %s", got)
	}
	if got := resolvePlanCLIKind(llmCfg{CLI: "off"}); got != "" {
		t.Fatalf("off: %s", got)
	}
	if got := resolvePlanCLIKind(llmCfg{Key: "sk", CLI: "auto"}); got != "" {
		t.Fatalf("clé API: %s", got)
	}
}

func TestLookClaudeInCursorExtension(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".cursor", "extensions", "anthropic.claude-code-2.1.233-win32-x64", "resources", "native-binary")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(dir, "claude.exe")
	if err := os.WriteFile(exe, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := lookClaudeInIDEExtensions(root)
	if got != exe {
		t.Fatalf("plugin Cursor pas vu: %q", got)
	}
}

func TestPlanLLMFillsFromTable(t *testing.T) {
	clearPlanCache()
	var gotTable string
	loadPlanLLMCfg = func() llmCfg { return llmCfg{Key: "test", Model: "x", Base: "http://x"} }
	callPlanLLM = func(cfg llmCfg, table string) (llmPlanOut, error) {
		gotTable = table
		return llmPlanOut{
			Title: "MF top vs Gnar",
			Sections: []PlanSection{{ID: "early", Title: "Very early", Body: "Niveau 1 Q. Crash wave 3."}},
			Speak:    "Tu joues Miss Fortune top Premier coup plus Ghost contre Gnar Jeu de jambes plus TP.",
		}, nil
	}
	t.Cleanup(func() {
		loadPlanLLMCfg = defaultLoadPlanLLMCfg
		callPlanLLM = defaultCallPlanLLM
		clearPlanCache()
	})
	in := planInput{
		Role: "TOP",
		Me:   planFighter{Key: 21, Name: "Miss Fortune", Sums: []string{"Ghost", "Flash"}, Keystone: "Premier coup", KeyID: 8369},
		Them: planFighter{Key: 150, Name: "Gnar", Sums: []string{"TP", "Flash"}, Keystone: "Jeu de jambes"},
	}
	p := buildPlan(in)
	if p.Status != "loading" {
		t.Fatalf("premier appel loading, obtenu %s", p.Status)
	}
	for i := 0; i < 80; i++ {
		time.Sleep(5 * time.Millisecond)
		p = buildPlan(in)
		if p.Ready {
			break
		}
		if p.Status == "error" {
			t.Fatalf("erreur LLM: %s", p.Wait)
		}
	}
	if !p.Ready {
		t.Fatal("plan LLM pas prêt")
	}
	if !strings.Contains(strings.ToLower(gotTable), "miss fortune") || !strings.Contains(strings.ToLower(gotTable), "gnar") {
		t.Errorf("le LLM n'a pas reçu la table: %s", gotTable)
	}
	if !strings.Contains(p.Speak, "Miss Fortune") {
		t.Errorf("speak: %s", p.Speak)
	}
}

func TestParseLLMPlanFence(t *testing.T) {
	raw := "```json\n{\"title\":\"T\",\"sections\":[{\"id\":\"early\",\"title\":\"Very early\",\"body\":\"go\"}],\"speak\":\"allo\"}\n```"
	out, err := parseLLMPlan(raw)
	if err != nil {
		t.Fatal(err)
	}
	if out.Title != "T" || out.Speak != "allo" || len(out.Sections) != 1 {
		t.Fatalf("%+v", out)
	}
}

func TestSummShort(t *testing.T) {
	if summShort("SummonerHaste", "Ghost", "") != "Ghost" {
		t.Errorf("ghost: %s", summShort("SummonerHaste", "Ghost", ""))
	}
	if summShort("SummonerTeleport", "Téléportation", "") != "TP" {
		t.Errorf("tp: %s", summShort("SummonerTeleport", "Téléportation", ""))
	}
	if lcuSummShort[6] != "Ghost" || lcuSummShort[12] != "TP" {
		t.Errorf("ids LCU: %+v", lcuSummShort)
	}
}
