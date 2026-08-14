package main

// Propositions de features depuis l'UI. Rien de secret dans le binaire :
// GitHub via GH_TOKEN / fichier local / `gh auth`, sinon formulaire GitHub
// dans le navigateur. Si le texte est dangereux, on ne publie pas : mail
// (.eml) ou SMS (app Phone Link), d'après notify.json ou l'env.

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"
)

const (
	ideaMaxTitle = 100
	ideaMaxBody  = 4000
	ideaMaxImage = 2 << 20
	ideaCooldown = 15 * time.Second
)

// Surcharge possible sans recompiler : CDSCOUT_NOTIFY_EMAIL / CDSCOUT_NOTIFY_SMS
// ou %LOCALAPPDATA%\lol-cd-scout\notify.json.
const (
	ideaDefaultEmail = ""
	ideaDefaultSMS   = ""
)

type ideaReq struct {
	Title     string
	Body      string
	Image     []byte
	ImageExt  string
	ImageMIME string
}

type ideaOut struct {
	OK      bool   `json:"ok"`
	Via     string `json:"via,omitempty"`
	URL     string `json:"url,omitempty"`
	Message string `json:"message"`
	Unsafe  bool   `json:"unsafe,omitempty"`
}

type ideaNotify struct {
	Email string `json:"email"`
	SMS   string `json:"sms"`
}

var (
	ideaMu          sync.Mutex
	ideaLastAt      time.Time
	ideaDirOverride string
	ideaHTTP        = &http.Client{Timeout: 20 * time.Second}

	lookupIdeaToken = defaultLookupIdeaToken
	postIdeaIssue   = defaultPostIdeaIssue
	openIdeaURL     = defaultOpenIdeaURL
)

func ideaDir() string {
	if ideaDirOverride != "" {
		return ideaDirOverride
	}
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		return filepath.Join(local, "lol-cd-scout")
	}
	return filepath.Join(os.TempDir(), "lol-cd-scout")
}

func apiIdea(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		email, sms := loadIdeaNotify()
		writeJSON(w, map[string]any{
			"github": lookupIdeaToken() != "",
			"mail":   email != "",
			"sms":    sms != "",
		})
	case http.MethodPost:
		req, err, code := parseIdeaReq(r)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		writeJSON(w, submitIdea(req))
	default:
		http.Error(w, "GET ou POST", http.StatusMethodNotAllowed)
	}
}

func parseIdeaReq(r *http.Request) (ideaReq, error, int) {
	var req ideaReq
	r.Body = http.MaxBytesReader(nil, r.Body, ideaMaxImage+ideaMaxBody+1<<16)
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/") {
		if err := r.ParseMultipartForm(ideaMaxImage + 1<<20); err != nil {
			return req, errors.New("formulaire trop lourd (image ≤ 2 Mo)"), http.StatusRequestEntityTooLarge
		}
		req.Title = r.FormValue("title")
		req.Body = r.FormValue("body")
		if f, hdr, err := r.FormFile("image"); err == nil {
			defer f.Close()
			raw, err := io.ReadAll(io.LimitReader(f, ideaMaxImage+1))
			if err != nil {
				return req, errors.New("image illisible"), http.StatusBadRequest
			}
			if len(raw) > ideaMaxImage {
				return req, errors.New("image trop lourde (max 2 Mo)"), http.StatusRequestEntityTooLarge
			}
			name := ""
			if hdr != nil {
				name = hdr.Filename
			}
			ext, mime, ok := sniffIdeaImage(raw, name)
			if !ok {
				return req, errors.New("image : PNG, JPEG, WebP ou GIF"), http.StatusBadRequest
			}
			req.Image, req.ImageExt, req.ImageMIME = raw, ext, mime
		}
	} else {
		var in struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return req, errors.New("JSON invalide"), http.StatusBadRequest
		}
		req.Title, req.Body = in.Title, in.Body
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	if err := validateIdea(req); err != nil {
		return req, err, http.StatusBadRequest
	}
	return req, nil, 0
}

func validateIdea(req ideaReq) error {
	if n := len([]rune(req.Title)); n < 3 || n > ideaMaxTitle {
		return fmt.Errorf("titre : 3 à %d caractères", ideaMaxTitle)
	}
	if n := len([]rune(req.Body)); n < 10 || n > ideaMaxBody {
		return fmt.Errorf("description : 10 à %d caractères", ideaMaxBody)
	}
	return nil
}

func submitIdea(req ideaReq) ideaOut {
	ideaMu.Lock()
	if time.Since(ideaLastAt) < ideaCooldown && !ideaLastAt.IsZero() {
		ideaMu.Unlock()
		return ideaOut{Message: "Patiente 15 s entre deux propositions."}
	}
	ideaLastAt = time.Now()
	ideaMu.Unlock()

	archiveIdea(req)
	if ideaUnsafe(req.Title, req.Body) {
		return notifyIdeaPrivate(req)
	}
	if tok := lookupIdeaToken(); tok != "" {
		if u, err := postIdeaIssue(tok, req); err == nil && u != "" {
			return ideaOut{OK: true, Via: "github", URL: u, Message: "Issue GitHub créée."}
		}
	}
	u := githubNewIssueURL(req)
	openIdeaURL(u)
	msg := "GitHub ouvert — clique Submit pour créer l'issue."
	if len(req.Image) > 0 {
		msg += " Colle la capture (Ctrl+V) dans le message."
	}
	return ideaOut{OK: true, Via: "form", URL: u, Message: msg}
}

func notifyIdeaPrivate(req ideaReq) ideaOut {
	email, sms := loadIdeaNotify()
	if len(req.Image) > 0 && email != "" {
		if err := sendIdeaMail(email, req); err == nil {
			return ideaOut{OK: true, Via: "mail", Unsafe: true, Message: "Contenu non publié. Mail ouvert — clique Envoyer."}
		}
	}
	if sms != "" {
		if err := sendIdeaSMS(sms, req); err == nil {
			return ideaOut{OK: true, Via: "sms", Unsafe: true, Message: "Contenu non publié. SMS ouvert — clique Envoyer."}
		}
	}
	if err := sendIdeaMail(email, req); err == nil {
		msg := "Contenu non publié. Mail ouvert — clique Envoyer."
		if email == "" {
			msg = "Contenu non publié. Complète le destinataire du mail (ou notify.json)."
		}
		return ideaOut{OK: true, Via: "mail", Unsafe: true, Message: msg}
	}
	return ideaOut{Unsafe: true, Message: "Contenu non publié (trop sensible pour GitHub). Ajoute un mail dans %LOCALAPPDATA%\\lol-cd-scout\\notify.json."}
}

func ideaUnsafe(title, body string) bool {
	n := ideaNorm(title + "\n" + body)
	for _, p := range ideaDanger {
		if strings.Contains(n, " "+p+" ") {
			return true
		}
	}
	return false
}

func ideaNorm(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte(' ')
	prevSpace := true
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			prevSpace = false
			continue
		}
		if !prevSpace {
			b.WriteByte(' ')
			prevSpace = true
		}
	}
	if !prevSpace {
		b.WriteByte(' ')
	}
	return b.String()
}

// Phrases (déjà en minuscules, mots séparés). « exploit » / « hack » seuls
// sont trop vagues (fenêtre de punish, overlay).
var ideaDanger = []string{
	"keylogger", "key logger", "enregistreur de frappe",
	"malware", "ransomware", "rancongiciel", "rançongiciel",
	"stealer", "infostealer", "token grabber",
	"remote access trojan",
	"dll inject", "dll injection", "injecter une dll", "injection dll",
	"process hollowing", "reverse shell", "meterpreter",
	"phishing", "hameçonnage",
	"csam", "child porn", "child sexual", "pornographie infantile",
	"steal password", "voler le mot de passe", "voler les identifiants",
	"account steal", "token discord", "github token",
	"cheat engine", "wallhack", "aimbot",
	"script d injection", "injecter dans league",
	"ddos", "botnet", "crypto miner",
	"0day", "0 day", "zero day",
}

func sniffIdeaImage(raw []byte, name string) (ext, mime string, ok bool) {
	if len(raw) < 12 {
		return "", "", false
	}
	switch {
	case bytes.HasPrefix(raw, []byte("\x89PNG\r\n\x1a\n")):
		return "png", "image/png", true
	case bytes.HasPrefix(raw, []byte{0xff, 0xd8, 0xff}):
		return "jpg", "image/jpeg", true
	case bytes.HasPrefix(raw, []byte("GIF87a")) || bytes.HasPrefix(raw, []byte("GIF89a")):
		return "gif", "image/gif", true
	case bytes.HasPrefix(raw, []byte("RIFF")) && string(raw[8:12]) == "WEBP":
		return "webp", "image/webp", true
	}
	_ = name
	return "", "", false
}

func ideaSlug(title string) string {
	n := ideaNorm(title)
	parts := strings.Fields(n)
	if len(parts) > 6 {
		parts = parts[:6]
	}
	s := strings.Join(parts, "-")
	if s == "" {
		s = "idee"
	}
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

func archiveIdea(req ideaReq) {
	dir := filepath.Join(ideaDir(), "ideas")
	if os.MkdirAll(dir, 0o755) != nil {
		return
	}
	stamp := time.Now().Format("20060102-150405")
	base := stamp + "-" + ideaSlug(req.Title)
	meta := map[string]any{"title": req.Title, "body": req.Body, "at": stamp, "unsafe": ideaUnsafe(req.Title, req.Body)}
	if len(req.Image) > 0 {
		img := base + "." + req.ImageExt
		if os.WriteFile(filepath.Join(dir, img), req.Image, 0o644) == nil {
			meta["image"] = img
		}
	}
	b, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, base+".json"), b, 0o644)
}

func loadIdeaNotify() (email, sms string) {
	email = strings.TrimSpace(os.Getenv("CDSCOUT_NOTIFY_EMAIL"))
	sms = strings.TrimSpace(os.Getenv("CDSCOUT_NOTIFY_SMS"))
	if email == "" {
		email = ideaDefaultEmail
	}
	if sms == "" {
		sms = ideaDefaultSMS
	}
	b, err := os.ReadFile(filepath.Join(ideaDir(), "notify.json"))
	if err == nil {
		var n ideaNotify
		if json.Unmarshal(b, &n) == nil {
			if email == "" {
				email = strings.TrimSpace(n.Email)
			}
			if sms == "" {
				sms = strings.TrimSpace(n.SMS)
			}
		}
	}
	return sanitizeEmail(email), sanitizeSMS(sms)
}

func sanitizeEmail(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || strings.ContainsAny(s, "\r\n<>") || strings.Count(s, "@") != 1 {
		return ""
	}
	return s
}

func sanitizeSMS(s string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(s) {
		if r == '+' && b.Len() == 0 {
			b.WriteRune(r)
			continue
		}
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if len(out) < 8 {
		return ""
	}
	return out
}

func defaultLookupIdeaToken() string {
	if testing.Testing() {
		return ""
	}
	if t := strings.TrimSpace(os.Getenv("GH_TOKEN")); t != "" {
		return t
	}
	if b, err := os.ReadFile(filepath.Join(ideaDir(), "github.token")); err == nil {
		if t := strings.TrimSpace(string(b)); t != "" {
			return t
		}
	}
	gh := ghBinary()
	if gh == "" {
		return ""
	}
	cmd := exec.Command(gh, "auth", "token")
	cmd.SysProcAttr = windowsHiddenProcAttr()
	cmd.Stderr = io.Discard
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func ghBinary() string {
	if p, err := exec.LookPath("gh"); err == nil {
		return p
	}
	cands := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "GitHub CLI", "gh.exe"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "GitHub CLI", "gh.exe"),
	}
	for _, c := range cands {
		if c == "" {
			continue
		}
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func defaultPostIdeaIssue(token string, req ideaReq) (string, error) {
	if testing.Testing() {
		return "", errors.New("github désactivé pendant les tests")
	}
	return postGitHubIssueAt("https://api.github.com", token, req)
}

func postGitHubIssueAt(apiBase, token string, req ideaReq) (string, error) {
	if token == "" {
		return "", errors.New("no token")
	}
	payload, _ := json.Marshal(map[string]any{
		"title": "[idée] " + req.Title,
		"body":  ideaMarkdown(req, ""),
	})
	httpReq, err := http.NewRequest(http.MethodPost, strings.TrimRight(apiBase, "/")+"/repos/"+updateRepo+"/issues", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Accept", "application/vnd.github+json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "LoL-CD-Scout/0.2")
	httpReq.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	res, err := ideaHTTP.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("GitHub HTTP %d", res.StatusCode)
	}
	var out struct {
		HTMLURL string `json:"html_url"`
		Number  int    `json:"number"`
		ID      int64  `json:"id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || out.HTMLURL == "" {
		return "", errors.New("réponse GitHub invalide")
	}
	if imgURL := uploadGitHubImage(apiBase, token, out.ID, req); imgURL != "" {
		patched := ideaMarkdown(req, imgURL)
		_ = patchGitHubIssue(apiBase, token, out.Number, patched)
	}
	return out.HTMLURL, nil
}

func uploadGitHubImage(apiBase, token string, issueID int64, req ideaReq) string {
	if len(req.Image) == 0 || issueID == 0 {
		return ""
	}
	host := "https://uploads.github.com"
	if strings.Contains(apiBase, "127.0.0.1") || strings.Contains(apiBase, "localhost") {
		host = apiBase
	}
	u := fmt.Sprintf("%s/repos/%s/issues/%d/attachments?name=maquette.%s", strings.TrimRight(host, "/"), updateRepo, issueID, req.ImageExt)
	httpReq, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(req.Image))
	if err != nil {
		return ""
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Accept", "application/vnd.github+json")
	httpReq.Header.Set("Content-Type", req.ImageMIME)
	httpReq.Header.Set("User-Agent", "LoL-CD-Scout/0.2")
	res, err := ideaHTTP.Do(httpReq)
	if err != nil {
		return ""
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return ""
	}
	var out struct {
		URL         string `json:"url"`
		AssetURL    string `json:"asset_url"`
		BrowserURL  string `json:"browser_download_url"`
		ContentType string `json:"content_type"`
	}
	if json.NewDecoder(res.Body).Decode(&out) != nil {
		return ""
	}
	for _, s := range []string{out.URL, out.AssetURL, out.BrowserURL} {
		if strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://127.0.0.1") {
			return s
		}
	}
	return ""
}

func patchGitHubIssue(apiBase, token string, number int, body string) error {
	payload, _ := json.Marshal(map[string]any{"body": body})
	httpReq, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/repos/%s/issues/%d", strings.TrimRight(apiBase, "/"), updateRepo, number), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Accept", "application/vnd.github+json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "LoL-CD-Scout/0.2")
	res, err := ideaHTTP.Do(httpReq)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("patch HTTP %d", res.StatusCode)
	}
	return nil
}

func ideaMarkdown(req ideaReq, imageURL string) string {
	var b strings.Builder
	b.WriteString("## Description\n\n")
	b.WriteString(req.Body)
	b.WriteString("\n\n")
	if imageURL != "" {
		b.WriteString("## Maquette\n\n![maquette](")
		b.WriteString(imageURL)
		b.WriteString(")\n\n")
	} else if len(req.Image) > 0 {
		b.WriteString("*Une capture était jointe — à coller en commentaire si elle n'apparaît pas.*\n\n")
	}
	b.WriteString("---\n*Envoyé depuis CD Scout ")
	if appVersion != "" {
		b.WriteString(appVersion)
	} else {
		b.WriteString("dev")
	}
	b.WriteString("*\n")
	return b.String()
}

func githubNewIssueURL(req ideaReq) string {
	body := req.Body
	if len([]rune(body)) > 1800 {
		body = string([]rune(body)[:1800]) + "\n\n…(tronqué)"
	}
	q := url.Values{}
	q.Set("title", "[idée] "+req.Title)
	q.Set("body", ideaMarkdown(ideaReq{Title: req.Title, Body: body, Image: req.Image}, ""))
	return "https://github.com/" + updateRepo + "/issues/new?" + q.Encode()
}

func defaultOpenIdeaURL(raw string) {
	if testing.Testing() || strings.TrimSpace(raw) == "" {
		return
	}
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", raw)
	cmd.SysProcAttr = windowsHiddenProcAttr()
	_ = cmd.Start()
}

func sendIdeaMail(email string, req ideaReq) error {
	if len(req.Image) == 0 {
		subj := "[CD Scout] idée : " + req.Title
		if ideaUnsafe(req.Title, req.Body) {
			subj = "[CD Scout] proposition privée : " + req.Title
		}
		openIdeaURL("mailto:" + email + "?subject=" + url.QueryEscape(subj) + "&body=" + url.QueryEscape(req.Body))
		return nil
	}
	path, err := writeIdeaEML(email, req)
	if err != nil {
		return err
	}
	openIdeaURL(path)
	return nil
}

func sendIdeaSMS(phone string, req ideaReq) error {
	phone = sanitizeSMS(phone)
	if phone == "" {
		return errors.New("numéro SMS invalide")
	}
	text := "CD Scout : " + req.Title + "\n" + req.Body
	if r := []rune(text); len(r) > 280 {
		text = string(r[:280]) + "…"
	}
	openIdeaURL("sms:" + phone + "?body=" + url.QueryEscape(text))
	return nil
}

func writeIdeaEML(email string, req ideaReq) (string, error) {
	dir := filepath.Join(ideaDir(), "ideas")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	boundary := "cdscout" + fmt.Sprintf("%d", time.Now().UnixNano())
	var b strings.Builder
	if email != "" {
		b.WriteString("To: " + email + "\r\n")
	}
	subj := "[CD Scout] idée : " + req.Title
	if ideaUnsafe(req.Title, req.Body) {
		subj = "[CD Scout] proposition privée : " + req.Title
	}
	b.WriteString("Subject: " + rfc2047(subj) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: multipart/mixed; boundary=\"" + boundary + "\"\r\n\r\n")
	b.WriteString("--" + boundary + "\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	b.WriteString(wrapBase64(base64.StdEncoding.EncodeToString([]byte(req.Body))) + "\r\n")
	if len(req.Image) > 0 {
		name := "maquette." + req.ImageExt
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString("Content-Type: " + req.ImageMIME + "; name=\"" + name + "\"\r\n")
		b.WriteString("Content-Transfer-Encoding: base64\r\n")
		b.WriteString("Content-Disposition: attachment; filename=\"" + name + "\"\r\n\r\n")
		b.WriteString(wrapBase64(base64.StdEncoding.EncodeToString(req.Image)) + "\r\n")
	}
	b.WriteString("--" + boundary + "--\r\n")
	path := filepath.Join(dir, time.Now().Format("20060102-150405")+"-"+ideaSlug(req.Title)+".eml")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func rfc2047(s string) string {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(s)) + "?="
}

func wrapBase64(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i += 76 {
		end := i + 76
		if end > len(s) {
			end = len(s)
		}
		b.WriteString(s[i:end])
		b.WriteString("\r\n")
	}
	return strings.TrimRight(b.String(), "\r\n")
}
