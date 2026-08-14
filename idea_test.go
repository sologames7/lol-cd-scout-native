package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var ideaPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89,
	0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41, 0x54,
	0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00, 0x05, 0x00, 0x01,
	0x0d, 0x0a, 0x2d, 0xb4,
	0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}

func stubIdea(t *testing.T, token string, post func(string, ideaReq) (string, error), open func(string)) {
	t.Helper()
	oldTok, oldPost, oldOpen, oldDir := lookupIdeaToken, postIdeaIssue, openIdeaURL, ideaDirOverride
	ideaMu.Lock()
	oldLast := ideaLastAt
	ideaLastAt = time.Time{}
	ideaMu.Unlock()
	t.Cleanup(func() {
		lookupIdeaToken, postIdeaIssue, openIdeaURL, ideaDirOverride = oldTok, oldPost, oldOpen, oldDir
		ideaMu.Lock()
		ideaLastAt = oldLast
		ideaMu.Unlock()
	})
	lookupIdeaToken = func() string { return token }
	if post != nil {
		postIdeaIssue = post
	} else {
		postIdeaIssue = func(string, ideaReq) (string, error) { return "", errors.New("no github") }
	}
	if open != nil {
		openIdeaURL = open
	} else {
		openIdeaURL = func(string) {}
	}
	ideaDirOverride = t.TempDir()
}

func TestIdeaUnsafe(t *testing.T) {
	safe := [][2]string{
		{"Timer Nashor", "Afficher le timer Baron sur le HUD comme le Flash."},
		{"Fenêtre de punish", "Alerte quand l'ulti ennemi est down pour exploiter la fenêtre."},
		{"Overlay click-through", "Garder les clics dans League hors des blocs."},
	}
	for _, c := range safe {
		if ideaUnsafe(c[0], c[1]) {
			t.Fatalf("faux positif: %s", c[0])
		}
	}
	unsafe := [][2]string{
		{"Keylogger", "Un enregistreur de frappe pour le chat adverse."},
		{"Inject DLL", "Injecter une DLL dans League of Legends."},
		{"Steal", "Voler le mot de passe du compte Riot."},
		{"Cheat Engine", "Brancher Cheat Engine sur le client."},
	}
	for _, c := range unsafe {
		if !ideaUnsafe(c[0], c[1]) {
			t.Fatalf("manqué: %s", c[0])
		}
	}
}

func TestSniffIdeaImage(t *testing.T) {
	ext, mime, ok := sniffIdeaImage(ideaPNG, "x.png")
	if !ok || ext != "png" || mime != "image/png" {
		t.Fatalf("png %s %s %v", ext, mime, ok)
	}
	if _, _, ok := sniffIdeaImage([]byte("MZ\x90\x00not-an-image-file"), "x.png"); ok {
		t.Fatal("exe acceptée")
	}
	if _, _, ok := sniffIdeaImage([]byte("<svg xmlns='x'></svg>"), "x.svg"); ok {
		t.Fatal("svg acceptée")
	}
}

func TestValidateIdea(t *testing.T) {
	if validateIdea(ideaReq{Title: "ab", Body: "assez long texte"}) == nil {
		t.Fatal("titre court")
	}
	if validateIdea(ideaReq{Title: "Timer Baron", Body: "court"}) == nil {
		t.Fatal("body court")
	}
	if err := validateIdea(ideaReq{Title: "Timer Baron", Body: "Afficher le timer Nashor sur le HUD."}); err != nil {
		t.Fatal(err)
	}
}

func TestSubmitIdeaGitHub(t *testing.T) {
	var got ideaReq
	stubIdea(t, "tok", func(_ string, req ideaReq) (string, error) {
		got = req
		return "https://github.com/sologames7/lol-cd-scout-native/issues/12", nil
	}, nil)
	out := submitIdea(ideaReq{Title: "Timer Baron", Body: "Afficher le timer Nashor sur le HUD."})
	if !out.OK || out.Via != "github" || !strings.Contains(out.URL, "/issues/12") {
		t.Fatalf("%+v", out)
	}
	if got.Title != "Timer Baron" {
		t.Fatalf("titre %q", got.Title)
	}
}

func TestSubmitIdeaFormSansToken(t *testing.T) {
	var opened string
	stubIdea(t, "", nil, func(u string) { opened = u })
	out := submitIdea(ideaReq{Title: "Timer Baron", Body: "Afficher le timer Nashor sur le HUD."})
	if !out.OK || out.Via != "form" {
		t.Fatalf("%+v", out)
	}
	if !strings.Contains(opened, "github.com/"+updateRepo+"/issues/new") {
		t.Fatalf("url %s", opened)
	}
}

func TestSubmitIdeaUnsafeMail(t *testing.T) {
	var opened string
	stubIdea(t, "tok", func(string, ideaReq) (string, error) {
		t.Fatal("GitHub ne doit pas être appelé")
		return "", nil
	}, func(u string) { opened = u })
	t.Setenv("CDSCOUT_NOTIFY_EMAIL", "dev@example.com")
	t.Setenv("CDSCOUT_NOTIFY_SMS", "")
	out := submitIdea(ideaReq{Title: "Keylogger chat", Body: "Un keylogger pour lire le chat ennemi en live."})
	if !out.OK || !out.Unsafe || out.Via != "mail" {
		t.Fatalf("%+v", out)
	}
	if !strings.HasPrefix(opened, "mailto:dev@example.com") {
		t.Fatalf("mailto %s", opened)
	}
}

func TestSubmitIdeaUnsafeMailAvecImage(t *testing.T) {
	var opened string
	stubIdea(t, "", nil, func(u string) { opened = u })
	t.Setenv("CDSCOUT_NOTIFY_EMAIL", "dev@example.com")
	t.Setenv("CDSCOUT_NOTIFY_SMS", "+33600000000")
	out := submitIdea(ideaReq{
		Title: "Keylogger chat", Body: "Un keylogger pour lire le chat ennemi en live.",
		Image: ideaPNG, ImageExt: "png", ImageMIME: "image/png",
	})
	if !out.OK || out.Via != "mail" || !strings.HasSuffix(opened, ".eml") {
		t.Fatalf("%+v opened=%s", out, opened)
	}
}

func TestSubmitIdeaUnsafeSMSSansImage(t *testing.T) {
	var opened string
	stubIdea(t, "", nil, func(u string) { opened = u })
	t.Setenv("CDSCOUT_NOTIFY_EMAIL", "")
	t.Setenv("CDSCOUT_NOTIFY_SMS", "+33 6 12 34 56 78")
	out := submitIdea(ideaReq{Title: "Cheat Engine", Body: "Brancher Cheat Engine sur le client live."})
	if !out.OK || out.Via != "sms" || !out.Unsafe {
		t.Fatalf("%+v", out)
	}
	if !strings.HasPrefix(opened, "sms:+33612345678") {
		t.Fatalf("sms %s", opened)
	}
}

func TestPostGitHubIssueAt(t *testing.T) {
	var created bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/issues") {
			created = true
			if got := r.Header.Get("Authorization"); got != "Bearer tok" {
				t.Errorf("auth %s", got)
			}
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"html_url":"https://github.com/sologames7/lol-cd-scout-native/issues/7","number":7,"id":99}`)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	u, err := postGitHubIssueAt(srv.URL, "tok", ideaReq{Title: "Timer Baron", Body: "Afficher le timer Nashor sur le HUD."})
	if err != nil || !strings.HasSuffix(u, "/issues/7") || !created {
		t.Fatalf("u=%s err=%v created=%v", u, err, created)
	}
}

func TestApiIdeaMultipart(t *testing.T) {
	var got ideaReq
	stubIdea(t, "tok", func(_ string, req ideaReq) (string, error) {
		got = req
		return "https://github.com/sologames7/lol-cd-scout-native/issues/3", nil
	}, nil)
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("title", "Bannières drake")
	_ = mw.WriteField("body", "Une bannière par type de dragon, déjà presque là.")
	part, err := mw.CreateFormFile("image", "drake.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(ideaPNG); err != nil {
		t.Fatal(err)
	}
	mw.Close()
	r := httptest.NewRequest(http.MethodPost, "/api/idea", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	apiIdea(w, r)
	if w.Code != 200 {
		t.Fatalf("code %d %s", w.Code, w.Body.Bytes())
	}
	var out ideaOut
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || !out.OK || out.Via != "github" {
		t.Fatalf("%s", w.Body.Bytes())
	}
	if got.ImageExt != "png" || len(got.Image) < 20 {
		t.Fatalf("image %+v %d", got.ImageExt, len(got.Image))
	}
	matches, _ := filepath.Glob(filepath.Join(ideaDirOverride, "ideas", "*.json"))
	if len(matches) != 1 {
		t.Fatalf("archive %v", matches)
	}
}

func TestApiIdeaRejectsShort(t *testing.T) {
	stubIdea(t, "", nil, nil)
	r := httptest.NewRequest(http.MethodPost, "/api/idea", strings.NewReader(`{"title":"x","body":"trop court"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	apiIdea(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code %d", w.Code)
	}
}

func TestGithubNewIssueURL(t *testing.T) {
	u := githubNewIssueURL(ideaReq{Title: "Timer Baron", Body: "Afficher le timer Nashor sur le HUD."})
	if !strings.Contains(u, "/issues/new?") || !strings.Contains(u, "title=") {
		t.Fatal(u)
	}
}
