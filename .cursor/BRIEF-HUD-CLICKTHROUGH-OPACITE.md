# Brief agent — click-through propre + opacité HUD

Tu implémentes **deux features** sur l’overlay Windows actuel. Le câblage HUD existe déjà. Ce n’est **pas** un rewrite, **pas** un changement de stack, **pas** un overlay injecté dans League.

Lis d’abord : `overlay_windows.go`, `overlay_other.go`, `overlay_geom.go`, le mode `?hud=1` dans `index.html` (menu 👁 `#hud-vis`), `.cursor/rules/projet-cd-scout.mdc` § overlay.

## Pourquoi les versions précédentes ont « presque » marché

Ce n’est **pas** une limite dure de WebView2. C’est une limite de **comment** on l’embarque, et les tentatives ont attaqué le mauvais HWND.

Architecture actuelle :

- 1 HWND parent plein écran `WS_POPUP` (`CdScoutHudWidget`), `WS_EX_TOPMOST | TOOLWINDOW | NOACTIVATE`.
- WebView2 **HWND embedding** (`cr.Embed`) : Chromium crée un **HWND enfant** (`Chrome_WidgetWin_*` / `Intermediate D3D Window`) qui **couvre tout le client**.
- Click-through aujourd’hui : `SetWindowRgn` + `WM_NCHITTEST` → `HTTRANSPARENT` d’après `/api/hud/hits`.
- Fond WebView2 **opaque** `A:255 #121c2c` (`hudSetControllerOpaque`).
- Interdit jusqu’ici : toggler `WS_EX_TRANSPARENT` / `WS_EX_LAYERED` **au survol** → rectangle gris (recréation de la surface DWM).

### Le bug réel

`WM_NCHITTEST` et `SetWindowRgn` sur le **parent** ne suffisent pas.

Windows hit-teste d’abord l’enfant sous le curseur. L’enfant WebView2 dit « c’est chez moi » sur **tout** le rectangle plein écran. League ne reçoit jamais le clic, même dans les « trous » CSS. `SetWindowRgn` du parent clippe surtout le **paint** ; le mouse routing de l’enfant reste.

`hudHits.solid` est écrit (`/api/hud/solid`) mais **jamais lu** dans `applyHudRegion`. Code mort.

Les hits JS sont des **bounding boxes** de `[data-hud-hit]`. `#hud-w-menu` a `data-hud-hit` sur **tout le widget** → un gros rectangle mange les clics autour des chips Flash (coins arrondis, gouttières, espace vide).

### Ce que l’utilisateur veut (produit)

1. **Click-through** : hors des vrais blocs UI (boutons, chips, cartes, grips, menu 👁 ouvert), le clic / drag / molette vont **uniquement** dans League. Sur un bloc, le HUD reçoit le clic (Flash, sorts, LOCK, slider). **Pas** de double livraison (HUD + jeu sur le même clic — ça ferait bouger le champion en cliquant un Flash).
2. **Opacité** : slider dans les **paramètres overlay** (menu 👁 déjà là) qui rend **tout** l’overlay plus transparent (widgets compris). Persisté. Défaut ~100 %. Plancher ~25 % (lisible). Pas un fond navy plein écran à 40 %.

League **fenêtré sans bordure** uniquement (déjà documenté). Plein écran exclusif = hors scope.

---

## Verdict techno — rester sur Go + WebView2

| Approche | Verdict |
|---|---|
| Stack actuelle + région **aussi sur l’enfant** WebView2 | **À faire.** Même techno, le trou manquant. |
| Fond WebView2 `A:0` + CSS transparent hudmode + slider CSS | **À faire** pour l’opacité. |
| `WS_EX_LAYERED` **à la création** + `SetLayeredWindowAttributes` (alpha) | Option B opacité native si CSS ne fade pas assez. **Jamais toggler** le style ensuite. |
| Toggler `WS_EX_TRANSPARENT` au hover | **Interdit.** Rectangle gris. |
| Toggler `WS_EX_LAYERED` au runtime | **Interdit.** Même famille de bug DWM. |
| `ICoreWebView2CompositionController` + DirectComposition | Rewrite. Seulement si A+B échouent après test en partie réelle. |
| 1 HWND / widget, ou overlay injecté (DX) | Hors scope. Anti-cheat / complexité. |
| Electron / Overwolf / CEF | Hors scope. |

**Réponse à « c’est la techno qui limite ? »** : HWND-embed WebView2 **peut** faire click-through + semi-transparence. Ce qui bloque aujourd’hui : (1) région/hit-test pas appliqués à l’enfant Chromium, (2) fond contrôleur opaque, (3) hits trop larges. Blitz/Overwolf font la même chose (shape + ignore mouse hors widgets). On n’a pas besoin de changer de moteur.

---

## Solution A (obligatoire) — click-through

### 1. Appliquer la région à l’enfant WebView2

Après `cr.Embed` + `Resize`, trouver le HWND enfant (EnumChildWindows depuis le parent HUD, classes `Chrome_WidgetWin_0/1`, `Intermediate D3D Window`, ou le HWND du controller si go-webview2 l’expose).

`applyHudRegion(hwnd)` doit `SetWindowRgn` sur :

1. le parent HUD
2. **chaque** enfant WebView2 pertinent (copie de la même HRGN ; `SetWindowRgn` prend ownership → `CreateRectRgn` / `CombineRgn` **deux fois**, ne pas réutiliser le même handle)

Si 0 hits : région 1×1 (déjà le cas) **aussi** sur l’enfant, sinon l’enfant reste un mur plein écran.

Appeler `applyHudRegion` après chaque `cr.Resize()` (wmSize, NavigationCompleted, bounds). L’enfant est parfois recréé au resize.

### 2. Ne plus se fier au seul `WM_NCHITTEST` parent

Le garder (pas de mal). La région enfant est la source de vérité souris.

Option de renfort (si la région enfant flicker au resize) : poll 16 ms `GetCursorPos` → si hors hits, `WS_EX_TRANSPARENT` **uniquement sur l’enfant**, jamais sur le parent. Pas de toggle LAYERED. Documenter dans un commentaire pourquoi.

### 3. Hits JS plus serrés

`reportHudHits()` :

- Ne **pas** poster le bbox du widget menu entier. Poster les blocs visibles : `#hud-mini-bar`, `#hud-flashes`, `#hud-wait`, `#hud-vis-menu` s’il est ouvert, `#hud-objs`, rows Tab `[data-hud-hit]`, alert/item visibles, `#sp-tip`, grips **seulement** si `opacity > 0` et `pointer-events: auto` (hover).
- Arrondir vers l’extérieur de 1–2 px max (pas 20).
- Debounce / skip si identique (Go a déjà `hudHitsEqual`).
- Après toggle visibilité / ouverture menu 👁 / alert in-out : `reportHudHits()` immédiat (déjà en partie).

Pendant un drag widget (`hud-grip-on`) : élargir un peu le hit du widget déplacé pour ne pas perdre la souris.

### 4. Nettoyer `solid`

Soit `hudHits.solid` force une région = tout l’écran (debug), soit tu supprimes `/api/hud/solid` + `hudSetSolid` des deux OS. Pas d’API zombie.

### 5. Molette / clic droit

Les trous doivent laisser passer **tout** l’input (LButton, RButton, wheel, hover targeting League). Ne pas `SetCapture` hors drag HUD.

---

## Solution B (obligatoire) — opacité

### Principe

Le navy plein écran opaque **empêche** toute vraie transparence. Il faut un contrôleur WebView2 **transparent**, un CSS hudmode **transparent**, et des widgets qui portent **leur** fond.

### 1. Contrôleur + CSS

- Remplacer `hudSetControllerOpaque` par fond `A:0` (garder RGB 18,28,44 au cas où A=0 est ignoré). Appeler après Embed **et** NavigationCompleted (déjà le pattern).
- `html.hudmode, body.hudmode` : `background: transparent !important` (aujourd’hui `#121c2c !important` — c’est le slab).
- `#hud` : `background: transparent`.
- Les `.hud-block` / `.hf` / `.hr` **gardent** leurs fonds. Sans ça les textes flottent illisibles.
- `WM_ERASEBKGND` return 1 + `nullBrush` : **conserver**.

Vérifier visuellement : hors widgets, on voit **la map League**, pas un voile navy.

### 2. Slider dans le menu 👁

Menu existant `#hud-vis-menu` / `renderHudVisMenu()`. Ajouter une ligne **sous** les yeux :

- Label `Opacité`
- `<input type="range">` 25–100, pas 5, défaut 100
- Valeur affichée `85 %`
- `input` en live (pas seulement `change`)

Persistance : `localStorage` clé `cdscout.hudopacity` (0.25–1). Même profil WebView2 HUD.

Application : `document.documentElement.style.setProperty('--hud-opacity', v)` et

```css
html.hudmode #hud { opacity: var(--hud-opacity, 1); }
```

Le menu 👁 lui-même est dans `#hud` : il fade aussi, c’est voulu (« overlay entierement semi transparent »). Plancher 25 % pour rester réglable.

Si le slider est trop dur à viser une fois fade : **ne pas** mettre opacity sur `#hud-vis-menu` (exception CSS).

### 3. Si CSS opacity ne fade pas (swapchain opaque)

Alors seulement :

- `WS_EX_LAYERED` **dans `CreateWindowEx`**, jamais ensuite via `SetWindowLong`.
- `SetLayeredWindowAttributes(hwnd, 0, alphaByte, LWA_ALPHA)` quand le slider bouge.
- API `POST /api/hud/opacity` `{alpha: 0.25–1}` → Go applique sur le parent.
- Stub no-op dans `overlay_other.go`.

Ne pas combiner LWA_COLORKEY (trous déjà faits par RGN). Ne pas `UpdateLayeredWindow` (incompatible WebView2 HWND).

Préférer CSS si le fond A=0 marche : zéro round-trip, pas de LAYERED.

---

## Fichiers touchés (attendu)

- `overlay_windows.go` — région enfant, fond A=0, éventuellement LAYERED create-time + opacity API
- `overlay_other.go` — stubs
- `index.html` — hits serrés, CSS transparent hudmode, `--hud-opacity`, slider menu 👁
- `main.go` — seulement si API opacity / suppression `solid`
- `overlay_geom_test.go` — hits ; pas de test WinAPI ici
- `.cursor/rules/projet-cd-scout.mdc` — 2 phrases : click-through = RGN parent **et** enfant WebView2 ; opacité slider 👁 + fond A=0. Retirer « fond opaque #121c2c ».

Ne pas toucher alertes PNG, livegame, LCU, quiz.

---

## Interdits (déjà cassé avant)

1. `SetWindowLong` pour ajouter/retirer `WS_EX_TRANSPARENT` ou `WS_EX_LAYERED` au hover / au clic.
2. Relancer Chromium / recréer la WebView2 pour « réparer » le compositeur.
3. Fond contrôleur `A:255` « pour que ça se voie ».
4. `pointer-events: none` sur tout `#hud` sans re-enable sur les contrôles (casse Flash / drag).
5. Click-through **aussi** sur les widgets (le jeu recevrait le clic Flash).
6. CompositionController / DirectX overlay / injection League.
7. Plusieurs instances WebView2.
8. Changer le titre HWND / retirer TOPMOST / NOACTIVATE.

---

## Plan d’implémentation

1. Fond A=0 + CSS hudmode transparent. Vérifier hors widgets = map visible (Démo en jeu + HUD). Si slab navy → enfant pas clippé ou A=0 ignoré ; d’abord région enfant, re-tester A=0.
2. Helper `hudWebViewChildHWND()` + `applyHudRegion` sur parent **et** enfants. Appels après Embed, NavigationCompleted, Resize.
3. Serrer `reportHudHits`. Menu 👁 ouvert = hit du dropdown.
4. Slider opacité + localStorage + CSS variable.
5. Supprimer ou brancher `solid`.
6. `go test .` puis `.\run.ps1`. Test manuel ci-dessous.
7. Si click-through encore mort : dump des class names enfants (commentaire + une fois `OutputDebugString` / log fichier temporaire). Ajuster le filtre EnumChildWindows. En dernier recours : `WS_EX_TRANSPARENT` **enfant only** selon curseur vs hits.

---

## Tests manuels (partie entraînement, fenêtré sans bordure)

Click-through :

- [ ] Clic / drag caméra / ciblage **entre** les widgets → le jeu réagit, curseur League (pas flèche OS).
- [ ] Clic Flash / sort / horloge / 👁 / RST → le HUD réagit, le champion **ne** se déplace **pas**.
- [ ] Coins arrondis et gouttières du menu : clic traverse vers le jeu.
- [ ] Tab : cartes latérales cliquables ; le centre du scoreboard League reste cliquable.
- [ ] Molette sur la map (pas sur un widget) = zoom jeu.
- [ ] Resize ⠿ / ⤡ : on ne « perd » pas la souris.
- [ ] Masquer Objectifs dans 👁 : la zone devenue vide click-through.
- [ ] Aucun flash gris / rectangle navy plein écran au survol.

Opacité :

- [ ] Slider 100 % → widgets pleins.
- [ ] Slider ~50 % → widgets lisibles, map très visible **à travers les panneaux**.
- [ ] 25 % encore utilisable pour remonter le slider.
- [ ] Relancer le HUD : l’opacité est restée.
- [ ] Alertes Baron / toast item : visibles, fade avec le reste (ou exception documentée).

Régression :

- [ ] Alt+1–5 Flash, Alt+Maj+1–5 ult, Tab cartes.
- [ ] Auto-open HUD en partie.
- [ ] Croix ferme le widget, pas l’app.
- [ ] `go test .` vert.

---

## Critère « c’est bon »

En fenêtré sans bordure, on **joue** à travers l’overlay : les clics hors blocs vont dans League sans rectangle mort. Les blocs restent cliquables. Un slider 👁 baisse l’opacité de **tout** l’HUD, persisté, sans slab navy et sans flash gris.

Si après région-enfant + A=0 ce n’est toujours pas le cas : **stop** et documente dans le PR les class names des HWND enfants + ce que `GetWindowLong(GWL_EXSTYLE)` donne. Ne pas empiler un 4ᵉ hack (TRANSPARENT parent, recréer WebView, etc.).
