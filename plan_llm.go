package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Coach LLM : la table de draft part, le plan (texte + script audio) revient.
// Sans clé API : Claude Code (`claude`, abo Max) ou Codex (`codex`, ChatGPT+).
// Clé optionnelle : CDSCOUT_LLM_KEY / OPENAI_API_KEY, ou
// %LOCALAPPDATA%\lol-cd-scout\llm.json / llm.key.

const planSystemPrompt = `Tu es un coach LoL (peak D2, vise Challenger). On te donne un TABLEAU DE DRAFT (toi, matchup, invocs, runes, 5v5, bans). Tu rédiges le plan de JEU de CETTE game, pas un guide générique.

Style obligatoire, en français, tutoiement, comme un briefing vocal :
- Ouvre par « Tu joues [champ] [rôle] [tes runes/invocs] contre [ennemi] [sa config]. »
- Very early précis : niveau 1 (quel sort, trades AA+sort, bushes), niveau 2, niveau 3.
- Waves : wave 1 / 2 / 3 (slow push, crash, freeze) selon le matchup ET les invocs (Ghost vs TP ≠ Ghost vs assassin mid).
- Dis quand s'arrêter (stacks, rage, Q touché, gap closer up).
- Ghost/Flash/TP/Ignite changent le plan. Ne pas inventer une rune ennemie si elle est inconnue.
- Résumé court à la fin. But ≠ forcément kill avant 5 min.

JSON strict :
{"title":"...","sections":[{"id":"setup","title":"Config","body":"..."},{"id":"early","title":"Very early","body":"..."},{"id":"waves","title":"Waves","body":"..."},{"id":"windows","title":"Fenêtres","body":"..."},{"id":"summary","title":"Résumé","body":"..."}],"speak":"script oral continu, phrases courtes, prêt pour synthèse vocale FR"}
Dans body, sépare les paragraphes par \n\n. speak = tout le briefing à lire à voix haute (~90 secondes).`

type llmCfg struct {
	Key   string `json:"key"`
	Model string `json:"model"`
	Base  string `json:"base"`
	CLI   string `json:"cli"` // claude | codex | auto | off
}

func (c llmCfg) ready() bool {
	return strings.TrimSpace(c.Key) != "" || c.cliKind() != ""
}

func (c llmCfg) cliKind() string {
	s := strings.ToLower(strings.TrimSpace(c.CLI))
	switch s {
	case "off", "none", "0", "api":
		return ""
	case "claude", "codex":
		return s
	}
	return ""
}

type llmPlanOut struct {
	Title    string        `json:"title"`
	Sections []PlanSection `json:"sections"`
	Speak    string        `json:"speak"`
}

type llmChatResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

var (
	loadPlanLLMCfg = defaultLoadPlanLLMCfg
	callPlanLLM    = defaultCallPlanLLM
	planLLMHTTP    = &http.Client{Timeout: 45 * time.Second}

	planCacheMu sync.Mutex
	planCacheID string
	planCached  GamePlan
	planBusy    bool
	planErr     string
)

func clearPlanCache() {
	planCacheMu.Lock()
	planCacheID, planCached, planBusy, planErr = "", GamePlan{}, false, ""
	planCacheMu.Unlock()
}

func defaultLoadPlanLLMCfg() llmCfg {
	cfg := llmCfg{Base: "https://api.openai.com/v1", Model: "gpt-4o-mini"}
	if testing.Testing() {
		return llmCfg{}
	}
	dir := ideaDir()
	if b, err := os.ReadFile(filepath.Join(dir, "llm.json")); err == nil {
		_ = json.Unmarshal(b, &cfg)
	}
	if k := strings.TrimSpace(os.Getenv("CDSCOUT_LLM_KEY")); k != "" {
		cfg.Key = k
	}
	if cfg.Key == "" {
		if k := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); k != "" {
			cfg.Key = k
		}
	}
	if m := strings.TrimSpace(os.Getenv("CDSCOUT_LLM_MODEL")); m != "" {
		cfg.Model = m
	}
	if b := strings.TrimSpace(os.Getenv("CDSCOUT_LLM_BASE")); b != "" {
		cfg.Base = b
	}
	if cfg.Key == "" {
		if b, err := os.ReadFile(filepath.Join(dir, "llm.key")); err == nil {
			cfg.Key = strings.TrimSpace(string(b))
		}
	}
	if cfg.Base == "" {
		cfg.Base = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4o-mini"
	}
	if v := strings.TrimSpace(os.Getenv("CDSCOUT_LLM_CLI")); v != "" {
		cfg.CLI = v
	}
	cfg.CLI = resolvePlanCLIKind(cfg)
	return cfg
}

func noLLMWait() string {
	return "ChatGPT+ / Claude Max ne donnent pas de clé API. Le plus simple : installe Claude Code, ouvre un terminal, tape claude (login une fois), relance CD Scout. Autre option ChatGPT+ : Codex (codex login). Ou une clé dans %LOCALAPPDATA%\\lol-cd-scout\\llm.json."
}

func finishPlanLLM(skel GamePlan, in planInput) GamePlan {
	cfg := loadPlanLLMCfg()
	skel.LLM = cfg.ready()
	id := skel.ID
	if !cfg.ready() {
		skel.Status, skel.Wait = "nollm", noLLMWait()
		return skel
	}
	planCacheMu.Lock()
	if planCacheID == id && planCached.Ready {
		p := planCached
		planCacheMu.Unlock()
		return p
	}
	if planCacheID == id && planBusy {
		skel.Status, skel.Wait = "loading", "Le coach rédige le plan…"
		planCacheMu.Unlock()
		return skel
	}
	if planCacheID == id && planErr != "" {
		skel.Status, skel.Wait = "error", "LLM : "+planErr+" — clique Réessayer."
		planCacheMu.Unlock()
		return skel
	}
	planCacheID, planBusy, planErr, planCached = id, true, "", GamePlan{}
	planCacheMu.Unlock()

	go func(cfg llmCfg, in planInput, skel GamePlan) {
		out, err := callPlanLLM(cfg, draftTable(in))
		planCacheMu.Lock()
		defer planCacheMu.Unlock()
		if planCacheID != skel.ID {
			return
		}
		planBusy = false
		if err != nil {
			planErr = err.Error()
			return
		}
		p := skel
		p.Ready, p.Status, p.LLM = true, "ready", true
		if out.Title != "" {
			p.Title = out.Title
		}
		p.Sections, p.Speak = out.Sections, strings.TrimSpace(out.Speak)
		if p.Speak == "" {
			var b strings.Builder
			for i, s := range p.Sections {
				if i > 0 {
					b.WriteString("\n\n")
				}
				if s.Title != "" {
					b.WriteString(s.Title + ". ")
				}
				b.WriteString(s.Body)
			}
			p.Speak = b.String()
		}
		planCached, planErr = p, ""
	}(cfg, in, skel)

	skel.Status, skel.Wait = "loading", "Le coach rédige le plan…"
	return skel
}

func draftTable(in planInput) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Patch: %s\n", strings.TrimSpace(in.Patch))
	if in.ARAM {
		b.WriteString("Mode: ARAM\n")
	} else {
		b.WriteString("Mode: Faille de l'invocateur\n")
	}
	fmt.Fprintf(&b, "Ton rôle: %s\n", in.Role)
	fmt.Fprintf(&b, "TOI: %s | %s | portée %d | %s\n", in.Me.Name, joinPlus(in.Me), in.Me.Range, strings.Join(in.Me.Tags, ", "))
	if in.Me.Window != "" {
		fmt.Fprintf(&b, "Note CD Scout (toi): %s\n", in.Me.Window)
	}
	fmt.Fprintf(&b, "MATCHUP: %s | %s | portée %d | %s\n", in.Them.Name, joinPlus(in.Them), in.Them.Range, strings.Join(in.Them.Tags, ", "))
	if in.Them.Window != "" {
		fmt.Fprintf(&b, "Fenêtre curated (ennemi): %s\n", in.Them.Window)
	}
	if in.AllyJungle != "" {
		fmt.Fprintf(&b, "JGL allié: %s\n", in.AllyJungle)
	}
	if in.EnemyJungle != "" {
		fmt.Fprintf(&b, "JGL ennemi: %s\n", in.EnemyJungle)
	}
	if len(in.Bans) > 0 {
		fmt.Fprintf(&b, "Bans: %s\n", strings.Join(in.Bans, ", "))
	}
	b.WriteString("\nDRAFT\nCôté | Rôle | Champion | Invocs | Rune | Flags\n")
	writeRows := func(rows []planRow) {
		for _, r := range rows {
			flags := []string{}
			if r.Me {
				flags = append(flags, "TOI")
			}
			if r.Lane {
				flags = append(flags, "LANE")
			}
			fmt.Fprintf(&b, "%s | %s | %s | %s | %s | %s\n", r.Side, r.Role, r.Name, strings.Join(r.Sums, "+"), r.Keystone, strings.Join(flags, ","))
		}
	}
	if len(in.Allies)+len(in.Enemies) == 0 {
		writeRows([]planRow{rowFromFighter(in.Me, "ALLIÉ", true, true), rowFromFighter(in.Them, "ENNEMI", false, true)})
	} else {
		writeRows(in.Allies)
		writeRows(in.Enemies)
	}
	b.WriteString("\nRédige le plan early de CETTE game au JSON demandé.")
	return b.String()
}

func defaultCallPlanLLM(cfg llmCfg, table string) (llmPlanOut, error) {
	if strings.TrimSpace(cfg.Key) != "" {
		return callPlanLLMHTTP(cfg, table)
	}
	if kind := cfg.cliKind(); kind != "" {
		return callPlanLLMCLI(cfg, kind, table)
	}
	return llmPlanOut{}, errors.New("pas de LLM (Claude Code / Codex / clé API)")
}

func callPlanLLMHTTP(cfg llmCfg, table string) (llmPlanOut, error) {
	base := strings.TrimRight(cfg.Base, "/")
	payload, err := json.Marshal(map[string]any{
		"model":           cfg.Model,
		"temperature":     0.45,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": planSystemPrompt},
			{"role": "user", "content": table},
		},
	})
	if err != nil {
		return llmPlanOut{}, err
	}
	req, err := http.NewRequest(http.MethodPost, base+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return llmPlanOut{}, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Key)
	req.Header.Set("Content-Type", "application/json")
	res, err := planLLMHTTP.Do(req)
	if err != nil {
		return llmPlanOut{}, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		msg := strings.TrimSpace(string(raw))
		if len(msg) > 240 {
			msg = msg[:240]
		}
		if res.StatusCode == 401 || res.StatusCode == 403 {
			return llmPlanOut{}, errors.New("clé refusée")
		}
		return llmPlanOut{}, fmt.Errorf("HTTP %d %s", res.StatusCode, msg)
	}
	var chat llmChatResp
	if err := json.Unmarshal(raw, &chat); err != nil {
		return llmPlanOut{}, err
	}
	if chat.Error != nil && chat.Error.Message != "" {
		return llmPlanOut{}, errors.New(chat.Error.Message)
	}
	if len(chat.Choices) == 0 {
		return llmPlanOut{}, errors.New("réponse LLM vide")
	}
	return parseLLMPlan(chat.Choices[0].Message.Content)
}

func parseLLMPlan(content string) (llmPlanOut, error) {
	s := strings.TrimSpace(content)
	if i := strings.Index(s, "```"); i >= 0 {
		s = s[i+3:]
		s = strings.TrimPrefix(s, "json")
		if j := strings.Index(s, "```"); j >= 0 {
			s = s[:j]
		}
		s = strings.TrimSpace(s)
	}
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			s = s[i : j+1]
		}
	}
	var out llmPlanOut
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return llmPlanOut{}, err
	}
	if len(out.Sections) == 0 && out.Speak == "" {
		return llmPlanOut{}, errors.New("JSON sans sections")
	}
	return out, nil
}
