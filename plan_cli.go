package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Claude Max / ChatGPT+ passent par le CLI local (login abo), pas une clé API.
var lookPlanCLIBin = defaultLookPlanCLIBin

func resolvePlanCLIKind(cfg llmCfg) string {
	s := strings.ToLower(strings.TrimSpace(cfg.CLI))
	switch s {
	case "off", "none", "0", "api":
		return ""
	case "claude", "codex":
		return s
	}
	if strings.TrimSpace(cfg.Key) != "" {
		return ""
	}
	if lookPlanCLIBin("claude") != "" {
		return "claude"
	}
	if lookPlanCLIBin("codex") != "" {
		return "codex"
	}
	return ""
}

func defaultLookPlanCLIBin(name string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	home, _ := os.UserHomeDir()
	local := os.Getenv("LOCALAPPDATA")
	appdata := os.Getenv("APPDATA")
	cands := []string{
		filepath.Join(home, ".local", "bin", name+".exe"),
		filepath.Join(home, ".local", "bin", name),
		filepath.Join(local, name, name+".exe"),
		filepath.Join(appdata, "npm", name+".cmd"),
		filepath.Join(appdata, "npm", name+".exe"),
		filepath.Join(appdata, "npm", name),
	}
	for _, c := range cands {
		if c == "" {
			continue
		}
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	if name == "claude" {
		return lookClaudeInIDEExtensions(home)
	}
	return ""
}

// Plugin Cursor / VS Code : le claude.exe est dans l'extension, pas dans le PATH.
func lookClaudeInIDEExtensions(home string) string {
	if home == "" {
		return ""
	}
	var best string
	var bestT time.Time
	for _, ide := range []string{".cursor", ".vscode", ".vscode-insiders"} {
		pats := []string{
			filepath.Join(home, ide, "extensions", "anthropic.claude-code-*", "resources", "native-binary", "claude.exe"),
			filepath.Join(home, ide, "extensions", "anthropic.claude-code-*", "resources", "native-binary", "claude"),
		}
		for _, pat := range pats {
			matches, _ := filepath.Glob(pat)
			for _, m := range matches {
				st, err := os.Stat(m)
				if err != nil || st.IsDir() {
					continue
				}
				if best == "" || st.ModTime().After(bestT) {
					best, bestT = m, st.ModTime()
				}
			}
		}
	}
	return best
}

func callPlanLLMCLI(cfg llmCfg, kind, table string) (llmPlanOut, error) {
	bin := lookPlanCLIBin(kind)
	if bin == "" {
		if kind == "claude" {
			return llmPlanOut{}, errors.New("Claude Code introuvable. Installe-le, tape claude dans un terminal (login une fois), relance CD Scout")
		}
		return llmPlanOut{}, errors.New("Codex introuvable. Installe Codex CLI, tape codex login, relance CD Scout")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd, err := planCLICmd(ctx, cfg, kind, bin, table)
	if err != nil {
		return llmPlanOut{}, err
	}
	cmd.Dir = os.TempDir()
	cmd.SysProcAttr = windowsHiddenProcAttr()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return llmPlanOut{}, errors.New("timeout coach (2 min) — relance ou réessaie")
		}
		low := strings.ToLower(msg)
		if kind == "claude" && (strings.Contains(low, "not logged") || strings.Contains(low, "login") || strings.Contains(low, "unauthorized")) {
			return llmPlanOut{}, errors.New("Claude Code pas connecté : tape claude dans un terminal, login, relance")
		}
		if kind == "codex" && (strings.Contains(low, "not logged") || strings.Contains(low, "login") || strings.Contains(low, "auth")) {
			return llmPlanOut{}, errors.New("Codex pas connecté : tape codex login, relance CD Scout")
		}
		if len(msg) > 280 {
			msg = msg[:280]
		}
		return llmPlanOut{}, fmt.Errorf("%s : %s", kind, msg)
	}
	return parseLLMPlan(string(out))
}

func planCLICmd(ctx context.Context, cfg llmCfg, kind, bin, table string) (*exec.Cmd, error) {
	switch kind {
	case "claude":
		args := []string{
			"-p", "--bare", "--tools=", "--max-turns", "1",
			"--no-session-persistence", "--dangerously-skip-permissions",
			"--output-format", "text",
		}
		if m := cliModel(cfg, kind); m != "" {
			args = append(args, "--model", m)
		}
		args = append(args, "--system-prompt", planSystemPrompt, table)
		cmd := planCLICommand(ctx, bin, args...)
		cmd.Env = stripEnvPrefixes("ANTHROPIC_API_KEY=", "CLAUDE_API_KEY=")
		return cmd, nil
	case "codex":
		args := []string{"exec", "--sandbox", "read-only"}
		if m := cliModel(cfg, kind); m != "" {
			args = append(args, "--model", m)
		}
		args = append(args, "-")
		cmd := planCLICommand(ctx, bin, args...)
		cmd.Stdin = strings.NewReader(planSystemPrompt + "\n\n" + table)
		cmd.Env = stripEnvPrefixes("OPENAI_API_KEY=")
		return cmd, nil
	default:
		return nil, fmt.Errorf("CLI inconnue : %s", kind)
	}
}

func planCLICommand(ctx context.Context, bin string, args ...string) *exec.Cmd {
	ext := strings.ToLower(filepath.Ext(bin))
	if ext == ".cmd" || ext == ".bat" {
		all := append([]string{"/d", "/s", "/c", bin}, args...)
		return exec.CommandContext(ctx, "cmd", all...)
	}
	return exec.CommandContext(ctx, bin, args...)
}

func cliModel(cfg llmCfg, kind string) string {
	m := strings.TrimSpace(cfg.Model)
	if m == "" {
		return ""
	}
	low := strings.ToLower(m)
	if kind == "claude" && (low == "gpt-4o-mini" || strings.HasPrefix(low, "gpt-") || strings.HasPrefix(low, "o1") || strings.HasPrefix(low, "o3") || strings.HasPrefix(low, "o4")) {
		return ""
	}
	if kind == "codex" && (low == "sonnet" || low == "opus" || low == "haiku" || strings.Contains(low, "claude")) {
		return ""
	}
	return m
}

func stripEnvPrefixes(prefixes ...string) []string {
	env := os.Environ()
	out := make([]string, 0, len(env))
	for _, e := range env {
		skip := false
		for _, p := range prefixes {
			if len(e) >= len(p) && strings.EqualFold(e[:len(p)], p) {
				skip = true
				break
			}
		}
		if !skip {
			out = append(out, e)
		}
	}
	return out
}
