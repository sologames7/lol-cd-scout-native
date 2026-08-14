package main

// Curated priorities for fast in-game reading.
// Bot carry coverage was refreshed against the Patch 26.15 ADC population on 2026-08-11.
// Spell names and standard Q/W/E/R cooldowns are intentionally resolved from Data Dragon at runtime.

func autoSpell(key string, importance int, note string) Spell {
	return Spell{Spell: key, CD: "auto", Importance: importance, Note: note}
}

var overrides = map[string]Override{
	// Existing general-purpose examples.
	"Anivia": {
		Summary: "Œuf > Q > Mur",
		Window:  "Q raté = meilleure fenêtre de trade/all-in. Si l'œuf est disponible, prévois le second kill.",
		Important: []Spell{
			{Spell: "P", Name: "Renaissance (œuf)", CD: "240s", Importance: 10, Note: "Timer n°1"},
			autoSpell("Q", 10, "Stun / self-peel"),
			autoSpell("W", 7, "Mur"),
			autoSpell("R", 5, "Revient vite"),
		},
	},
	"Blitzcrank": {
		Summary:   "Q > E",
		Window:    "Q raté = grosse fenêtre pour avancer, ward ou engager.",
		Important: []Spell{autoSpell("Q", 10, "Sort clé"), autoSpell("E", 7, "Knock-up")},
	},
	"Morgana": {
		Summary:   "E > Q",
		Window:    "Bouclier noir utilisé = fenêtre importante pour CC sa cible.",
		Important: []Spell{autoSpell("E", 10, "Anti-CC"), autoSpell("Q", 8, "Root")},
	},
	"Zed": {
		Summary:   "W > R",
		Window:    "W utilisé offensivement = mobilité réduite jusqu'à son retour.",
		Important: []Spell{autoSpell("W", 10, "Mobilité / combo"), autoSpell("R", 9, "All-in")},
	},

	// ADC / APC bot — Patch 26.15 population.
	"Caitlyn": {
		Summary:   "E > W > R",
		Window:    "E utilisé = plus de dash de sécurité. C'est la meilleure fenêtre pour lui rentrer dessus.",
		Important: []Spell{autoSpell("E", 10, "Dash / self-peel"), autoSpell("W", 7, "Pièges / contrôle"), autoSpell("R", 5, "Finisher")},
	},
	"MissFortune": {
		Summary:   "R > E",
		Window:    "R interrompu ou indisponible = son gros teamfight disparaît. E utilisé = moins de setup/slow.",
		Important: []Spell{autoSpell("R", 10, "Teamfight"), autoSpell("E", 7, "Slow / setup")},
	},
	"Kaisa": {
		Summary:   "E > R > W",
		Window:    "E utilisé = moins de self-peel; R utilisé = plus de repositionnement/bouclier pour le prochain all-in.",
		Important: []Spell{autoSpell("E", 10, "Self-peel / invisibilité évoluée"), autoSpell("R", 9, "Reposition + bouclier"), autoSpell("W", 6, "Poke / marque")},
	},
	"Jhin": {
		Summary:   "W > R",
		Window:    "W raté = moins de catch à distance. Pendant R, il est immobile et vulnérable à l'engage/flank.",
		Important: []Spell{autoSpell("W", 9, "Root / catch"), autoSpell("R", 8, "Longue portée")},
	},
	"Ezreal": {
		Summary:   "E > R",
		Window:    "E utilisé = fenêtre majeure d'engage. Ne le laisse pas attendre son retour gratuitement.",
		Important: []Spell{autoSpell("E", 10, "Escape principal"), autoSpell("R", 6, "Global / wave")},
	},
	"Jinx": {
		Summary:   "E > W > R",
		Window:    "E utilisé = presque plus de self-peel. C'est la fenêtre la plus simple pour la dive/all-in.",
		Important: []Spell{autoSpell("E", 10, "Pièges / self-peel"), autoSpell("W", 7, "Slow / poke"), autoSpell("R", 6, "Execute global")},
	},
	"Ashe": {
		Summary:   "R > W",
		Window:    "R utilisé = grosse baisse de menace d'engage. W en CD ouvre une petite fenêtre de trade/push.",
		Important: []Spell{autoSpell("R", 10, "Engage global"), autoSpell("W", 7, "Poke / slow")},
	},
	"Tristana": {
		Summary:   "W > R > E",
		Window:    "W utilisé sans reset = fenêtre d'all-in. R down = elle perd son meilleur bouton de disengage.",
		Important: []Spell{autoSpell("W", 10, "Jump / reset"), autoSpell("R", 9, "Knockback"), autoSpell("E", 6, "Burst")},
	},
	"Lucian": {
		Summary:   "E > R",
		Window:    "E utilisé = fenêtre pour punir son positionnement. Le dash peut se réduire via son passif, donc ne compte pas au dixième près.",
		Important: []Spell{autoSpell("E", 10, "Dash"), autoSpell("R", 7, "Burst / zoning")},
	},
	"Samira": {
		Summary: "W > E > Style R",
		Window:  "W utilisé = projectiles/CC à distance beaucoup plus faciles à placer. E dépend des resets; R dépend surtout du rang de Style. Duo Nautilus: engage Q/R Naut → follow E pour Style.",
		Important: []Spell{
			autoSpell("W", 10, "Bloque les projectiles"),
			autoSpell("E", 8, "Dash / resets"),
			autoSpell("R", 6, "Ultime gated par Style"),
		},
		// Emerald+ patch 16.15 (LoLalytics) + counters duo Samira/Nautilus.
		Counters: []string{
			"Viktor", "Veigar", "Hwei", "Ziggs", "Nilah", "Xayah", "Kog'Maw", "Jinx",
			"Sivir", "Caitlyn", "Ezreal", "Brand", "Swain", "Seraphine",
		},
		HardMatchups: []string{
			"Morgana", "Braum", "Lulu", "Milio", "Janna", "Rell", "Leona", "Renata Glasc",
			"Ashe", "Senna", "Smolder", "Xerath", "Vel'Koz", "Karthus", "Malzahar",
		},
		Synergies: []string{
			"Nautilus", "Rell", "Leona", "Thresh", "Pyke", "Blitzcrank", "Alistar",
			"Graves", "Lee Sin", "Ahri", "Sylas", "Jarvan IV",
		},
	},
	"Nautilus": {
		Summary: "Q > R > E",
		Window:  "Q raté = grosse fenêtre pour avancer/trade/ward. R down = nettement moins de chain CC et d'all-in duo avec Samira.",
		Important: []Spell{
			autoSpell("Q", 10, "Hook / engage"),
			autoSpell("R", 9, "Chain CC / all-in"),
			autoSpell("E", 6, "Slow / zone"),
		},
		// Emerald+ patch 16.15 support vs support + peel anti-engage vs duo Samira.
		Counters: []string{
			"Rell", "Leona", "Braum", "Renata Glasc", "Taric", "Rakan", "Alistar", "Thresh",
		},
		HardMatchups: []string{
			"Morgana", "Lulu", "Milio", "Janna", "Poppy", "Zilean", "Karma", "Nami",
			"Sivir", "Xayah", "Ezreal", "Caitlyn",
		},
		Synergies: []string{
			"Samira", "Kai'Sa", "Miss Fortune", "Jhin", "Tristana", "Ezreal",
			"Graves", "Lee Sin", "Ahri", "Caitlyn",
		},
	},
	"Vayne": {
		Summary:   "E > R",
		Window:    "E utilisé = elle perd son stun/repoussement. R down = moins de duel et plus d'invisibilité via Q.",
		Important: []Spell{autoSpell("E", 10, "Condemn / stun mur"), autoSpell("R", 8, "Duel / invisibilité")},
	},
	"Yunara": {
		Summary:   "R > E > W",
		Window:    "R terminé = énorme baisse de puissance temporaire. E utilisé = moins de mobilité pour esquiver ou kite.",
		Important: []Spell{autoSpell("R", 10, "État transcendant"), autoSpell("E", 9, "Mobilité / dash sous R"), autoSpell("W", 6, "Slow / poke")},
	},
	"Ziggs": {
		Summary:   "W > E > R",
		Window:    "W utilisé = son principal self-peel disparaît. Engage avant qu'il puisse recréer de la distance.",
		Important: []Spell{autoSpell("W", 10, "Self-peel / tour"), autoSpell("E", 7, "Zone / slow"), autoSpell("R", 6, "Global")},
	},
	"Twitch": {
		Summary:   "Q > R > W",
		Window:    "Q utilisé = plus de repositionnement furtif immédiat. Attention aux resets de Q après élimination.",
		Important: []Spell{autoSpell("Q", 10, "Furtivité / reposition"), autoSpell("R", 9, "Portée / teamfight"), autoSpell("W", 5, "Slow / stacks")},
	},
	"Smolder": {
		Summary:   "E > R",
		Window:    "E utilisé = vraie fenêtre d'engage, surtout loin des murs. R down = moins de sustain et de retournement.",
		Important: []Spell{autoSpell("E", 10, "Escape / traverse terrain"), autoSpell("R", 8, "Heal + dégâts")},
	},
	"Mel": {
		Summary:   "W > E > R",
		Window:    "W utilisé = fenêtre critique: ses réflexions/projectile immunity ne sont plus disponibles. Ensuite, cherche le CC ou le burst.",
		Important: []Spell{autoSpell("W", 10, "Réflexion / défense"), autoSpell("E", 8, "Root / zone"), autoSpell("R", 7, "Burst à distance")},
	},
	"Xayah": {
		Summary:   "R > E",
		Window:    "R utilisé = énorme fenêtre pour l'all-in. E utilisé avec peu de plumes = moins de root/burst immédiat.",
		Important: []Spell{autoSpell("R", 10, "Invulnérabilité / self-peel"), autoSpell("E", 9, "Rappel des plumes / root")},
	},
	"Aphelios": {
		Summary:   "R > Q",
		Window:    "Le danger du Q dépend de l'arme équipée; chaque arme gère son propre usage. R down réduit fortement ses gros combos de teamfight.",
		Important: []Spell{autoSpell("R", 10, "Gros combo"), autoSpell("Q", 8, "Dépend de l'arme")},
	},
	"Sivir": {
		Summary:   "E > R",
		Window:    "E utilisé = fenêtre pour placer ton sort clé/CC. R down = moins de kite et d'engage collectif.",
		Important: []Spell{autoSpell("E", 10, "Spell shield"), autoSpell("R", 8, "Vitesse équipe")},
	},
	"Xerath": {
		Summary:   "E > R > W",
		Window:    "E raté = sa meilleure défense est partie. C'est la fenêtre pour fermer la distance.",
		Important: []Spell{autoSpell("E", 10, "Stun / self-peel"), autoSpell("R", 8, "Artillerie"), autoSpell("W", 5, "Slow")},
	},
	"Draven": {
		Summary:   "E > R",
		Window:    "E utilisé = moins de disengage/interruption. Son W se reset quand il rattrape une hache, donc le timer brut est peu utile.",
		Important: []Spell{autoSpell("E", 10, "Knock-aside / interrupt"), autoSpell("R", 7, "Execute global")},
	},
	"Seraphine": {
		Summary:   "R > W > E",
		Window:    "R utilisé = grosse fenêtre de teamfight. W utilisé = moins de survie; E raté = moins de contrôle immédiat.",
		Important: []Spell{autoSpell("R", 10, "Charm / engage"), autoSpell("W", 9, "Shield / heal"), autoSpell("E", 7, "CC")},
	},
	"Viktor": {
		Summary:   "W > R",
		Window:    "W utilisé = beaucoup moins de self-peel et de contrôle de zone. R down = moins de disruption/burst prolongé.",
		Important: []Spell{autoSpell("W", 10, "Zone / stun"), autoSpell("R", 8, "Disruption / DPS")},
	},
	"Syndra": {
		Summary:   "E > R",
		Window:    "E utilisé = fenêtre majeure pour dive/engage: elle perd son knockback/stun. R down = moins de burst ciblé.",
		Important: []Spell{autoSpell("E", 10, "Knockback / stun"), autoSpell("R", 8, "Burst")},
	},
	"Veigar": {
		Summary:   "E > R",
		Window:    "E utilisé = cage absente, donc grosse fenêtre pour entrer dans sa zone. R down = moins de menace d'exécution.",
		Important: []Spell{autoSpell("E", 10, "Cage / self-peel"), autoSpell("R", 8, "Execute")},
	},
	"Varus": {
		Summary:   "R > E > Q",
		Window:    "R utilisé = beaucoup moins de catch. E utilisé = moins de slow/anti-heal; Q reste surtout du poke.",
		Important: []Spell{autoSpell("R", 10, "Root / engage"), autoSpell("E", 7, "Slow / anti-heal"), autoSpell("Q", 5, "Poke")},
	},
	"Yasuo": {
		Summary:   "W > R",
		Window:    "W utilisé = grosse fenêtre pour les projectiles. Son E et son R dépendent de cibles/knock-ups, donc les CD bruts racontent moins toute l'histoire.",
		Important: []Spell{autoSpell("W", 10, "Mur de vent"), autoSpell("R", 7, "All-in sur knock-up")},
	},
	"Zeri": {
		Summary:   "E > R",
		Window:    "E utilisé = fenêtre majeure, surtout si elle n'est pas près d'un mur. R down = moins de kite et de DPS prolongé.",
		Important: []Spell{autoSpell("E", 10, "Dash / traverse murs"), autoSpell("R", 9, "Teamfight / kite")},
	},
	"Senna": {
		Summary:   "E > W > R",
		Window:    "E utilisé = moins de protection/reposition collective. W raté = fenêtre d'engage; R down = moins de shield global.",
		Important: []Spell{autoSpell("E", 9, "Shroud / reposition"), autoSpell("W", 8, "Root"), autoSpell("R", 7, "Global shield + dégâts")},
	},
	"KogMaw": {
		Summary:   "W > E",
		Window:    "Quand W expire, sa portée et son DPS chutent nettement: meilleure fenêtre pour trade ou engager.",
		Important: []Spell{autoSpell("W", 10, "Portée + %PV"), autoSpell("E", 7, "Slow / zone")},
	},
	"Hwei": {
		Summary:   "E > R > W",
		Window:    "Une compétence E utilisée met toute sa famille de CC en récupération: grosse fenêtre pour avancer. R down réduit son combo long.",
		Important: []Spell{autoSpell("E", 10, "Famille de CC"), autoSpell("R", 9, "Combo / slow"), autoSpell("W", 6, "Utilitaire")},
	},
	"Kalista": {
		Summary:   "R > E",
		Window:    "R down = moins de save/engage avec son support. E sans reset = petite fenêtre où son burst de Rend est indisponible.",
		Important: []Spell{autoSpell("R", 10, "Save / engage support"), autoSpell("E", 8, "Rend / resets")},
	},
	"Nilah": {
		Summary:   "W > E > R",
		Window:    "W utilisé = grosse fenêtre pour dégâts d'AA et certains effets. E a des charges; vérifie si elle en a encore avant de commit.",
		Important: []Spell{autoSpell("W", 10, "Esquive attaques / défense"), autoSpell("E", 9, "Dash à charges"), autoSpell("R", 8, "Pull / heal")},
	},
	"Brand": {
		Summary:   "Q > R",
		Window:    "Q raté = son stun immédiat disparaît. R down = beaucoup moins de menace dans un fight groupé.",
		Important: []Spell{autoSpell("Q", 10, "Stun si cible brûlée"), autoSpell("R", 8, "Teamfight / rebonds")},
	},
	"Swain": {
		Summary:   "E > R",
		Window:    "E raté = meilleure fenêtre pour avancer. R down = son all-in prolongé et son drain chutent fortement.",
		Important: []Spell{autoSpell("E", 10, "Root / pull setup"), autoSpell("R", 9, "Drain / all-in")},
	},
	"Corki": {
		Summary:   "W > E",
		Window:    "W utilisé = fenêtre d'engage avant qu'il puisse recréer la distance. Les missiles R fonctionnent surtout avec des munitions, pas un simple timer.",
		Important: []Spell{autoSpell("W", 10, "Dash / escape"), autoSpell("E", 7, "Shred / DPS")},
	},
	"Katarina": {
		Summary:   "E > R",
		Window:    "E peut se reset via dagues/takedowns: ne timer pas naïvement. Interrompre R reste prioritaire.",
		Important: []Spell{autoSpell("E", 9, "Shunpo / resets"), autoSpell("R", 10, "Canalisé / interruptible")},
	},
	"Velkoz": {
		Summary:   "E > R > Q",
		Window:    "E raté = grosse fenêtre pour entrer au contact. R est canalisé et peut être interrompu par du CC.",
		Important: []Spell{autoSpell("E", 10, "Knock-up / self-peel"), autoSpell("R", 9, "Canalisation"), autoSpell("Q", 5, "Slow / poke")},
	},
	"Vladimir": {
		Summary:   "W > R",
		Window:    "W utilisé = énorme fenêtre: il perd son principal outil d'invulnérabilité/escape. Punis avant son retour.",
		Important: []Spell{autoSpell("W", 10, "Pool / untargetable"), autoSpell("R", 8, "Amplification + heal")},
	},
	"AurelionSol": {
		Summary:   "W > R > E",
		Window:    "W utilisé = moins de reposition/escape. R down = beaucoup moins de menace de CC massif en teamfight.",
		Important: []Spell{autoSpell("W", 10, "Vol / reposition"), autoSpell("R", 9, "Knock-up / grosse zone"), autoSpell("E", 7, "Zone / execute")},
	},
	"Karthus": {
		Summary:   "R > W",
		Window:    "R down = menace globale absente. W utilisé = moins de slow et de réduction de résistance magique pour son combo.",
		Important: []Spell{autoSpell("R", 10, "Global"), autoSpell("W", 7, "Slow / shred MR")},
	},
	"Cassiopeia": {
		Summary:   "W > R",
		Window:    "W utilisé = grosse fenêtre pour les champions à dash. R raté ou down = elle perd son meilleur retournement frontal.",
		Important: []Spell{autoSpell("W", 10, "Ground / anti-dash"), autoSpell("R", 9, "Stun si face à elle")},
	},
	"Taliyah": {
		Summary:   "W > E > R",
		Window:    "W raté = beaucoup moins de catch/self-peel. E down = les dashs sont nettement plus sûrs.",
		Important: []Spell{autoSpell("W", 10, "Knockback / setup"), autoSpell("E", 9, "Anti-dash / zone"), autoSpell("R", 6, "Roam / wall")},
	},
}
