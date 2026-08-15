# Brief agent — suite HUD CD Scout

Pars de **ce fichier**. Workspace : `lol-cd-scout-native`. UI en français. Stack : binaire Go + `index.html` embarqué + WebView2 HWND. **Ne change pas de stack.** Ne pas injecter dans League. Ne pas toggler `WS_EX_TRANSPARENT` / `WS_EX_LAYERED` au runtime.

Lis : `.cursor/rules/projet-cd-scout.mdc`, `overlay_windows.go`, `overlay_geom.go`, `index.html` (mode `?hud=1`).

---

## Contexte utilisateur (dernier prompt)

> Reprends le HUD CD Scout. Lis `.cursor/BRIEF-HUD-NEXT.md`. Ne change pas de stack. Ne développe pas Studio / Figer. Vérifie overlay dans l’écran, SUM discret, pop bannières, ⚙ Notifs gank jgl. Corrige seulement ce qui casse encore. go test . puis test fenêtré sans bordure.

Donc : **pas de nouvelles features Studio / Figer**. Recaler le HUD, puis s’arrêter.

---

## Déjà fait (ne pas refaire)

### Click-through + opacité
- `SetWindowRgn` parent **et** enfants WebView2 (`Chrome_WidgetWin_*`, `Intermediate D3D Window`).
- Hits JS serrés (`reportHudHits`) : mini-barre, flashes, wait, vis-menu ouvert, objs, rows Tab, alert/item visibles, grips au hover. **Pas** le bbox du widget menu entier.
- Fond WebView2 `A:0`, CSS hudmode transparent. Slider opacité 25–100 % (`cdscout.hudopacity`). Menu ⚙ lisible via `hud-vis-open`.
- Interdit : toggle LAYERED/TRANSPARENT au hover, recréer WebView2, overlay DX.

### Barre permanente + SUM
- Horloge retirée. Nav : ⠿ · ⤡ · **SUM** · point d’état · ⚙. ♪ / RST / ✕ dans ⚙ (Audio / Overlay).
- SUM : Entrée puis clic → colle les invocs ennemis encore utiles (sums > 10 s, ult > 5 s exclus). `/allsum` dans le chat League → 7 backspaces puis sums + R (`watchAllsum` / `writeClipboard`).
- SUM **adouci** (orange lisible, pas un pavé `#ff6a00`) : fond `rgba(255,122,40,.22)`, texte `#ffc090`. Vérifié : inchangé cette passe.

### Cartes Tab
- Timer **sur** l’icône. Sous l’icône : `Q 70` violet (`spellCostLabel` / `s.cost`).

### Position défaut + anims
- Géométrie **v5** (`hudGeomVer`). Menu **520 px** (`hudMenuW` / `HUD_MENU_W`) : les 5 chips Flash font ~491 px ; 420 laissait ~70 px hors écran à droite (1080p et 2560). Merge v<5 **recalcule** menu / alert / item. Clamp Go `clampHudWidgetID` avec cette estimée (un vieux défaut `sw-420-16` est recalé sans bump de version).
- Clamp JS `keepHudWidgetsOnScreen()` : `offsetWidth * scale` (pas `getBoundingClientRect` pendant le pop 1.14). Aussi après `fireAlert` / `showItemPop`. Pas pendant grip.
- Cartes alliés Tab : `rightX + tabCardW` ne dépasse plus `sw`.
- Anims bannière / achat : pop « bulle » (`hudPopIn` / `hudPopOut`, origin `right center`). Pas de `translateX(100vw)`.

### Réglages ⚙
- Widgets, Alertes (**Notifs gank jgl** = `hudOpt.ganks` : bannière texte gank seulement ; **ne coupe plus** `enqueueVoice`), Audio + volumes, **opacité par widget** (`cdscout.hudwopacity`, 25–100 %), Overlay (reset / quitter), Studio seulement si `force` UI.
- Couper ganks ne cache plus une bannière PNG d’objectif (`has-banner`).
- **Voix jungle** (`hudOpt.voices`) = ogg, séparé des notifs visuelles.

---

## À ne pas faire

- Pas de Studio / Figer / preview bannière supplémentaire.
- Pas de rewrite Electron / overlay injecté / toggle styles HWND au hover.
- Ne pas toucher alertes PNG art, LCU, quiz, livegame haste, sauf si un bug clamp/anim l’exige.
- Ne pas resetter `hud-geom.json` de l’utilisateur sans demande (layout custom 2560×1440 déjà en place).

---

## Prochaines étapes (si l’utilisateur relance)

Ordre suggéré, **une chose à la fois**, tester en **fenêtré sans bordure** :

1. **Vérif manuelle** (cette passe : `go test .` OK, overlay relancé, HUD `window: true` pendant InProgress 2560×1440) : menu 5 chips entièrement dans l’écran ; SUM orange discret ; bannière obj / gank / achat en pop ; ⚙ → Alertes → **Notifs gank jgl** coupe le texte, **Voix jungle** reste à part.
2. **Reset dispo** seulement si un widget custom est encore coupé (⚙ Overlay).
3. **Studio Figer** : seulement si l’utilisateur le redemande.
4. Bugs éventuels à surveiller : hits trop larges pendant le pop (overshoot scale 1.14) ; menu ⚙ trop long sur petit écran ; SUM qui n’envoie pas si le chat League n’est pas ouvert ; `tabE` custom collé à droite sur 2560 (choix user, pas un défaut).

---

## Fichiers clés

| Fichier | Rôle |
|---|---|
| `index.html` | HUD JS/CSS, SUM, clamp écran, pop bannière, ⚙, Tab |
| `overlay_geom.go` | défauts v5, `hudMenuW=520`, clamp par widget, merge |
| `overlay_windows.go` | HWND, RGN enfant, `/allsum`, WebView2 A:0 |
| `clipboard_windows.go` | `writeClipboard(s, erase)` |
| `main.go` | `POST /api/clipboard?erase=` |

Tests : `go test .` (PATH Go : `C:\Program Files\Go\bin`). Relance : `.\run.ps1`. Overlay visible seulement en **fenêtré sans bordure**.

---

## Prompt de départ (copier-coller)

> Reprends le HUD CD Scout (`lol-cd-scout-native`). Lis `.cursor/BRIEF-HUD-NEXT.md`. Ne change pas de stack. Ne développe pas Studio / Figer.
>
> Vérifie d’abord que l’overlay au lancement ne dépasse plus l’écran, que SUM n’est plus un pavé fluo, que les bannières poppent sur place, et que ⚙ Alertes a **Notifs gank jgl** (voix jungle à part). Corrige uniquement ce qui casse encore parmi ça.
>
> Fichiers : `index.html`, `overlay_geom.go` / `overlay_windows.go` si clamp HWND. `go test .` puis test manuel fenêtré sans bordure.
