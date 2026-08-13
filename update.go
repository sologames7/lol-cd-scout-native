package main

// Auto-update depuis les GitHub Releases du repo. Modèle "Chrome" : la
// vérification et le téléchargement se font en tâche de fond, la nouvelle
// version est déposée à côté de l'exe et installée au lancement suivant (ou
// tout de suite via /api/update). Rien ne retarde jamais le démarrage.

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// appVersion est injectée au build : -ldflags "-X main.appVersion=v1.0.0".
// La valeur "dev" désactive l'auto-update (builds locaux via run.ps1).
var appVersion = "dev"

const (
	updateRepo  = "sologames7/lol-cd-scout-native"
	updateAsset = "lol-cd-scout-native.exe"
)

// Client dédié : les ~7 Mo du binaire dépassent le timeout de httpClient.
var updateHTTP = &http.Client{Timeout: 5 * time.Minute}

type UpdateInfo struct {
	Current string `json:"current"`
	Latest  string `json:"latest,omitempty"`
	State   string `json:"state"` // idle | checking | downloading | ready | error
	Percent int    `json:"percent"`
	Error   string `json:"error,omitempty"`
}

var (
	updateMu    sync.Mutex
	updateState = UpdateInfo{State: "idle"}
)

func updateSnapshot() UpdateInfo {
	updateMu.Lock()
	defer updateMu.Unlock()
	u := updateState
	u.Current = appVersion
	return u
}

func setUpdate(f func(*UpdateInfo)) {
	updateMu.Lock()
	defer updateMu.Unlock()
	f(&updateState)
}

func exePath() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", err
	}
	if r, err := filepath.EvalSymlinks(p); err == nil {
		p = r
	}
	return p, nil
}

// applyStagedUpdate installe la version téléchargée lors d'une session
// précédente puis relance. Windows interdit d'écraser un exe en cours
// d'exécution mais autorise son renommage, d'où la valse .old / .new.
func applyStagedUpdate() bool {
	exe, err := exePath()
	if err != nil {
		return false
	}
	_ = os.Remove(exe + ".old")
	staged := exe + ".new"
	if !looksLikeExe(staged) {
		_ = os.Remove(staged)
		return false
	}
	if err := os.Rename(exe, exe+".old"); err != nil {
		return false
	}
	if err := os.Rename(staged, exe); err != nil {
		_ = os.Rename(exe+".old", exe)
		return false
	}
	return relaunch(exe)
}

func looksLikeExe(p string) bool {
	f, err := os.Open(p)
	if err != nil {
		return false
	}
	defer f.Close()
	if st, err := f.Stat(); err != nil || st.Size() < 1<<20 {
		return false
	}
	var magic [2]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil {
		return false
	}
	return magic[0] == 'M' && magic[1] == 'Z'
}

func relaunch(exe string) bool {
	cmd := exec.Command(exe)
	cmd.Dir = filepath.Dir(exe)
	cmd.SysProcAttr = windowsHiddenProcAttr()
	if err := cmd.Start(); err != nil {
		return false
	}
	_ = cmd.Process.Release()
	return true
}

type ghAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
	Size int64  `json:"size"`
}

type ghRelease struct {
	Tag        string    `json:"tag_name"`
	Draft      bool      `json:"draft"`
	Prerelease bool      `json:"prerelease"`
	Assets     []ghAsset `json:"assets"`
}

// checkAndStage interroge GitHub et dépose la nouvelle version à côté de l'exe.
func checkAndStage() {
	if appVersion == "dev" || appVersion == "" {
		return
	}
	exe, err := exePath()
	if err != nil {
		return
	}
	setUpdate(func(u *UpdateInfo) { u.State = "checking"; u.Error = "" })

	var rel ghRelease
	url := "https://api.github.com/repos/" + updateRepo + "/releases/latest"
	if err := jsonGET(url, &rel); err != nil {
		setUpdate(func(u *UpdateInfo) { u.State = "idle"; u.Error = "vérification impossible" })
		return
	}
	if rel.Draft || rel.Prerelease || !isNewer(rel.Tag, appVersion) {
		setUpdate(func(u *UpdateInfo) { u.State = "idle"; u.Latest = rel.Tag })
		return
	}
	var asset *ghAsset
	for i := range rel.Assets {
		if strings.EqualFold(rel.Assets[i].Name, updateAsset) {
			asset = &rel.Assets[i]
			break
		}
	}
	if asset == nil {
		setUpdate(func(u *UpdateInfo) { u.State = "error"; u.Error = "binaire absent de la release" })
		return
	}

	setUpdate(func(u *UpdateInfo) { u.State = "downloading"; u.Latest = rel.Tag; u.Percent = 0 })
	staged := exe + ".new"
	if err := downloadTo(asset.URL, staged, asset.Size); err != nil {
		_ = os.Remove(staged)
		setUpdate(func(u *UpdateInfo) { u.State = "error"; u.Error = err.Error() })
		return
	}
	setUpdate(func(u *UpdateInfo) { u.State = "ready"; u.Percent = 100 })
}

func downloadTo(url, dst string, size int64) error {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "LoL-CD-Scout/"+appVersion)
	res, err := updateHTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", res.StatusCode)
	}
	// Temporaire dans le même dossier : le renommage final y est atomique.
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".cdscout-upd-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	written, err := io.Copy(tmp, &progressReader{r: res.Body, total: size})
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil && size > 0 && written != size {
		err = fmt.Errorf("téléchargement incomplet (%d/%d octets)", written, size)
	}
	if err == nil && !looksLikeExe(tmpName) {
		err = fmt.Errorf("fichier téléchargé invalide")
	}
	if err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	_ = os.Remove(dst)
	if err := os.Rename(tmpName, dst); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

type progressReader struct {
	r     io.Reader
	total int64
	done  int64
	last  int
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.done += int64(n)
	if p.total > 0 {
		if pct := int(p.done * 100 / p.total); pct != p.last {
			p.last = pct
			setUpdate(func(u *UpdateInfo) { u.Percent = pct })
		}
	}
	return n, err
}

// isNewer compare deux tags de type "v1.4.0" composant par composant.
func isNewer(latest, current string) bool {
	a, b := splitVersion(latest), splitVersion(current)
	if len(a) == 0 {
		return false
	}
	for i := 0; i < len(a) || i < len(b); i++ {
		var x, y int
		if i < len(a) {
			x = a[i]
		}
		if i < len(b) {
			y = b[i]
		}
		if x != y {
			return x > y
		}
	}
	return false
}

func splitVersion(s string) []int {
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	out := []int{}
	for _, p := range strings.Split(s, ".") {
		n, err := strconv.Atoi(p)
		if err != nil {
			return out
		}
		out = append(out, n)
	}
	return out
}

// listenLocal réessaie brièvement : après un redémarrage pour mise à jour,
// l'ancien process peut encore tenir le port quelques centaines de ms.
func listenLocal() (net.Listener, error) {
	var err error
	for i := 0; i < 8; i++ {
		var ln net.Listener
		if ln, err = net.Listen("tcp", listenAddr); err == nil {
			return ln, nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return net.Listen("tcp", "127.0.0.1:0")
}

func apiUpdate(w http.ResponseWriter, r *http.Request) {
	if updateSnapshot().State != "ready" {
		http.Error(w, "aucune mise à jour prête", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	go func() {
		// Le listener n'est pas fermé ici : http.Serve rendrait la main, main()
		// retournerait et le process mourrait avant le remplacement. La nouvelle
		// instance patiente sur le port le temps que celle-ci sorte (listenLocal).
		time.Sleep(200 * time.Millisecond)
		if !applyStagedUpdate() {
			// L'exe est resté intact : on relance la version actuelle.
			if exe, err := exePath(); err == nil {
				relaunch(exe)
			}
		}
		os.Exit(0)
	}()
}
