# Brief agent — bannières d’alerte d’objectifs (LoL CD Scout)

Tu prends le relais **uniquement pour régénérer les 12 images de bannières**. Le câblage Go/front est déjà en place. Ne recadre pas, ne compose pas avec PowerShell, ne réutilise pas les portraits carrés.

## Contexte produit
Outil HUD LoL. À **1:15** du spawn d’un objectif, une **seule image** s’affiche (notif). Pas de texte CSS par-dessus : tout est **cuit dans le PNG**.

Fichiers servis : `GET /alerts/{id}.png`  
Embed Go : `branding/alerts/*.png` (pas les sous-dossiers `art/` ni `text/`)  
IDs whitelist : `grubs | dragon | baron | herald | infernal | mountain | ocean | cloud | hextech | chemtech | soul | elder`

## Ce qui a foiré (ne pas recommencer)
1. **Portraits carrés collés à gauche** + texte CSS/GDI → l’utilisateur a rejeté.
2. **Réutiliser les mêmes images** (`branding/alerts/art/*.png`) en référence → illustrations pas adaptées au format bannière.
3. **Recadrer du 16:9 vers du 970×250** → titres Mario 64 coupés en haut (`BARON`, `DRAKE INFERNAL`, `HÉRAUT` mangés). **Le format généré 16:9 de base était le bon. NE PAS DÉCOUPER.**
4. **Listes numérotées dans le prompt** (`1. HERAUT` / `2. DANS 1:15` / `3. …`) → l’IA a **peint les numéros** dans l’image. Interdit.
5. **Tips de jeu** (« ward le pit », « pousse ta vague », « regroupe-toi ») → l’utilisateur les trouve **inutiles**. Troisième ligne = **avantage à la capture**, pas un conseil de macro.
6. Fautes IA fréquentes : `MONTANGE` (doit être **MONTAGNE**), `HÉRAU` (il manque le T), `regronpe-toi`, `DANS DANS 1:15`, titre répété deux fois (`DRAKE INFERNAL INFERNAL`). Relancer si ça arrive.
7. Prompt Herald trop violent (« charging / explose une tour ») → parfois **bloqué safety**. Formuler en mascotte cartoon, pas en gore.

## Format image (obligatoire)
- **Aspect : 16:9** (outil GenerateImage : `aspect_ratio: "16:9"`). C’est le format validé.
- **Ne pas cropper** après génération. Copier tel quel dans `branding/alerts/{id}.png`.
- Si le fichier est trop lourd (> ~400 Ko) : redimensionner **en conservant le 16:9** (ex. 960×540) + JPEG qualité ~88, garder l’extension `.png`. `apiAlertArt` détecte déjà JPEG vs PNG via magic bytes. **Ne change pas le ratio.**
- Composition type **bannière pub large** (leaderboard), **pas un poster vertical** :
  - **Gauche (~45 %)** : illustration **nouvelle**, créée pour ce format. Monstre **de profil**, qui traverse le cadre gauche → droite (frise panoramique). Gros contours noirs style Foot 2 Rue / cartoon. Fond couleur + speed lines **horizontales**.
  - **Droite (~55 %)** : typo **Super Mario 64 titre** : lettres 3D, face jaune-or, côtés rouge-orange, brillance blanche, gros contour noir. Les 3 lignes occupent presque toute la hauteur.
- **Pas** de petit icône carré. **Pas** de portrait centré. **Pas** de chrome UI (boutons, watermark, logo CD Scout).

## Les 3 lignes de texte (toujours dans cet ordre, jamais numérotées)
1. **OBJ** — nom de l’objectif, une ligne, capitales.
2. **TIMER** — toujours exactement `DANS 1:15` (l’alerte part à 1 min 15 du spawn).
3. **AVANTAGE** — ce que **gagne l’équipe qui le prend**. Court, lisible en 3D. **Pas** un tip (« ward », « groupe », « pousse ta vague »).

Dans le prompt, écrire : *Exactly three unnumbered lines, no "1." "2." "3.", no bullets, each line once.*

## Les 12 assets à générer

Remplacer uniquement les fichiers **à la racine** de `branding/alerts/`.  
**Ne pas** écraser `branding/alerts/art/` (portraits carrés, archives).  
**Ne pas** s’en servir comme `reference_image_paths` (c’est ça qui recyclait les mêmes images).

| Fichier | Sujet visuel (nouvelle illus, profil, large) | Ligne 1 OBJ | Ligne 2 | Ligne 3 AVANTAGE (capture) |
|---|---|---|---|---|
| `grubs.png` | 3 larves du Néant (vers violets, gros yeux, sourire), rangée horizontale, fond jaune | `LARVES` | `DANS 1:15` | `Degats permanents aux tours` |
| `herald.png` | Héraut de la Faille, cyclope violet, armure rocheuse, course vers la droite, fond magenta | `HERAUT` | `DANS 1:15` | `Siege de tour` |
| `baron.png` | Baron Nashor, long serpent violet, ventre or, 1 œil, fond violet | `BARON` | `DANS 1:15` | `Sbires boostes + rappel rapide` |
| `dragon.png` | Drake générique (premier spawn, type encore inconnu) : dragon rouge de profil, fond orange | `DRAKE` | `DANS 1:15` | `Stack de stats vers l ame` |
| `infernal.png` | Drake Infernal, rouge/feu, vol de profil, fond orange | `DRAKE INFERNAL` | `DANS 1:15` | `+AD/AP  ame explosions` |
| `mountain.png` | Drake Montagne, pierre brun/ocre, vol de profil | `DRAKE MONTAGNE` | `DANS 1:15` | `+Armure/MR  ame bouclier` |
| `ocean.png` | Drake Océan, teal, nage/vol de profil, fond cyan | `DRAKE OCEAN` | `DANS 1:15` | `Regen  ame soin au fight` |
| `cloud.png` | Drake Nuage, blanc/argent, traînées de vent, fond ciel | `DRAKE NUAGE` | `DANS 1:15` | `+Vitesse  ame dash post-ult` |
| `hextech.png` | Drake Hextech, or + cristaux cyan, étincelles | `DRAKE HEXTECH` | `DANS 1:15` | `Haste  ame foudre` |
| `chemtech.png` | Drake Chemtech, vert toxique / violet, slime | `DRAKE CHEMTECH` | `DANS 1:15` | `Tenacite  ame tank low HP` |
| `soul.png` | Même famille que l’Infernal mais **aura d’âme dorée** (4e drake = âme) | `DRAKE D AME` | `DANS 1:15` | `Ame permanente d equipe` |
| `elder.png` | Dragon Ancestral, or + charbon, feu menthe, fond or-violet | `ANCESTRAL` | `DANS 1:15` | `Brulure + execute a 20%` |

Accents : l’IA 3D Mario gère mal `É Â`. Préférer les formes **sans accent** dans l’image (`HERAUT`, `OCEAN`, `AME`, `Tenacite`, `Degats`, `Brulure`) plutôt qu’un glyphe cassé. L’UI FR réelle reste dans le code (`Héraut`, `ÂME`, etc.).

## Prompt type (copier-coller, une génération par fichier)

```
Wide 16:9 HUD notification banner, panoramic comic-strip layout, not a poster, not a square icon.
LEFT ~45%: brand-new cartoon illustration of [SUJET], SIDE PROFILE moving left-to-right, body filling the left half, Foot 2 Rue very thick black outlines, cel-shaded, high saturation, [COULEUR FOND] with HORIZONTAL speed lines.
RIGHT ~55%: Super Mario 64 title 3D lettering filling the height — yellow-gold faces, red-orange extruded sides, white shine, heavy black outline.

Exactly three unnumbered text lines, written once each, no "1." "2." "3.", no bullets, no extra English:
[LIGNE 1]
DANS 1:15
[LIGNE 3]

No watermark, no logo, no UI chrome, do not repeat the title.
```

`aspect_ratio`: `"16:9"`. **Aucun** `reference_image_paths` vers `art/` ni vers d’anciennes bannières (sinon tu recolles le même perso).

Relancer **individuellement** si : numéros peints, titre coupé, titre doublé, faute (`MONTANGE`), ou composition poster (tête centrée).

## Après génération
1. Copier chaque image vers `branding/alerts/{id}.png` **sans crop**.
2. Si > ~400 Ko : resize 960×540 (16:9) + JPEG q88, garder le nom `.png`.
3. `go test .` (dont `TestAlertArtEmbed` : JPEG `FF D8` ou PNG OK).
4. Relance app : `.\run.ps1` puis **Démo en jeu** → bannière `soul.png` (drake d’âme).

## Hors scope (ne pas toucher sauf si cassé)
- Front : `index.html` `bannerFile()` / `#page-alert` / `.alert-banner` — déjà branché.
- Go : `objsfx.go` `apiAlertArt`, `LiveObjective.Kind` (`infernal`…`soul`/`elder` depuis `DragonType` live).
- Pop item (portrait + tiroir-caisse) : autre feature, déjà livrée.
- `compose.ps1` + `text/*.png` : ancienne pipeline GDI, **ne plus s’en servir** pour ces bannières.

## Critère « c’est bon »
Une bannière 16:9 complète, illus **originale** adaptée à la largeur, 3 lignes Mario 64 **sans numéros** : nom + `DANS 1:15` + avantage de capture. Rien de coupé.
