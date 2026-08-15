# CD Scout

Companion LoL Windows : **scouting en champ select** + **HUD de CD en jeu**. Un binaire Go, une page `index.html` embarquée, overlay WebView2. Pas d’injection dans League.

```
League (LCU lockfile + Live Client :2999)
        │
        ▼
  Go  127.0.0.1:27182     ← Data Dragon, CommunityDragon, DeepLoL, wiki
        │
        ├── navigateur     index.html          (draft / live / quiz / idées)
        └── WebView2 HWND  index.html?hud=1    (overlay click-through)
```

Repo : [sologames7/lol-cd-scout-native](https://github.com/sologames7/lol-cd-scout-native). Contexte agent : `.cursor/rules/projet-cd-scout.mdc`.

---

## Démarrer

Prérequis : [Go](https://go.dev/dl/) ≥ 1.23, WebView2 (Edge, déjà sur Win11), League en **fenêtré sans bordure**.

```powershell
cd lol-cd-scout-native
$env:Path += ";C:\Program Files\Go\bin"

.\run.ps1          # quit + build + lance
go test .          # avant de pousser un fix
```

Build seul :

```powershell
try { Invoke-RestMethod -Method POST "http://127.0.0.1:27182/api/quit" | Out-Null } catch {}
go build -ldflags="-s -w -H=windowsgui" -o lol-cd-scout-native.exe .
Start-Process .\lol-cd-scout-native.exe
```

Toujours `/api/quit` avant de recompiler (sinon l’exe est verrouillé). `run.ps1` retombe sur `go run .` si Smart App Control bloque l’exe.

Tester **sans League** : boutons **Démo draft** / **Démo en jeu** dans l’UI. Tester le HUD live : Outil d’entraînement, overlay visible seulement en fenêtré sans bordure.

---

## Où toucher quoi

| Fichier | Rôle |
|---|---|
| `main.go` | HTTP, LCU, Data Dragon, `/api/*` |
| `index.html` | Tout le front (CSS+JS inline, 0 lib). Mode HUD = `?hud=1` |
| `overlay_windows.go` | HWND overlay, RGN click-through, WebView2 A:0 |
| `overlay_geom.go` | Positions défaut v5, clamp écran, Tab 1080p→scale |
| `clipboard_windows.go` | SUM / `/allsum` → presse-papier + SendInput League |
| `livegame.go` | Live Client Data, haste items, or estimé, objectifs 2026 |
| `voicecue.go` | Voix jungle (ganks) + cue achat item |
| `objsfx.go` | SFX + bannières PNG `/alerts/` |
| `profiles.go` / `deeplol.go` | Pseudo, rang, WR, AI-score, tags live |
| `draft.go` | Conseils pick/ban |
| `curated.go` | Fiches FR (priorités sorts, fenêtres) |
| `matchups.go` / `synergies.go` / `roles.go` | Données générées (ne pas éditer à la main) |
| `passivecd.go` | CD des passifs (Meraki → texte → curated) |
| `itemmeta.go` | Items « finis » (nom, or, légendaire) |
| `tracks.go` | Timers CD partagés entre fenêtres |
| `quiz.go` / `idea.go` / `update.go` | Quiz, issues GitHub, auto-update |
| `huddev.go` | Overlay hors partie (`CDSCOUT_DEV=1` ou build `dev`) |
| `run.ps1` | Relance quotidienne |

Stubs `*_other.go` = no-op hors Windows. Tests = `*_test.go` à côté.

**HUD (règle d’or)** : click-through = `SetWindowRgn` parent **et** enfants WebView2. Jamais toggler `WS_EX_TRANSPARENT` / `WS_EX_LAYERED` au hover. Jamais d’overlay Chromium / Electron / injection.

---

## Produit

| Mode | Quand | Quoi |
|---|---|---|
| Draft | Champ select LCU | Cartes ennemis/alliés, counters, synergies, pick/ban |
| Live (fenêtre) | Partie | Flash, P/Q/W/E/R, invocs, or, objectifs, DeepLoL |
| HUD overlay | Partie (`?hud=1`) | Mini-barre ⠿ ⤡ **SUM** ⚙ + chips Flash ; Tab = cartes latérales ; bannières obj / gank / achat |

HUD clavier : `Alt+1–5` Flash, `Alt+Maj+1–5` ult, Tab cartes. Croix = ferme le widget (`/api/hud/close`), pas l’app.

Hors overlay : Quiz CD, Moments forts (YouTube), S’entraîner (Skill Capped), Idée → issue GitHub.

---

## HTTP local (`127.0.0.1:27182`)

| Route | Usage |
|---|---|
| `GET /` `GET /hud` | UI / overlay |
| `GET /api/status` | Phase LCU + draft + version app |
| `GET /api/live` | Snapshot in-game (poll 2 s) |
| `GET /api/cards` `GET /api/card` `GET /api/search` | Fiches champions |
| `GET /api/draft` | Conseils pick/ban |
| `POST /api/quit` | Arrête l’app (avant rebuild) |
| `POST /api/clipboard?erase=` | SUM / `/allsum` |
| `GET\|POST /api/hud/*` | open / close / hits / bounds / drag / geom / reset |
| `GET /api/voice?key=` `GET /api/objsfx?k=` | Audio proxifié |
| `GET\|POST /api/quiz` `POST /api/idea` | Quiz / idées |

`/api/open` n’accepte que des URLs `https://www.deeplol.gg/…`.

---

## Données locales

`%LOCALAPPDATA%\lol-cd-scout\`

| Fichier | |
|---|---|
| `hud-geom.json` | Positions widgets (v5). Reset = ⚙ Overlay, pas un rm manuel |
| `hud-webview2\` | Profil WebView2 |
| `devmode` | Overlay forcé (builds `dev` seulement) |
| `ideas\` `github.token` `notify.json` | Formulaire Idée |

Front : `localStorage` `cdscout.hudwopacity` (opacité par widget, 25–100 %).

---

## Pièges

- UI, commentaires, réponses : **français**.
- Joueur = `pid = riotId + '#' + championKey` — jamais riotId seul (bots, anonymat ranked).
- Timers côté client, horloge = `gameTime` interpolé (`gameNow()`), ticker 300 ms.
- Overlay **invisible** en plein écran exclusif League.
- Or estimé, ordre de skill Q/W/E estimé, runes = keystone + secondaire seulement.
- Variables PowerShell ne persistent pas ; PATH peut perdre Go.
- `appVersion` défaut = `dev` → pas d’auto-update (`run.ps1` n’est jamais impacté).

---

## Données générées

```powershell
go run gensynergies.go    # → synergies.go + roles.go (LoLTheory / Meraki)
```

`matchups.go` = counters LoLalytics (auto-généré). Une fiche `curated.go` avec counters non vides **remplace** ces listes.

Bannières d’alerte : `branding/alerts/compose.ps1` → PNG servies en `/alerts/`.

---

## Release

```powershell
git tag vX.Y.Z
git push origin vX.Y.Z
```

CI (`.github/workflows/release.yml`) build Windows avec `-X main.appVersion=vX.Y.Z` et attache l’exe. Les clients à jour le récupèrent au lancement suivant.

L’exe doit vivre dans un dossier **inscriptible** (pas Program Files). Binaire non signé → SmartScreen / Smart App Control possibles.
