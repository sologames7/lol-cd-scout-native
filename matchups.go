package main

// Auto-generated from LoLalytics Emerald+ counters (~patch 16.15/16.16).
// Counters = 12 pire WR; HardMatchups = 12 suivants.
// Une override curated avec Counters/HardMatchups non vides remplace ces listes.

type MatchupData struct {
	Counters     []string
	HardMatchups []string
}

var matchups = map[string]MatchupData{
	"Aatrox": {
		Counters:     []string{"Vel'Koz", "Ahri", "Zilean", "Katarina", "Lissandra", "Ekko", "Singed", "Zed", "Quinn", "Veigar", "Urgot", "Heimerdinger"},
		HardMatchups: []string{"Malphite", "Irelia", "Azir", "Briar", "Akshan", "Kai'Sa", "Warwick", "Kled", "Karma", "Fiora", "Cassiopeia", "Pantheon"},
	},
	"Ahri": {
		Counters:     []string{"Singed", "Kalista", "Swain", "Pantheon", "Neeko", "Kayle", "Annie", "Vladimir", "Akshan", "Katarina", "Qiyana", "Zed"},
		HardMatchups: []string{"Ekko", "Veigar", "Twisted Fate", "Quinn", "Tryndamere", "Vex", "Kled", "Naafiri", "Talon", "Xerath", "Viktor", "Kassadin"},
	},
	"Akali": {
		Counters:     []string{"Singed", "Gragas", "Kog'Maw", "Sett", "Morgana", "Kled", "Swain", "Karthus", "Twisted Fate", "Karma", "Rumble", "Galio"},
		HardMatchups: []string{"Vex", "Ahri", "Briar", "Tryndamere", "Kassadin", "Aatrox", "Lissandra", "Pantheon", "Riven", "Quinn", "Lucian", "Taliyah"},
	},
	"Akshan": {
		Counters:     []string{"Gragas", "Neeko", "Malphite", "Tryndamere", "Zilean", "Garen", "Vladimir", "Yasuo", "Cho'Gath", "Pantheon", "Gwen", "Irelia"},
		HardMatchups: []string{"Brand", "Vex", "Syndra", "Nasus", "Twisted Fate", "Viktor", "Katarina", "Riven", "Lissandra", "Diana", "Galio", "Quinn"},
	},
	"Alistar": {
		Counters:     []string{"Amumu", "Syndra", "Renata Glasc", "Senna", "Ivern", "Rakan", "Jax", "Nidalee", "Poppy", "Janna", "Zac", "Seraphine"},
		HardMatchups: []string{"Soraka", "Morgana", "Lulu", "Veigar", "Sona", "Karma", "Nami", "Thresh", "Hwei", "LeBlanc", "Zilean", "Milio"},
	},
	"Ambessa": {
		Counters:     []string{"Qiyana", "Warwick", "Cassiopeia", "Lissandra", "Wukong", "Camille", "Zilean", "Twisted Fate", "Irelia", "Ahri", "Singed", "Akshan"},
		HardMatchups: []string{"Poppy", "Tryndamere", "Renekton", "Malphite", "Varus", "Lee Sin", "Teemo", "Heimerdinger", "Gangplank", "Zed", "Sejuani", "Kled"},
	},
	"Amumu": {
		Counters:     []string{"Cho'Gath", "Zaahen", "Ivern", "Ambessa", "Aatrox", "Gragas", "Naafiri", "Mordekaiser", "Lillia", "Shyvana", "Taliyah", "Sylas"},
		HardMatchups: []string{"Nasus", "Shaco", "Zac", "Volibear", "Olaf", "Rek'Sai", "Dr. Mundo", "Maokai", "Xin Zhao", "Ekko", "Kayn", "Karthus"},
	},
	"Anivia": {
		Counters:     []string{"Zilean", "Riven", "Morgana", "Naafiri", "Ziggs", "Vel'Koz", "Ahri", "Hwei", "Kayle", "Annie", "Zoe", "Quinn"},
		HardMatchups: []string{"Kog'Maw", "Nasus", "Gwen", "Locke", "Kassadin", "Qiyana", "Aurelion Sol", "Veigar", "Katarina", "Garen", "Syndra", "Twisted Fate"},
	},
	"Annie": {
		Counters:     []string{"Viktor", "Sion", "Riven", "Naafiri", "Pantheon", "Syndra", "Talon", "Galio", "Twisted Fate", "Ekko", "Qiyana", "Cho'Gath"},
		HardMatchups: []string{"Xerath", "Diana", "Fizz", "Malzahar", "Veigar", "Zed", "Morgana", "Hwei", "Akshan", "Vel'Koz", "Vladimir", "Zoe"},
	},
	"Aphelios": {
		Counters:     []string{"Heimerdinger", "Zyra", "Vel'Koz", "Lux", "Tahm Kench", "Xerath", "Seraphine", "Vladimir", "Hwei", "Karthus", "Kog'Maw", "Nilah"},
		HardMatchups: []string{"Ziggs", "Yasuo", "Veigar", "Xayah", "Syndra", "Swain", "Viktor", "Samira", "Draven", "Ashe", "Jinx", "Zeri"},
	},
	"Ashe": {
		Counters:     []string{"Karthus", "Heimerdinger", "Xerath", "Zyra", "Hwei", "Lux", "Vel'Koz", "Senna", "Seraphine", "Anivia", "Tristana", "Brand"},
		HardMatchups: []string{"Ziggs", "Yasuo", "Viktor", "Twitch", "Malzahar", "Veigar", "Swain", "Aurelion Sol", "Tahm Kench", "Nilah", "Irelia", "Kalista"},
	},
	"AurelionSol": {
		Counters:     []string{"Riven", "Gwen", "Naafiri", "Fizz", "Katarina", "Kayle", "Yone", "Zed", "Ahri", "Kassadin", "LeBlanc", "Vladimir"},
		HardMatchups: []string{"Sion", "Jayce", "Hwei", "Annie", "Xerath", "Irelia", "Taliyah", "Lissandra", "Viktor", "Cassiopeia", "Vex", "Twisted Fate"},
	},
	"Aurora": {
		Counters:     []string{"Gragas", "Tryndamere", "Talon", "Kai'Sa", "Vel'Koz", "Neeko", "Swain", "Cho'Gath", "Jayce", "Zed", "Veigar", "Akshan"},
		HardMatchups: []string{"Katarina", "Annie", "Zoe", "Ekko", "Hwei", "Twisted Fate", "Lux", "Nasus", "Diana", "Viktor", "Fizz", "Kassadin"},
	},
	"Azir": {
		Counters:     []string{"Vex", "LeBlanc", "Xerath", "Zoe", "Syndra", "Kayle", "Vladimir", "Ziggs", "Ekko", "Brand", "Vel'Koz", "Twisted Fate"},
		HardMatchups: []string{"Nasus", "Aurora", "Lissandra", "Qiyana", "Talon", "Kassadin", "Viktor", "Akali", "Katarina", "Sylas", "Veigar", "Pantheon"},
	},
	"Bard": {
		Counters:     []string{"Taric", "Syndra", "Veigar", "Sona", "Elise", "Poppy", "Thresh", "Amumu", "Pantheon", "Rell", "Blitzcrank", "Cho'Gath"},
		HardMatchups: []string{"Maokai", "Zilean", "Nautilus", "Braum", "Tahm Kench", "Senna", "Janna", "Trundle", "Morgana", "Soraka", "Leona", "Galio"},
	},
	"Belveth": {
		Counters:     []string{"Taliyah", "Rek'Sai", "Riven", "Ivern", "Poppy", "Malphite", "Zyra", "Kindred", "Hecarim", "Wukong", "Sejuani", "Jarvan IV"},
		HardMatchups: []string{"Fizz", "Lillia", "Amumu", "Evelynn", "Udyr", "Zac", "Elise", "Maokai", "Rammus", "Kha'Zix", "Nidalee", "Darius"},
	},
	"Blitzcrank": {
		Counters:     []string{"Nidalee", "Taric", "Rakan", "Rell", "Alistar", "Jax", "Poppy", "Twisted Fate", "Lissandra", "Sion", "Elise", "Leona"},
		HardMatchups: []string{"Zac", "Braum", "Zoe", "Sona", "Maokai", "Janna", "Thresh", "Nautilus", "Zilean", "Annie", "Pantheon", "Shen"},
	},
	"Brand": {
		Counters:     []string{"Maokai", "Zilean", "Elise", "Karma", "Senna", "Zyra", "Milio", "Janna", "Galio", "Alistar", "Rell", "Yuumi"},
		HardMatchups: []string{"Sona", "Leona", "Soraka", "Lulu", "Thresh", "Hwei", "Nautilus", "Blitzcrank", "Taric", "Ashe", "Shaco", "Nami"},
	},
	"Braum": {
		Counters:     []string{"Taric", "Renata Glasc", "Senna", "Zac", "Zilean", "Rell", "Soraka", "Sona", "Poppy", "Alistar", "Nami", "Seraphine"},
		HardMatchups: []string{"Leona", "Morgana", "Thresh", "Rakan", "Janna", "Gragas", "Bard", "Galio", "Pyke", "Blitzcrank", "Brand", "Vel'Koz"},
	},
	"Briar": {
		Counters:     []string{"Sion", "Quinn", "Zyra", "Morgana", "Cho'Gath", "Aatrox", "Wukong", "Nasus", "Rammus", "Nocturne", "Jax", "Teemo"},
		HardMatchups: []string{"Shyvana", "Graves", "Poppy", "Talon", "Udyr", "Zaahen", "Rek'Sai", "Fizz", "Taliyah", "Amumu", "Hecarim", "Kindred"},
	},
	"Caitlyn": {
		Counters:     []string{"Zyra", "Karthus", "Katarina", "Master Yi", "Heimerdinger", "Naafiri", "Singed", "Vel'Koz", "Xerath", "Lux", "Hwei", "Tahm Kench"},
		HardMatchups: []string{"Seraphine", "Veigar", "Ziggs", "Yasuo", "Aurora", "Ahri", "Brand", "Aurelion Sol", "Viktor", "Syndra", "Swain", "Senna"},
	},
	"Camille": {
		Counters:     []string{"Taric", "Ivern", "Janna", "Rell", "Braum", "Poppy", "Leona", "Nami", "Rammus", "Sona", "Galio", "Rakan"},
		HardMatchups: []string{"Alistar", "Soraka", "Thresh", "Fiddlesticks", "Zilean", "Vex", "Milio", "Amumu", "Lulu", "Gragas", "Annie", "Nautilus"},
	},
	"Cassiopeia": {
		Counters:     []string{"Cho'Gath", "Syndra", "Xerath", "Qiyana", "Vel'Koz", "Annie", "Malzahar", "Taliyah", "Pantheon", "Hwei", "Ahri", "Tryndamere"},
		HardMatchups: []string{"Gwen", "Anivia", "Viktor", "Kayle", "Fizz", "Nasus", "Zed", "Lux", "Vladimir", "Katarina", "Twisted Fate", "Galio"},
	},
	"Chogath": {
		Counters:     []string{"Udyr", "Zed", "Zilean", "Vayne", "Illaoi", "Ornn", "Sett", "Shen", "Kled", "Quinn", "Varus", "Dr. Mundo"},
		HardMatchups: []string{"Anivia", "Mordekaiser", "Yone", "Master Yi", "Poppy", "Aatrox", "Zaahen", "Gwen", "Rengar", "K'Sante", "Yorick", "Warwick"},
	},
	"Corki": {
		Counters:     []string{"Hwei", "Brand", "Xerath", "Vel'Koz", "Veigar", "Kog'Maw", "Vladimir", "Senna", "Ziggs", "Viktor", "Jinx", "Karthus"},
		HardMatchups: []string{"Xayah", "Swain", "Smolder", "Seraphine", "Katarina", "Sivir", "Caitlyn", "Draven", "Cassiopeia", "Samira", "Ashe", "Aphelios"},
	},
	"Darius": {
		Counters:     []string{"Ahri", "Zilean", "Kog'Maw", "Lissandra", "Wukong", "Maokai", "Qiyana", "Aurora", "Briar", "Varus", "Malzahar", "Gangplank"},
		HardMatchups: []string{"Vayne", "Ashe", "Teemo", "Kennen", "Jax", "Urgot", "Quinn", "Anivia", "Naafiri", "Ornn", "Zed", "Dr. Mundo"},
	},
	"Diana": {
		Counters:     []string{"Kayle", "Sett", "Zilean", "Nasus", "Briar", "Singed", "Riven", "Gwen", "Kog'Maw", "Swain", "Vel'Koz", "Garen"},
		HardMatchups: []string{"Cho'Gath", "Gangplank", "Twisted Fate", "Tryndamere", "Gragas", "Nunu & Willump", "Lissandra", "Vex", "Taliyah", "Heimerdinger", "Quinn", "Morgana"},
	},
	"Draven": {
		Counters:     []string{"Karthus", "Yone", "Vel'Koz", "Malzahar", "Brand", "Hwei", "Seraphine", "Swain", "Vladimir", "Lux", "Zyra", "Viktor"},
		HardMatchups: []string{"Senna", "Xerath", "Tahm Kench", "Veigar", "Nilah", "Yasuo", "Syndra", "Kog'Maw", "Samira", "Ashe", "Mel", "Ziggs"},
	},
	"DrMundo": {
		Counters:     []string{"Kled", "Zilean", "Aurelion Sol", "Udyr", "Zac", "Warwick", "Irelia", "Gwen", "Aatrox", "Naafiri", "Illaoi", "Fiora"},
		HardMatchups: []string{"Yorick", "Ekko", "Ambessa", "Garen", "Shen", "Zed", "Zaahen", "Olaf", "Singed", "Yone", "Tryndamere", "Akshan"},
	},
	"Ekko": {
		Counters:     []string{"Kha'Zix", "Malphite", "Rek'Sai", "Evelynn", "Ambessa", "Kindred", "Talon", "Rammus", "Cho'Gath", "Ivern", "Shaco", "Aatrox"},
		HardMatchups: []string{"Briar", "Kayn", "Hecarim", "Udyr", "Skarner", "Lillia", "Teemo", "Zyra", "Vi", "Shyvana", "Xin Zhao", "Mordekaiser"},
	},
	"Elise": {
		Counters:     []string{"Rammus", "Gwen", "Karthus", "Trundle", "Gragas", "Malphite", "Zyra", "Amumu", "Shaco", "Skarner", "Zac", "Rek'Sai"},
		HardMatchups: []string{"Nocturne", "Briar", "Maokai", "Aatrox", "Shyvana", "Nunu & Willump", "Jarvan IV", "Hecarim", "Dr. Mundo", "Vi", "Udyr", "Nasus"},
	},
	"Evelynn": {
		Counters:     []string{"Trundle", "Zaahen", "Briar", "Teemo", "Darius", "Cho'Gath", "Zyra", "Fiddlesticks", "Gragas", "Rek'Sai", "Udyr", "Ambessa"},
		HardMatchups: []string{"Ivern", "Quinn", "Rammus", "Shyvana", "Xin Zhao", "Jayce", "Shaco", "Nidalee", "Skarner", "Jarvan IV", "Dr. Mundo", "Aatrox"},
	},
	"Ezreal": {
		Counters:     []string{"Naafiri", "Olaf", "Zed", "Tahm Kench", "Yasuo", "Heimerdinger", "Gwen", "Katarina", "Sivir", "Zyra", "Nilah", "Xerath"},
		HardMatchups: []string{"Kalista", "Karthus", "Malzahar", "Seraphine", "Vladimir", "Zeri", "Tristana", "Viktor", "Brand", "Vayne", "Master Yi", "Hwei"},
	},
	"Fiddlesticks": {
		Counters:     []string{"Cho'Gath", "Maokai", "Nocturne", "Trundle", "Udyr", "Nasus", "Rek'Sai", "Mordekaiser", "Zaahen", "Aatrox", "Olaf", "Ekko"},
		HardMatchups: []string{"Kindred", "Talon", "Darius", "Bel'Veth", "Briar", "Xin Zhao", "Zac", "Elise", "Rammus", "Shyvana", "Hecarim", "Jarvan IV"},
	},
	"Fiora": {
		Counters:     []string{"Ahri", "Zilean", "Akshan", "Aurora", "Wukong", "Skarner", "Tryndamere", "Camille", "Heimerdinger", "Malphite", "Urgot", "Xin Zhao"},
		HardMatchups: []string{"Quinn", "Kayle", "Lissandra", "Udyr", "Anivia", "Cassiopeia", "Warwick", "Poppy", "Viktor", "Sett", "Teemo", "Shen"},
	},
	"Fizz": {
		Counters:     []string{"Master Yi", "Aatrox", "Gragas", "Neeko", "Rumble", "Tryndamere", "Renekton", "Kennen", "Yorick", "Sett", "Ambessa", "Tristana"},
		HardMatchups: []string{"Zac", "Nasus", "Gangplank", "Kassadin", "Lissandra", "Akali", "Heimerdinger", "Zilean", "Riven", "Kayle", "Pantheon", "Vladimir"},
	},
	"Galio": {
		Counters:     []string{"Singed", "Quinn", "Vel'Koz", "K'Sante", "Zilean", "Zac", "Mordekaiser", "Cho'Gath", "Kayle", "Illaoi", "Sett", "Brand"},
		HardMatchups: []string{"Veigar", "Hwei", "Swain", "Ahri", "Gragas", "Vayne", "Ambessa", "Heimerdinger", "Yorick", "Jhin", "Pantheon", "Zed"},
	},
	"Gangplank": {
		Counters:     []string{"Ahri", "Xin Zhao", "Veigar", "Kayle", "Fiddlesticks", "Aurora", "Zed", "Kog'Maw", "Warwick", "Anivia", "Kled", "Quinn"},
		HardMatchups: []string{"Dr. Mundo", "Tryndamere", "Olaf", "Udyr", "Urgot", "Sion", "Tahm Kench", "Cho'Gath", "Aatrox", "Zac", "Camille", "Viktor"},
	},
	"Garen": {
		Counters:     []string{"Lissandra", "Ahri", "Quinn", "Viktor", "Vayne", "Ashe", "Kayle", "Qiyana", "Tryndamere", "Camille", "Akshan", "Teemo"},
		HardMatchups: []string{"Swain", "Rek'Sai", "Urgot", "Brand", "Gnar", "Kennen", "Tristana", "Cassiopeia", "Zed", "Anivia", "Kled", "Gangplank"},
	},
	"Gnar": {
		Counters:     []string{"Ahri", "Fiddlesticks", "Cassiopeia", "Viktor", "Akshan", "Zilean", "Lissandra", "Zed", "Poppy", "Kai'Sa", "Sylas", "Singed"},
		HardMatchups: []string{"Kayle", "Kled", "Ornn", "Akali", "Olaf", "Quinn", "Qiyana", "Malphite", "Kog'Maw", "Camille", "Kennen", "Irelia"},
	},
	"Gragas": {
		Counters:     []string{"Singed", "Maokai", "Ornn", "Sion", "Cho'Gath", "Heimerdinger", "Rumble", "Malzahar", "Renekton", "Olaf", "Yorick", "Dr. Mundo"},
		HardMatchups: []string{"Camille", "Poppy", "Kled", "Garen", "Lissandra", "Gangplank", "Quinn", "Fiora", "Shen", "Vladimir", "Irelia", "Ambessa"},
	},
	"Graves": {
		Counters:     []string{"Aurelion Sol", "Zyra", "Nidalee", "Poppy", "Udyr", "Dr. Mundo", "Fiddlesticks", "Wukong", "Rek'Sai", "Evelynn", "Taliyah", "Naafiri"},
		HardMatchups: []string{"Rammus", "Talon", "Ivern", "Sejuani", "Shyvana", "Skarner", "Aatrox", "Sion", "Mordekaiser", "Darius", "Gwen", "Hecarim"},
	},
	"Gwen": {
		Counters:     []string{"Master Yi", "Warwick", "Singed", "Vayne", "Zac", "Cassiopeia", "Tryndamere", "Yasuo", "Fiora", "Akali", "Garen", "Poppy"},
		HardMatchups: []string{"Yone", "Malzahar", "Sett", "Jax", "Ornn", "Jayce", "Aurora", "Kled", "Wukong", "Gragas", "Pantheon", "Quinn"},
	},
	"Hecarim": {
		Counters:     []string{"Taliyah", "Cho'Gath", "Riven", "Udyr", "Volibear", "Nasus", "Maokai", "Shyvana", "Rek'Sai", "Trundle", "Shen", "Olaf"},
		HardMatchups: []string{"Sylas", "Malphite", "Gwen", "Mordekaiser", "Lee Sin", "Evelynn", "Warwick", "Jarvan IV", "Aatrox", "Ivern", "Sejuani", "Gragas"},
	},
	"Heimerdinger": {
		Counters:     []string{"Galio", "Zed", "Veigar", "Yasuo", "Gangplank", "Viktor", "Dr. Mundo", "Aurora", "Ryze", "Anivia", "Kayle", "Akali"},
		HardMatchups: []string{"Vladimir", "Cho'Gath", "Ornn", "Quinn", "Garen", "Kled", "Lissandra", "Sett", "Gnar", "Yone", "Wukong", "Urgot"},
	},
	"Hwei": {
		Counters:     []string{"Quinn", "Kennen", "Gwen", "Kayle", "Gragas", "Ambessa", "Fizz", "Vel'Koz", "Neeko", "Katarina", "Xerath", "Garen"},
		HardMatchups: []string{"Zed", "Ahri", "Akshan", "Singed", "Talon", "Lux", "Naafiri", "Locke", "Zilean", "Aurelion Sol", "Ekko", "Syndra"},
	},
	"Illaoi": {
		Counters:     []string{"Zac", "Fiddlesticks", "Aurora", "Zed", "Ryze", "Teemo", "Anivia", "Cassiopeia", "Kayle", "Varus", "Brand", "Heimerdinger"},
		HardMatchups: []string{"Vayne", "Mordekaiser", "Malzahar", "Garen", "Warwick", "Kled", "Quinn", "Fiora", "Tryndamere", "Gwen", "Gangplank", "Yasuo"},
	},
	"Irelia": {
		Counters:     []string{"Singed", "Zac", "Warwick", "Jax", "Wukong", "Camille", "Sett", "Malphite", "Skarner", "Trundle", "Zaahen", "Garen"},
		HardMatchups: []string{"Volibear", "Tryndamere", "Maokai", "Poppy", "Xin Zhao", "Briar", "Sejuani", "Illaoi", "Udyr", "Renekton", "Fiora", "Shen"},
	},
	"Ivern": {
		Counters:     []string{"Darius", "Fiddlesticks", "Zyra", "Taliyah", "Sejuani", "Elise", "Quinn", "Talon", "Zac", "Zed", "Jarvan IV", "Gwen"},
		HardMatchups: []string{"Volibear", "Hecarim", "Evelynn", "Shaco", "Briar", "Naafiri", "Kha'Zix", "Nidalee", "Maokai", "Qiyana", "Aatrox", "Sylas"},
	},
	"Janna": {
		Counters:     []string{"Amumu", "Thresh", "Braum", "Blitzcrank", "Soraka", "Bard", "Leona", "Sona", "Alistar", "Senna", "Lee Sin", "Rakan"},
		HardMatchups: []string{"Jarvan IV", "Nami", "Zyra", "Zac", "Nautilus", "Annie", "Lulu", "Yuumi", "Ivern", "Maokai", "Shen", "Karma"},
	},
	"JarvanIV": {
		Counters:     []string{"Gwen", "Cho'Gath", "Shyvana", "Mordekaiser", "Fizz", "Zaahen", "Dr. Mundo", "Talon", "Rek'Sai", "Aatrox", "Wukong", "Nidalee"},
		HardMatchups: []string{"Kindred", "Briar", "Nasus", "Gragas", "Lillia", "Olaf", "Kha'Zix", "Ekko", "Sylas", "Zac", "Master Yi", "Evelynn"},
	},
	"Jax": {
		Counters:     []string{"Hwei", "Ahri", "Azir", "Rek'Sai", "Qiyana", "Shyvana", "Zac", "Singed", "Zoe", "Karma", "Garen", "Poppy"},
		HardMatchups: []string{"Dr. Mundo", "Urgot", "Quinn", "Illaoi", "Cassiopeia", "Vladimir", "Kennen", "Anivia", "Syndra", "Brand", "Aurora", "Kog'Maw"},
	},
	"Jayce": {
		Counters:     []string{"Sejuani", "Galio", "Quinn", "Vel'Koz", "Zac", "Xin Zhao", "Zoe", "Ahri", "Azir", "Graves", "Ornn", "Dr. Mundo"},
		HardMatchups: []string{"Irelia", "Fiddlesticks", "Malphite", "Qiyana", "Wukong", "Tahm Kench", "Poppy", "Anivia", "Olaf", "Illaoi", "Rengar", "Syndra"},
	},
	"Jhin": {
		Counters:     []string{"Fiora", "Singed", "Riven", "Gwen", "Katarina", "Karthus", "Yasuo", "Aurelion Sol", "Master Yi", "Irelia", "Tahm Kench", "Tristana"},
		HardMatchups: []string{"Seraphine", "Akshan", "Ekko", "Kog'Maw", "Lux", "Swain", "Zyra", "Zeri", "Twitch", "Samira", "Heimerdinger", "Veigar"},
	},
	"Jinx": {
		Counters:     []string{"Lux", "Hwei", "Master Yi", "Akshan", "Seraphine", "Xerath", "Ahri", "Vel'Koz", "Tahm Kench", "Katarina", "Veigar", "Karthus"},
		HardMatchups: []string{"Swain", "Twitch", "Yone", "Ziggs", "Irelia", "Zyra", "Heimerdinger", "Aurelion Sol", "Anivia", "Vladimir", "Viktor", "Tristana"},
	},
	"Kaisa": {
		Counters:     []string{"Riven", "Karthus", "Heimerdinger", "Nilah", "Kog'Maw", "Master Yi", "Xayah", "Gwen", "Vladimir", "Malzahar", "Swain", "Jinx"},
		HardMatchups: []string{"Zeri", "Cassiopeia", "Veigar", "Draven", "Sion", "Seraphine", "Aphelios", "Naafiri", "Zed", "Samira", "Zoe", "Hwei"},
	},
	"Kalista": {
		Counters:     []string{"Vel'Koz", "Brand", "Nilah", "Seraphine", "Cassiopeia", "Draven", "Tristana", "Yunara", "Veigar", "Aurelion Sol", "Sivir", "Zeri"},
		HardMatchups: []string{"Xayah", "Samira", "Aphelios", "Jinx", "Twitch", "Hwei", "Senna", "Lucian", "Varus", "Kog'Maw", "Lux", "Ashe"},
	},
	"Karma": {
		Counters:     []string{"Amumu", "Syndra", "Nautilus", "Sona", "Zilean", "Rell", "LeBlanc", "Poppy", "Blitzcrank", "Galio", "Taric", "Janna"},
		HardMatchups: []string{"Thresh", "Rakan", "Soraka", "Maokai", "Braum", "Nami", "Leona", "Jax", "Pyke", "Fiddlesticks", "Skarner", "Twisted Fate"},
	},
	"Karthus": {
		Counters:     []string{"Ivern", "Pantheon", "Nasus", "Skarner", "Volibear", "Darius", "Sejuani", "Quinn", "Teemo", "Aatrox", "Udyr", "Kindred"},
		HardMatchups: []string{"Briar", "Zyra", "Nunu & Willump", "Hecarim", "Evelynn", "Rammus", "Zac", "Nocturne", "Dr. Mundo", "Kayn", "Warwick", "Xin Zhao"},
	},
	"Kassadin": {
		Counters:     []string{"Quinn", "Heimerdinger", "Ambessa", "Riven", "Kog'Maw", "Zilean", "Corki", "Gwen", "Cho'Gath", "Nasus", "Galio", "Zed"},
		HardMatchups: []string{"Pantheon", "Lucian", "Tryndamere", "Akshan", "Kayle", "Yone", "Garen", "Gangplank", "Lissandra", "Hwei", "Qiyana", "Diana"},
	},
	"Katarina": {
		Counters:     []string{"Briar", "Zaahen", "Singed", "Kled", "Neeko", "Tryndamere", "Vladimir", "Sett", "Quinn", "Zac", "Gragas", "Sion"},
		HardMatchups: []string{"Kennen", "Vex", "Cho'Gath", "Wukong", "Heimerdinger", "Shen", "Kayle", "Jax", "Galio", "Lissandra", "Garen", "Mordekaiser"},
	},
	"Kayle": {
		Counters:     []string{"Zed", "Zac", "Irelia", "Malphite", "Pantheon", "Sylas", "Akshan", "Naafiri", "Ryze", "Jax", "Teemo", "Briar"},
		HardMatchups: []string{"Akali", "Nasus", "Zilean", "Wukong", "Yone", "Gragas", "Quinn", "Ambessa", "Camille", "Cassiopeia", "Anivia", "Cho'Gath"},
	},
	"Kayn": {
		Counters:     []string{"Rek'Sai", "Ivern", "Rammus", "Zaahen", "Wukong", "Maokai", "Quinn", "Darius", "Master Yi", "Fiddlesticks", "Ambessa", "Nocturne"},
		HardMatchups: []string{"Evelynn", "Zyra", "Shyvana", "Talon", "Elise", "Warwick", "Trundle", "Kha'Zix", "Fizz", "Gragas", "Jarvan IV", "Kindred"},
	},
	"Kennen": {
		Counters:     []string{"Sion", "Nasus", "Locke", "Trundle", "Dr. Mundo", "Anivia", "Ornn", "Gangplank", "Cho'Gath", "Yorick", "Zed", "Sylas"},
		HardMatchups: []string{"Malphite", "Pantheon", "Lissandra", "Gragas", "Gwen", "Kled", "Irelia", "Malzahar", "Master Yi", "Kayle", "Urgot", "Quinn"},
	},
	"Khazix": {
		Counters:     []string{"Sion", "Trundle", "Zyra", "Skarner", "Naafiri", "Rammus", "Aatrox", "Hecarim", "Nasus", "Zac", "Shyvana", "Talon"},
		HardMatchups: []string{"Nidalee", "Nocturne", "Ivern", "Briar", "Fiddlesticks", "Wukong", "Evelynn", "Amumu", "Sejuani", "Shaco", "Rek'Sai", "Pantheon"},
	},
	"Kindred": {
		Counters:     []string{"Ivern", "Poppy", "Sejuani", "Zyra", "Taliyah", "Shaco", "Darius", "Talon", "Cho'Gath", "Naafiri", "Kha'Zix", "Xin Zhao"},
		HardMatchups: []string{"Nasus", "Aatrox", "Wukong", "Shyvana", "Elise", "Briar", "Zac", "Trundle", "Rammus", "Nidalee", "Fiddlesticks", "Vi"},
	},
	"Kled": {
		Counters:     []string{"Maokai", "Cassiopeia", "Singed", "Teemo", "Fiora", "Zac", "Shen", "Kayle", "Tryndamere", "Vayne", "Zaahen", "Yone"},
		HardMatchups: []string{"Malphite", "Darius", "Ornn", "Camille", "Aurora", "Tahm Kench", "Kennen", "Garen", "Renekton", "Warwick", "Poppy", "Volibear"},
	},
	"KogMaw": {
		Counters:     []string{"Hwei", "Aurelion Sol", "Senna", "Seraphine", "Yasuo", "Lux", "Zeri", "Kalista", "Nilah", "Jinx", "Vel'Koz", "Veigar"},
		HardMatchups: []string{"Malzahar", "Viktor", "Brand", "Xerath", "Katarina", "Ziggs", "Smolder", "Tristana", "Syndra", "Xayah", "Vayne", "Twitch"},
	},
	"KSante": {
		Counters:     []string{"Lissandra", "Zilean", "Kayle", "Singed", "Udyr", "Zed", "Garen", "Poppy", "Fiora", "Fiddlesticks", "Quinn", "Riven"},
		HardMatchups: []string{"Smolder", "Vayne", "Camille", "Shen", "Illaoi", "Ambessa", "Wukong", "Briar", "Aurora", "Pantheon", "Zaahen", "Darius"},
	},
	"Leblanc": {
		Counters:     []string{"Gragas", "Naafiri", "Vex", "Sion", "Zilean", "Malzahar", "Kassadin", "Quinn", "Ekko", "Twisted Fate", "Taliyah", "Akshan"},
		HardMatchups: []string{"Vladimir", "Nasus", "Katarina", "Singed", "Viktor", "Gwen", "Gangplank", "Ahri", "Aurora", "Annie", "Pantheon", "Morgana"},
	},
	"LeeSin": {
		Counters:     []string{"Zaahen", "Sion", "Warwick", "Nasus", "Rek'Sai", "Taliyah", "Skarner", "Aatrox", "Rumble", "Naafiri", "Zac", "Shaco"},
		HardMatchups: []string{"Briar", "Morgana", "Udyr", "Shyvana", "Ivern", "Elise", "Cho'Gath", "Gwen", "Evelynn", "Darius", "Zyra", "Quinn"},
	},
	"Leona": {
		Counters:     []string{"Wukong", "Elise", "Qiyana", "Alistar", "Janna", "Rell", "Morgana", "Soraka", "Taric", "Braum", "Seraphine", "Poppy"},
		HardMatchups: []string{"Milio", "Bard", "Teemo", "Swain", "Thresh", "Zilean", "Senna", "Rakan", "Galio", "Shen", "Sona", "Nami"},
	},
	"Lillia": {
		Counters:     []string{"Morgana", "Zaahen", "Gwen", "Ivern", "Kindred", "Quinn", "Ambessa", "Darius", "Maokai", "Rek'Sai", "Viego", "Talon"},
		HardMatchups: []string{"Aatrox", "Hecarim", "Elise", "Qiyana", "Xin Zhao", "Evelynn", "Briar", "Fiddlesticks", "Gragas", "Volibear", "Diana", "Nidalee"},
	},
	"Lissandra": {
		Counters:     []string{"Kog'Maw", "Neeko", "Ziggs", "Jhin", "Varus", "Xerath", "Swain", "Kayle", "Mordekaiser", "Hwei", "Quinn", "Briar"},
		HardMatchups: []string{"Garen", "Vladimir", "Kennen", "Gragas", "Karma", "Viktor", "Cho'Gath", "Nunu & Willump", "Anivia", "Lux", "Annie", "Morgana"},
	},
	"Locke": {
		Counters:     []string{"Singed", "Shen", "Zaahen", "Kayle", "Kalista", "Zac", "Nasus", "Kassadin", "Briar", "Morgana", "Akali", "Ekko"},
		HardMatchups: []string{"Lee Sin", "Katarina", "Kled", "Gragas", "Gwen", "Zilean", "Galio", "Riven", "Nunu & Willump", "Yorick", "Heimerdinger", "Talon"},
	},
	"Lucian": {
		Counters:     []string{"Singed", "Heimerdinger", "Zed", "Ahri", "Veigar", "Tahm Kench", "Sion", "Master Yi", "Seraphine", "Fiora", "Taliyah", "Malzahar"},
		HardMatchups: []string{"Aurora", "Aurelion Sol", "Viktor", "Kog'Maw", "Lux", "Katarina", "Brand", "Nilah", "Vladimir", "Smolder", "Samira", "Vel'Koz"},
	},
	"Lulu": {
		Counters:     []string{"Cho'Gath", "Rammus", "Jarvan IV", "Leona", "Thresh", "Elise", "Sona", "Taric", "Rell", "Senna", "Blitzcrank", "Janna"},
		HardMatchups: []string{"Anivia", "Fiddlesticks", "Braum", "Zilean", "Nami", "Amumu", "Jayce", "Rakan", "Ahri", "Vel'Koz", "Soraka", "Annie"},
	},
	"Lux": {
		Counters:     []string{"Amumu", "Twisted Fate", "Elise", "LeBlanc", "Jax", "Heimerdinger", "Sona", "Blitzcrank", "Rell", "Hwei", "Annie", "Galio"},
		HardMatchups: []string{"Poppy", "Sylas", "Zac", "Pyke", "Soraka", "Janna", "Thresh", "Nami", "Zilean", "Pantheon", "Braum", "Yuumi"},
	},
	"Malphite": {
		Counters:     []string{"Annie", "Kassadin", "Zac", "Skarner", "Zilean", "Aurelion Sol", "Qiyana", "Dr. Mundo", "Sylas", "Ornn", "Anivia", "Zoe"},
		HardMatchups: []string{"Swain", "Mordekaiser", "Singed", "Fiddlesticks", "Fizz", "Shen", "Malzahar", "Tahm Kench", "Kog'Maw", "Sion", "Aurora", "Sejuani"},
	},
	"Malzahar": {
		Counters:     []string{"Singed", "Tryndamere", "Sion", "Kayle", "Quinn", "Karthus", "Neeko", "Zyra", "Viego", "Viktor", "Zac", "Twitch"},
		HardMatchups: []string{"Gangplank", "Gwen", "Taliyah", "Twisted Fate", "Ahri", "Galio", "Zilean", "Fizz", "Syndra", "Xerath", "Kassadin", "Morgana"},
	},
	"Maokai": {
		Counters:     []string{"Skarner", "Jarvan IV", "Milio", "Poppy", "Braum", "Taric", "Tahm Kench", "Alistar", "Janna", "Morgana", "Rell", "Senna"},
		HardMatchups: []string{"Zyra", "Thresh", "Nautilus", "Sona", "Blitzcrank", "Leona", "Swain", "Zilean", "Lulu", "Pantheon", "Shen", "Nami"},
	},
	"MasterYi": {
		Counters:     []string{"Sion", "Rek'Sai", "Wukong", "Rammus", "Shaco", "Elise", "Ivern", "Cho'Gath", "Zac", "Gragas", "Evelynn", "Udyr"},
		HardMatchups: []string{"Dr. Mundo", "Warwick", "Kha'Zix", "Volibear", "Jax", "Shen", "Talon", "Maokai", "Shyvana", "Aatrox", "Darius", "Lee Sin"},
	},
	"Mel": {
		Counters:     []string{"Viego", "Tahm Kench", "Aurelion Sol", "Karthus", "Malzahar", "Anivia", "Viktor", "Vladimir", "Senna", "Ziggs", "Hwei", "Vel'Koz"},
		HardMatchups: []string{"Syndra", "Lux", "Veigar", "Zyra", "Zeri", "Xayah", "Swain", "Brand", "Xerath", "Jinx", "Kai'Sa", "Corki"},
	},
	"Milio": {
		Counters:     []string{"Blitzcrank", "Braum", "Fiddlesticks", "Thresh", "Ivern", "Janna", "Rell", "Sona", "Amumu", "Rakan", "Lulu", "Galio"},
		HardMatchups: []string{"Alistar", "Zilean", "Taric", "Nami", "Vel'Koz", "Nautilus", "Skarner", "Bard", "Senna", "Gragas", "Elise", "Soraka"},
	},
	"MissFortune": {
		Counters:     []string{"Olaf", "Zyra", "Akshan", "Fiora", "Master Yi", "Ahri", "Orianna", "Nilah", "Tahm Kench", "Brand", "Locke", "Vel'Koz"},
		HardMatchups: []string{"Hwei", "Seraphine", "Lux", "Swain", "Viktor", "Zeri", "Heimerdinger", "Aurelion Sol", "Kog'Maw", "Katarina", "Senna", "Yone"},
	},
	"MonkeyKing": {
		Counters:     []string{"Taliyah", "Sion", "Tryndamere", "Gragas", "Zyra", "Skarner", "Lillia", "Karthus", "Nasus", "Hecarim", "Rek'Sai", "Gwen"},
		HardMatchups: []string{"Nidalee", "Darius", "Mordekaiser", "Shyvana", "Cho'Gath", "Ivern", "Riven", "Talon", "Nunu & Willump", "Shaco", "Aatrox", "Evelynn"},
	},
	"Mordekaiser": {
		Counters:     []string{"Vel'Koz", "Ahri", "Kog'Maw", "Zed", "Anivia", "Cassiopeia", "Qiyana", "Twisted Fate", "Vayne", "Viktor", "Zilean", "Heimerdinger"},
		HardMatchups: []string{"Aurora", "Singed", "Xin Zhao", "Veigar", "Brand", "Galio", "Ekko", "Katarina", "Yone", "Pantheon", "Olaf", "Malzahar"},
	},
	"Morgana": {
		Counters:     []string{"Elise", "Fiddlesticks", "Annie", "Sona", "Zilean", "Rell", "Janna", "Milio", "Poppy", "Senna", "Rakan", "Karma"},
		HardMatchups: []string{"Yuumi", "Zyra", "Shen", "Jarvan IV", "Pantheon", "Lulu", "Nami", "Blitzcrank", "Vel'Koz", "Xerath", "Braum", "LeBlanc"},
	},
	"Naafiri": {
		Counters:     []string{"Nocturne", "Hecarim", "Taliyah", "Poppy", "Xin Zhao", "Wukong", "Gragas", "Kayn", "Bel'Veth", "Zaahen", "Talon", "Shyvana"},
		HardMatchups: []string{"Fiddlesticks", "Zyra", "Rek'Sai", "Qiyana", "Briar", "Ekko", "Teemo", "Elise", "Warwick", "Aatrox", "Quinn", "Evelynn"},
	},
	"Nami": {
		Counters:     []string{"Rammus", "Ivern", "Nidalee", "Gragas", "Sona", "Trundle", "Zilean", "Janna", "Galio", "Thresh", "Blitzcrank", "Sion"},
		HardMatchups: []string{"Skarner", "Leona", "Amumu", "Rell", "Maokai", "Senna", "Singed", "Braum", "Nautilus", "Poppy", "Taric", "Elise"},
	},
	"Nasus": {
		Counters:     []string{"Anivia", "Zac", "Fiddlesticks", "Camille", "Sylas", "Urgot", "Shen", "Kled", "Swain", "Master Yi", "Aatrox", "Cho'Gath"},
		HardMatchups: []string{"Garen", "Ornn", "Quinn", "Gangplank", "Ambessa", "Yone", "Ahri", "Kayn", "Gragas", "Poppy", "Pantheon", "Singed"},
	},
	"Nautilus": {
		Counters:     []string{"Rell", "Trundle", "Lissandra", "Leona", "Renata Glasc", "Vex", "Braum", "Taric", "Rakan", "Alistar", "Singed", "Thresh"},
		HardMatchups: []string{"Zilean", "Janna", "Jarvan IV", "Sona", "Tahm Kench", "Senna", "Zac", "Seraphine", "Galio", "Soraka", "Morgana", "Taliyah"},
	},
	"Neeko": {
		Counters:     []string{"Renata Glasc", "Maokai", "Alistar", "Braum", "Zilean", "Shen", "Rell", "Elise", "Karma", "Senna", "Hwei", "Bard"},
		HardMatchups: []string{"Janna", "Vel'Koz", "Leona", "Poppy", "Zoe", "Sylas", "LeBlanc", "Brand", "Amumu", "Ashe", "Blitzcrank", "Galio"},
	},
	"Nidalee": {
		Counters:     []string{"Darius", "Amumu", "Rammus", "Udyr", "Warwick", "Briar", "Quinn", "Nunu & Willump", "Skarner", "Fizz", "Zac", "Ivern"},
		HardMatchups: []string{"Elise", "Vi", "Gragas", "Lillia", "Master Yi", "Fiddlesticks", "Shaco", "Nasus", "Diana", "Hecarim", "Talon", "Evelynn"},
	},
	"Nilah": {
		Counters:     []string{"Vladimir", "Lux", "Vel'Koz", "Corki", "Brand", "Veigar", "Viktor", "Seraphine", "Xayah", "Hwei", "Swain", "Senna"},
		HardMatchups: []string{"Xerath", "Kalista", "Jinx", "Sivir", "Syndra", "Smolder", "Tristana", "Varus", "Karthus", "Kog'Maw", "Ashe", "Jhin"},
	},
	"Nocturne": {
		Counters:     []string{"Zaahen", "Nasus", "Olaf", "Shyvana", "Rammus", "Udyr", "Tryndamere", "Aatrox", "Bel'Veth", "Wukong", "Dr. Mundo", "Gwen"},
		HardMatchups: []string{"Talon", "Taliyah", "Ivern", "Hecarim", "Evelynn", "Ekko", "Trundle", "Zac", "Jarvan IV", "Sion", "Cho'Gath", "Nidalee"},
	},
	"Nunu": {
		Counters:     []string{"Cho'Gath", "Zaahen", "Aatrox", "Gwen", "Ivern", "Quinn", "Fiddlesticks", "Taliyah", "Briar", "Shyvana", "Amumu", "Bel'Veth"},
		HardMatchups: []string{"Qiyana", "Malphite", "Xin Zhao", "Rammus", "Rek'Sai", "Jarvan IV", "Vi", "Kindred", "Lillia", "Olaf", "Sylas", "Master Yi"},
	},
	"Olaf": {
		Counters:     []string{"Zilean", "Aurora", "Vayne", "Kled", "Illaoi", "Kayle", "Wukong", "Master Yi", "Trundle", "Singed", "Naafiri", "Fiora"},
		HardMatchups: []string{"Sejuani", "Darius", "Camille", "Heimerdinger", "Quinn", "Sett", "Aatrox", "Tryndamere", "Anivia", "Akali", "Tahm Kench", "Jax"},
	},
	"Orianna": {
		Counters:     []string{"Zilean", "Vel'Koz", "Kog'Maw", "Singed", "Xerath", "Aurelion Sol", "Gwen", "LeBlanc", "Swain", "Katarina", "Morgana", "Nunu & Willump"},
		HardMatchups: []string{"Syndra", "Zoe", "Cho'Gath", "Locke", "Ahri", "Nasus", "Ekko", "Ambessa", "Hwei", "Brand", "Neeko", "Twisted Fate"},
	},
	"Ornn": {
		Counters:     []string{"Warwick", "Zed", "Poppy", "Singed", "Shen", "Rengar", "Olaf", "Fiora", "Yone", "Fiddlesticks", "Akshan", "Illaoi"},
		HardMatchups: []string{"Garen", "Ambessa", "Kayle", "Kled", "Dr. Mundo", "Zaahen", "Udyr", "Gangplank", "Shyvana", "Kayn", "Pantheon", "Yasuo"},
	},
	"Pantheon": {
		Counters:     []string{"Taric", "Janna", "Rell", "Skarner", "Sona", "Amumu", "Poppy", "Qiyana", "Elise", "Lulu", "Rakan", "Alistar"},
		HardMatchups: []string{"Rammus", "Twisted Fate", "Braum", "Thresh", "Zyra", "Milio", "Nami", "Locke", "Malphite", "Senna", "Soraka", "Karma"},
	},
	"Poppy": {
		Counters:     []string{"Gragas", "Taric", "Janna", "Zilean", "Sona", "Milio", "Thresh", "Elise", "Renata Glasc", "Nautilus", "Soraka", "Brand"},
		HardMatchups: []string{"Galio", "Braum", "Leona", "Senna", "Tahm Kench", "Nami", "Rengar", "Lulu", "Hwei", "Twisted Fate", "Rakan", "Rell"},
	},
	"Pyke": {
		Counters:     []string{"Ekko", "Poppy", "Zac", "Rell", "Renata Glasc", "Maokai", "Rakan", "Qiyana", "Gragas", "Annie", "Galio", "Thresh"},
		HardMatchups: []string{"LeBlanc", "Blitzcrank", "Leona", "Janna", "Nautilus", "Milio", "Alistar", "Skarner", "Nami", "Camille", "Wukong", "Locke"},
	},
	"Qiyana": {
		Counters:     []string{"Malphite", "Rammus", "Cho'Gath", "Nocturne", "Warwick", "Wukong", "Skarner", "Karthus", "Amumu", "Hecarim", "Fiddlesticks", "Elise"},
		HardMatchups: []string{"Fizz", "Taliyah", "Rek'Sai", "Jarvan IV", "Udyr", "Nasus", "Zyra", "Maokai", "Nidalee", "Shyvana", "Briar", "Vi"},
	},
	"Quinn": {
		Counters:     []string{"Rammus", "Dr. Mundo", "Ambessa", "Talon", "Elise", "Shyvana", "Wukong", "Aatrox", "Amumu", "Zac", "Fizz", "Xin Zhao"},
		HardMatchups: []string{"Kha'Zix", "Sejuani", "Ekko", "Taliyah", "Shaco", "Kindred", "Hecarim", "Viego", "Jarvan IV", "Diana", "Rek'Sai", "Udyr"},
	},
	"Rakan": {
		Counters:     []string{"Twisted Fate", "Maokai", "Neeko", "Senna", "Skarner", "Lissandra", "Taric", "Renata Glasc", "Janna", "Rell", "Soraka", "Poppy"},
		HardMatchups: []string{"Seraphine", "Hwei", "Galio", "Leona", "Sona", "Braum", "Thresh", "Amumu", "Shaco", "Bard", "Zilean", "Nami"},
	},
	"Rammus": {
		Counters:     []string{"Zaahen", "Maokai", "Zyra", "Darius", "Nasus", "Shyvana", "Gwen", "Lillia", "Udyr", "Gragas", "Zac", "Vi"},
		HardMatchups: []string{"Ivern", "Teemo", "Wukong", "Jarvan IV", "Rek'Sai", "Hecarim", "Dr. Mundo", "Kindred", "Bel'Veth", "Shaco", "Volibear", "Jayce"},
	},
	"RekSai": {
		Counters:     []string{"Trundle", "Zyra", "Ivern", "Nasus", "Dr. Mundo", "Aatrox", "Zaahen", "Quinn", "Nocturne", "Taliyah", "Kindred", "Nidalee"},
		HardMatchups: []string{"Shyvana", "Udyr", "Xin Zhao", "Rammus", "Skarner", "Graves", "Fiddlesticks", "Zac", "Briar", "Vi", "Jarvan IV", "Shaco"},
	},
	"Rell": {
		Counters:     []string{"Janna", "Annie", "Poppy", "Renata Glasc", "Alistar", "Soraka", "Fiddlesticks", "Zilean", "Zac", "Senna", "Galio", "Twisted Fate"},
		HardMatchups: []string{"Leona", "Thresh", "Amumu", "Jarvan IV", "Elise", "Braum", "Nami", "Sona", "Seraphine", "Maokai", "Bard", "Rakan"},
	},
	"Renata": {
		Counters:     []string{"Taric", "Zilean", "Janna", "Milio", "Senna", "Karma", "Bard", "Leona", "Lulu", "Sona", "Zyra", "Soraka"},
		HardMatchups: []string{"Blitzcrank", "Vel'Koz", "Nami", "Morgana", "Shen", "Alistar", "Shaco", "Xerath", "Thresh", "Pantheon", "Lux", "Rell"},
	},
	"Renekton": {
		Counters:     []string{"Ahri", "Kog'Maw", "Rek'Sai", "Illaoi", "Dr. Mundo", "Quinn", "Vayne", "Veigar", "Kayle", "Zilean", "Gangplank", "Singed"},
		HardMatchups: []string{"Heimerdinger", "Cassiopeia", "Ornn", "Zac", "Mordekaiser", "Kennen", "Maokai", "Swain", "Rumble", "Fiddlesticks", "Anivia", "Brand"},
	},
	"Rengar": {
		Counters:     []string{"Trundle", "Zaahen", "Zac", "Olaf", "Dr. Mundo", "Quinn", "Nasus", "Rammus", "Warwick", "Jax", "Fizz", "Aatrox"},
		HardMatchups: []string{"Master Yi", "Nunu & Willump", "Shen", "Briar", "Bel'Veth", "Wukong", "Udyr", "Mordekaiser", "Hecarim", "Naafiri", "Jarvan IV", "Darius"},
	},
	"Riven": {
		Counters:     []string{"Cassiopeia", "Quinn", "Wukong", "Urgot", "Zaahen", "Camille", "Zilean", "Fiddlesticks", "Gragas", "Renekton", "Maokai", "Illaoi"},
		HardMatchups: []string{"Kled", "Poppy", "Cho'Gath", "Malphite", "Olaf", "Garen", "Vayne", "Singed", "Lissandra", "Jayce", "Varus", "Kennen"},
	},
	"Rumble": {
		Counters:     []string{"Warwick", "Illaoi", "Anivia", "Swain", "Kled", "Tryndamere", "Galio", "Heimerdinger", "Zed", "Kayle", "Yone", "Cho'Gath"},
		HardMatchups: []string{"Sylas", "Varus", "Ryze", "Master Yi", "Zaahen", "Yorick", "Ornn", "Mordekaiser", "Riven", "Singed", "Shen", "Camille"},
	},
	"Ryze": {
		Counters:     []string{"Gragas", "Kayle", "Singed", "Aurelion Sol", "Ziggs", "Zilean", "Swain", "Neeko", "Nasus", "Malzahar", "Tryndamere", "Veigar"},
		HardMatchups: []string{"Xerath", "Syndra", "Hwei", "Jayce", "Aurora", "Vel'Koz", "Gwen", "Cho'Gath", "Akshan", "Viktor", "Anivia", "Twisted Fate"},
	},
	"Samira": {
		Counters:     []string{"Karthus", "Malzahar", "Viktor", "Veigar", "Aurelion Sol", "Taliyah", "Hwei", "Xerath", "Vladimir", "Swain", "Brand", "Ziggs"},
		HardMatchups: []string{"Nilah", "Xayah", "Kog'Maw", "Lux", "Vel'Koz", "Seraphine", "Syndra", "Jinx", "Zeri", "Cassiopeia", "Senna", "Tahm Kench"},
	},
	"Sejuani": {
		Counters:     []string{"Zaahen", "Rek'Sai", "Ambessa", "Pantheon", "Sylas", "Lillia", "Olaf", "Gwen", "Fiddlesticks", "Shyvana", "Rammus", "Aatrox"},
		HardMatchups: []string{"Dr. Mundo", "Mordekaiser", "Wukong", "Xin Zhao", "Trundle", "Fizz", "Nasus", "Lee Sin", "Jarvan IV", "Nocturne", "Nidalee", "Kayn"},
	},
	"Senna": {
		Counters:     []string{"Gragas", "Ivern", "Zac", "Twisted Fate", "Elise", "Blitzcrank", "Taric", "Janna", "Pyke", "LeBlanc", "Zyra", "Leona"},
		HardMatchups: []string{"Thresh", "Vel'Koz", "Zilean", "Rell", "Nami", "Milio", "Soraka", "Amumu", "Zoe", "Seraphine", "Poppy", "Camille"},
	},
	"Seraphine": {
		Counters:     []string{"Rammus", "Qiyana", "Amumu", "Nidalee", "Milio", "Lissandra", "Zilean", "Taric", "Gragas", "Janna", "Annie", "Soraka"},
		HardMatchups: []string{"Renata Glasc", "Sona", "Bard", "Pyke", "Nami", "Ivern", "Poppy", "Zed", "Twisted Fate", "Yuumi", "Rell", "Skarner"},
	},
	"Sett": {
		Counters:     []string{"Zilean", "Kog'Maw", "Zed", "Swain", "Anivia", "Kayle", "Malzahar", "Kai'Sa", "Twisted Fate", "Vayne", "Quinn", "Ahri"},
		HardMatchups: []string{"Kled", "Singed", "Kennen", "Wukong", "Zac", "Aatrox", "Malphite", "Ornn", "Pantheon", "Vladimir", "Renekton", "Cassiopeia"},
	},
	"Shaco": {
		Counters:     []string{"Morgana", "Nasus", "Nocturne", "Talon", "Pantheon", "Hecarim", "Nunu & Willump", "Rek'Sai", "Sejuani", "Aatrox", "Graves", "Naafiri"},
		HardMatchups: []string{"Maokai", "Darius", "Tryndamere", "Ivern", "Jarvan IV", "Briar", "Bel'Veth", "Nidalee", "Lillia", "Zac", "Udyr", "Shyvana"},
	},
	"Shen": {
		Counters:     []string{"Brand", "Anivia", "Qiyana", "Twisted Fate", "Singed", "Zac", "Ahri", "Malzahar", "Heimerdinger", "Kayn", "Illaoi", "Viktor"},
		HardMatchups: []string{"Vayne", "Mordekaiser", "Gangplank", "Warwick", "Teemo", "Aatrox", "Olaf", "Sett", "Urgot", "Gragas", "Pantheon", "Darius"},
	},
	"Shyvana": {
		Counters:     []string{"Sylas", "Darius", "Nidalee", "Taliyah", "Nasus", "Aatrox", "Ivern", "Udyr", "Morgana", "Trundle", "Talon", "Rek'Sai"},
		HardMatchups: []string{"Cho'Gath", "Evelynn", "Lillia", "Fizz", "Zyra", "Ambessa", "Bel'Veth", "Wukong", "Graves", "Olaf", "Zac", "Elise"},
	},
	"Singed": {
		Counters:     []string{"Twisted Fate", "Fiddlesticks", "Gangplank", "Lissandra", "Urgot", "Kayle", "Quinn", "Anivia", "Kassadin", "Zilean", "Master Yi", "Teemo"},
		HardMatchups: []string{"Zed", "Galio", "Rengar", "Darius", "Tryndamere", "Briar", "Fiora", "Camille", "Jayce", "Cho'Gath", "Kennen", "Varus"},
	},
	"Sion": {
		Counters:     []string{"Skarner", "Anivia", "Riven", "Zilean", "Viktor", "Zaahen", "Dr. Mundo", "Aatrox", "Zed", "Singed", "Shen", "Garen"},
		HardMatchups: []string{"Darius", "Fiddlesticks", "Yone", "Udyr", "Kled", "Renekton", "Kayle", "Vayne", "Poppy", "Fiora", "Tryndamere", "Ambessa"},
	},
	"Sivir": {
		Counters:     []string{"Viego", "Aurelion Sol", "Kog'Maw", "Katarina", "Seraphine", "Twitch", "Cassiopeia", "Lux", "Brand", "Viktor", "Yunara", "Senna"},
		HardMatchups: []string{"Zeri", "Hwei", "Aphelios", "Jinx", "Karthus", "Nilah", "Vayne", "Ashe", "Vel'Koz", "Draven", "Vladimir", "Xayah"},
	},
	"Skarner": {
		Counters:     []string{"Dr. Mundo", "Darius", "Ambessa", "Ivern", "Lillia", "Gwen", "Shyvana", "Teemo", "Nasus", "Shaco", "Bel'Veth", "Taliyah"},
		HardMatchups: []string{"Trundle", "Maokai", "Hecarim", "Nunu & Willump", "Aatrox", "Jayce", "Nocturne", "Udyr", "Briar", "Diana", "Vi", "Fiddlesticks"},
	},
	"Smolder": {
		Counters:     []string{"Karthus", "Heimerdinger", "Lux", "Yasuo", "Viego", "Syndra", "Katarina", "Brand", "Aurelion Sol", "Aphelios", "Seraphine", "Twitch"},
		HardMatchups: []string{"Zeri", "Hwei", "Vladimir", "Vel'Koz", "Xerath", "Jinx", "Sivir", "Kai'Sa", "Ziggs", "Kog'Maw", "Veigar", "Viktor"},
	},
	"Sona": {
		Counters:     []string{"Vi", "Sion", "Thresh", "Elise", "Rell", "Taric", "Maokai", "Neeko", "Janna", "Senna", "Leona", "Soraka"},
		HardMatchups: []string{"Annie", "Alistar", "Nautilus", "Rakan", "Blitzcrank", "Tahm Kench", "Braum", "Yuumi", "Renata Glasc", "Zyra", "Pyke", "Milio"},
	},
	"Soraka": {
		Counters:     []string{"Ivern", "Twisted Fate", "Jarvan IV", "Zac", "Nami", "Milio", "Thresh", "Elise", "Blitzcrank", "Sona", "Annie", "Janna"},
		HardMatchups: []string{"Senna", "Rell", "Maokai", "Lulu", "Yuumi", "Morgana", "Neeko", "Alistar", "Galio", "Leona", "Bard", "Nautilus"},
	},
	"Swain": {
		Counters:     []string{"Elise", "Nami", "Hwei", "Janna", "Soraka", "Annie", "Karma", "Renata Glasc", "Sona", "Fiddlesticks", "Milio", "Senna"},
		HardMatchups: []string{"Lulu", "Yuumi", "Zilean", "Braum", "Morgana", "Brand", "Poppy", "Seraphine", "Vel'Koz", "Zoe", "Taric", "Zyra"},
	},
	"Sylas": {
		Counters:     []string{"Heimerdinger", "Kled", "Fiora", "Quinn", "Singed", "Kog'Maw", "Illaoi", "Briar", "Neeko", "Vex", "Riven", "Gwen"},
		HardMatchups: []string{"Swain", "K'Sante", "Taliyah", "Kayle", "Hwei", "Tryndamere", "Galio", "Vel'Koz", "Pantheon", "Ornn", "Zoe", "Brand"},
	},
	"Syndra": {
		Counters:     []string{"Xin Zhao", "Kog'Maw", "Gwen", "Gragas", "Katarina", "Zilean", "Nasus", "Ambessa", "Singed", "Rumble", "Fizz", "Morgana"},
		HardMatchups: []string{"Xerath", "Kassadin", "Zed", "Ekko", "Swain", "Heimerdinger", "Aurelion Sol", "Naafiri", "Cho'Gath", "Master Yi", "Locke", "Irelia"},
	},
	"TahmKench": {
		Counters:     []string{"Rell", "Zilean", "Fiddlesticks", "Thresh", "Taric", "Nami", "Renata Glasc", "Sona", "Janna", "Rakan", "Blitzcrank", "Zac"},
		HardMatchups: []string{"Lulu", "Pyke", "Braum", "Leona", "Soraka", "Morgana", "Senna", "Vel'Koz", "LeBlanc", "Seraphine", "Alistar", "Milio"},
	},
	"Taliyah": {
		Counters:     []string{"Xerath", "Zed", "Brand", "Vel'Koz", "Katarina", "Swain", "Annie", "Talon", "Kassadin", "Zoe", "Fizz", "Ekko"},
		HardMatchups: []string{"Lux", "Hwei", "Lissandra", "Syndra", "Veigar", "Twisted Fate", "Gangplank", "Locke", "Viktor", "Gwen", "Vladimir", "Ahri"},
	},
	"Talon": {
		Counters:     []string{"Gwen", "Rammus", "Taliyah", "Warwick", "Aatrox", "Udyr", "Rek'Sai", "Nasus", "Amumu", "Zac", "Hecarim", "Sion"},
		HardMatchups: []string{"Bel'Veth", "Evelynn", "Elise", "Nidalee", "Skarner", "Mordekaiser", "Nunu & Willump", "Zyra", "Sejuani", "Kayn", "Pantheon", "Lee Sin"},
	},
	"Taric": {
		Counters:     []string{"Janna", "Alistar", "Zilean", "Soraka", "Rell", "Nami", "Morgana", "Lux", "Rakan", "Milio", "Pyke", "Maokai"},
		HardMatchups: []string{"Neeko", "Leona", "Brand", "Thresh", "Lulu", "Sona", "Swain", "Elise", "Seraphine", "Senna", "Karma", "Fiddlesticks"},
	},
	"Teemo": {
		Counters:     []string{"Zac", "Annie", "Galio", "Karma", "Aurelion Sol", "Fiddlesticks", "Naafiri", "Zed", "Ekko", "Katarina", "Viktor", "Zilean"},
		HardMatchups: []string{"Hwei", "Olaf", "Ornn", "Sion", "Kassadin", "Yasuo", "Ahri", "Sejuani", "Ryze", "Malphite", "Irelia", "Aurora"},
	},
	"Thresh": {
		Counters:     []string{"Annie", "Wukong", "Ekko", "Zilean", "Taric", "Galio", "Rell", "Renata Glasc", "Taliyah", "Braum", "Leona", "Seraphine"},
		HardMatchups: []string{"Rammus", "Rakan", "Alistar", "Janna", "Syndra", "Morgana", "Blitzcrank", "Amumu", "Hwei", "Soraka", "Senna", "Elise"},
	},
	"Tristana": {
		Counters:     []string{"Heimerdinger", "Karthus", "Master Yi", "Zyra", "Hwei", "Veigar", "Yasuo", "Vladimir", "Ahri", "Brand", "Lux", "Viktor"},
		HardMatchups: []string{"Cassiopeia", "Xayah", "Seraphine", "Vel'Koz", "Corki", "Draven", "Katarina", "Syndra", "Nilah", "Ziggs", "Kog'Maw", "Xerath"},
	},
	"Trundle": {
		Counters:     []string{"Ryze", "Quinn", "Sylas", "Wukong", "Jax", "Warwick", "Heimerdinger", "Malphite", "Gnar", "Kayle", "Singed", "Gragas"},
		HardMatchups: []string{"Ornn", "Gangplank", "Aatrox", "Tryndamere", "Akali", "Teemo", "Mordekaiser", "Anivia", "Shen", "Vayne", "Kled", "Urgot"},
	},
	"Tryndamere": {
		Counters:     []string{"Quinn", "Malphite", "Ryze", "Udyr", "Gragas", "Vayne", "Warwick", "Poppy", "Camille", "Cassiopeia", "Tahm Kench", "Viktor"},
		HardMatchups: []string{"Teemo", "Nasus", "Darius", "Ornn", "Fiddlesticks", "Jax", "Shen", "Anivia", "Volibear", "Kayle", "Heimerdinger", "Gnar"},
	},
	"TwistedFate": {
		Counters:     []string{"Singed", "Rumble", "Karthus", "Ambessa", "Mordekaiser", "Xin Zhao", "Neeko", "Swain", "Teemo", "Kai'Sa", "Dr. Mundo", "Gragas"},
		HardMatchups: []string{"Fizz", "Sion", "Hwei", "Gwen", "Xerath", "Veigar", "Quinn", "Lissandra", "Kassadin", "Locke", "Pantheon", "Naafiri"},
	},
	"Twitch": {
		Counters:     []string{"Nilah", "Tahm Kench", "Yone", "Viego", "Zyra", "Katarina", "Karthus", "Anivia", "Aurelion Sol", "Yasuo", "Swain", "Vladimir"},
		HardMatchups: []string{"Viktor", "Draven", "Samira", "Xerath", "Tristana", "Taliyah", "Zeri", "Kog'Maw", "Lux", "Hwei", "Xayah", "Kai'Sa"},
	},
	"Udyr": {
		Counters:     []string{"Teemo", "Aatrox", "Ivern", "Kindred", "Zed", "Nasus", "Rek'Sai", "Wukong", "Jarvan IV", "Elise", "Zaahen", "Amumu"},
		HardMatchups: []string{"Zac", "Lillia", "Sylas", "Kha'Zix", "Shaco", "Quinn", "Darius", "Maokai", "Nunu & Willump", "Warwick", "Naafiri", "Jax"},
	},
	"Urgot": {
		Counters:     []string{"Zilean", "Naafiri", "Zed", "Anivia", "Wukong", "Kayle", "Maokai", "Yorick", "Dr. Mundo", "Malphite", "Ornn", "Aurora"},
		HardMatchups: []string{"Kled", "Brand", "Olaf", "Heimerdinger", "Vladimir", "Ambessa", "Camille", "Sion", "Shen", "Quinn", "Illaoi", "Udyr"},
	},
	"Varus": {
		Counters:     []string{"Aurelion Sol", "Kog'Maw", "Yasuo", "Karthus", "Tristana", "Xerath", "Ziggs", "Senna", "Lux", "Zeri", "Veigar", "Katarina"},
		HardMatchups: []string{"Vel'Koz", "Viktor", "Brand", "Seraphine", "Miss Fortune", "Smolder", "Ashe", "Swain", "Jhin", "Jinx", "Syndra", "Sivir"},
	},
	"Vayne": {
		Counters:     []string{"Zyra", "Aurelion Sol", "Vladimir", "Xayah", "Tristana", "Nilah", "Aphelios", "Samira", "Draven", "Corki", "Ahri", "Veigar"},
		HardMatchups: []string{"Zeri", "Jinx", "Kog'Maw", "Miss Fortune", "Seraphine", "Kalista", "Smolder", "Hwei", "Ashe", "Twitch", "Lucian", "Karthus"},
	},
	"Veigar": {
		Counters:     []string{"Twitch", "Vel'Koz", "Katarina", "Quinn", "Zed", "Ziggs", "Kog'Maw", "Kassadin", "Xerath", "Akshan", "Nasus", "Syndra"},
		HardMatchups: []string{"Fizz", "Qiyana", "LeBlanc", "Zilean", "Ekko", "Naafiri", "Gwen", "Hwei", "Viktor", "Aurelion Sol", "Anivia", "Talon"},
	},
	"Velkoz": {
		Counters:     []string{"Elise", "Zac", "Taric", "Skarner", "Janna", "Zilean", "Blitzcrank", "Leona", "Sona", "Xerath", "Poppy", "Thresh"},
		HardMatchups: []string{"Karma", "Nautilus", "Sylas", "Rakan", "LeBlanc", "Pantheon", "Shaco", "Amumu", "Camille", "Alistar", "Nami", "Rell"},
	},
	"Vex": {
		Counters:     []string{"Gwen", "Annie", "Kayle", "Garen", "Taliyah", "Twisted Fate", "Cho'Gath", "Vladimir", "Morgana", "Viktor", "Hwei", "Veigar"},
		HardMatchups: []string{"Gangplank", "Anivia", "Nasus", "Tryndamere", "Xerath", "Pantheon", "Swain", "Brand", "Vel'Koz", "Kassadin", "Aurora", "Fizz"},
	},
	"Vi": {
		Counters:     []string{"Ivern", "Malphite", "Taliyah", "Wukong", "Aatrox", "Nasus", "Zaahen", "Lillia", "Poppy", "Shyvana", "Rek'Sai", "Zac"},
		HardMatchups: []string{"Sejuani", "Morgana", "Briar", "Xin Zhao", "Brand", "Sion", "Ambessa", "Olaf", "Zyra", "Quinn", "Talon", "Shen"},
	},
	"Viego": {
		Counters:     []string{"Sion", "Rek'Sai", "Rammus", "Zac", "Maokai", "Nidalee", "Morgana", "Yorick", "Ekko", "Kha'Zix", "Talon", "Udyr"},
		HardMatchups: []string{"Ivern", "Wukong", "Riven", "Evelynn", "Taliyah", "Darius", "Elise", "Zyra", "Nasus", "Aatrox", "Jax", "Shyvana"},
	},
	"Viktor": {
		Counters:     []string{"Singed", "Tryndamere", "Zilean", "Karthus", "Nunu & Willump", "Gragas", "Xerath", "Talon", "Morgana", "Kassadin", "Lee Sin", "Seraphine"},
		HardMatchups: []string{"Hwei", "Twisted Fate", "Vel'Koz", "Fizz", "Ziggs", "Akali", "Twitch", "Zoe", "Syndra", "Neeko", "Viego", "Ahri"},
	},
	"Vladimir": {
		Counters:     []string{"Cho'Gath", "Kog'Maw", "Malzahar", "Neeko", "Vel'Koz", "Heimerdinger", "Zilean", "Syndra", "Hwei", "Zoe", "Viktor", "Anivia"},
		HardMatchups: []string{"Singed", "Gragas", "Ekko", "Corki", "Kayle", "Xerath", "Kassadin", "Twisted Fate", "Yorick", "Taliyah", "Quinn", "Veigar"},
	},
	"Volibear": {
		Counters:     []string{"Quinn", "Lissandra", "Anivia", "Sejuani", "Kennen", "Jax", "Kayle", "Teemo", "Aurora", "Malzahar", "Heimerdinger", "Smolder"},
		HardMatchups: []string{"Singed", "Poppy", "Cassiopeia", "Naafiri", "Gnar", "Urgot", "Zed", "Akali", "Dr. Mundo", "Warwick", "Illaoi", "Gragas"},
	},
	"Warwick": {
		Counters:     []string{"Darius", "Ivern", "Nasus", "Dr. Mundo", "Quinn", "Rek'Sai", "Poppy", "Taliyah", "Ekko", "Shen", "Fiddlesticks", "Graves"},
		HardMatchups: []string{"Evelynn", "Wukong", "Gragas", "Sejuani", "Rammus", "Kindred", "Elise", "Shyvana", "Amumu", "Fizz", "Shaco", "Zac"},
	},
	"Xayah": {
		Counters:     []string{"Lux", "Veigar", "Tahm Kench", "Xerath", "Taliyah", "Hwei", "Brand", "Vel'Koz", "Swain", "Syndra", "Viktor", "Senna"},
		HardMatchups: []string{"Karthus", "Jinx", "Ziggs", "Vladimir", "Sivir", "Kog'Maw", "Draven", "Jhin", "Seraphine", "Cassiopeia", "Twitch", "Ashe"},
	},
	"Xerath": {
		Counters:     []string{"Singed", "Irelia", "Ekko", "Fiora", "Katarina", "Yasuo", "Malzahar", "Vayne", "Twisted Fate", "Zeri", "Aurelion Sol", "Viego"},
		HardMatchups: []string{"Kai'Sa", "Jayce", "Anivia", "Vladimir", "Kalista", "Nilah", "Yone", "Akali", "Swain", "Locke", "Viktor", "Tristana"},
	},
	"XinZhao": {
		Counters:     []string{"Zaahen", "Olaf", "Ivern", "Shen", "Nasus", "Mordekaiser", "Aatrox", "Shyvana", "Talon", "Taliyah", "Ambessa", "Warwick"},
		HardMatchups: []string{"Nidalee", "Trundle", "Zyra", "Rek'Sai", "Jax", "Master Yi", "Skarner", "Briar", "Darius", "Kha'Zix", "Cho'Gath", "Hecarim"},
	},
	"Yasuo": {
		Counters:     []string{"Zaahen", "Rumble", "Tryndamere", "Zilean", "Ambessa", "Illaoi", "Volibear", "Cho'Gath", "Kayle", "Darius", "Singed", "Kennen"},
		HardMatchups: []string{"Riven", "Quinn", "Malzahar", "Zac", "Brand", "Taliyah", "Fiora", "Lissandra", "Vladimir", "Gangplank", "Nasus", "Annie"},
	},
	"Yone": {
		Counters:     []string{"Zoe", "Rek'Sai", "Akshan", "Lissandra", "LeBlanc", "Camille", "Qiyana", "Draven", "Zed", "Twisted Fate", "Ahri", "Warwick"},
		HardMatchups: []string{"Zac", "Wukong", "Jax", "Akali", "Gangplank", "Vladimir", "Poppy", "Rammus", "Singed", "Azir", "Irelia", "Kennen"},
	},
	"Yorick": {
		Counters:     []string{"Yone", "Zed", "Lissandra", "Briar", "Malzahar", "Zilean", "Tryndamere", "Sett", "Singed", "Shen", "Fiora", "Warwick"},
		HardMatchups: []string{"Wukong", "Renekton", "Jax", "Yasuo", "Irelia", "Sylas", "Anivia", "Gangplank", "Skarner", "Ornn", "Quinn", "Kayle"},
	},
	"Yunara": {
		Counters:     []string{"Vel'Koz", "Olaf", "Tahm Kench", "Karthus", "Locke", "Twisted Fate", "Hwei", "Katarina", "Seraphine", "Veigar", "Swain", "Kog'Maw"},
		HardMatchups: []string{"Viktor", "Brand", "Zeri", "Xerath", "Heimerdinger", "Cassiopeia", "Tristana", "Jinx", "Lux", "Twitch", "Draven", "Nilah"},
	},
	"Yuumi": {
		Counters:     []string{"Sion", "Rell", "Taric", "Rakan", "Rammus", "Alistar", "Nautilus", "Thresh", "Leona", "Braum", "Shen", "Anivia"},
		HardMatchups: []string{"Amumu", "Nunu & Willump", "Zilean", "Renata Glasc", "Maokai", "Skarner", "Galio", "Blitzcrank", "Milio", "Lulu", "Pantheon", "Camille"},
	},
	"Zaahen": {
		Counters:     []string{"Ahri", "Zed", "Cassiopeia", "Teemo", "Singed", "Lissandra", "Kayn", "Poppy", "Quinn", "Garen", "Sejuani", "Viktor"},
		HardMatchups: []string{"Jax", "Heimerdinger", "Illaoi", "Aurora", "Camille", "Olaf", "Urgot", "Ryze", "Varus", "Zac", "Kayle", "Kennen"},
	},
	"Zac": {
		Counters:     []string{"Lillia", "Mordekaiser", "Zyra", "Aatrox", "Kayn", "Olaf", "Shyvana", "Nunu & Willump", "Evelynn", "Rek'Sai", "Wukong", "Skarner"},
		HardMatchups: []string{"Hecarim", "Graves", "Volibear", "Udyr", "Cho'Gath", "Bel'Veth", "Sejuani", "Shaco", "Briar", "Fiddlesticks", "Xin Zhao", "Ekko"},
	},
	"Zed": {
		Counters:     []string{"Singed", "Rumble", "Yorick", "Nasus", "Tahm Kench", "Lee Sin", "Malphite", "Tryndamere", "Neeko", "Kennen", "Garen", "Xin Zhao"},
		HardMatchups: []string{"Jax", "Dr. Mundo", "Ornn", "Zac", "Gwen", "Gragas", "Wukong", "Sett", "Karthus", "Cho'Gath", "Vladimir", "Fizz"},
	},
	"Zeri": {
		Counters:     []string{"Viego", "Karthus", "Vel'Koz", "Xayah", "Taliyah", "Nilah", "Zyra", "Viktor", "Tahm Kench", "Draven", "Brand", "Jinx"},
		HardMatchups: []string{"Lux", "Yasuo", "Yone", "Corki", "Tristana", "Seraphine", "Hwei", "Veigar", "Swain", "Syndra", "Aphelios", "Samira"},
	},
	"Ziggs": {
		Counters:     []string{"Zed", "Fiora", "Tahm Kench", "Ahri", "Master Yi", "Ryze", "Viego", "Malzahar", "Akali", "Aurelion Sol", "Karthus", "Nilah"},
		HardMatchups: []string{"Vel'Koz", "Hwei", "Kai'Sa", "Yasuo", "Vayne", "Zeri", "Xerath", "Lux", "Yone", "Kog'Maw", "Vladimir", "Ekko"},
	},
	"Zilean": {
		Counters:     []string{"Elise", "LeBlanc", "Janna", "Sona", "Zac", "Alistar", "Soraka", "Rakan", "Vi", "Leona", "Annie", "Shaco"},
		HardMatchups: []string{"Pyke", "Bard", "Maokai", "Milio", "Thresh", "Senna", "Taric", "Blitzcrank", "Amumu", "Sion", "Pantheon", "Rell"},
	},
	"Zoe": {
		Counters:     []string{"Zilean", "Kennen", "Heimerdinger", "Malzahar", "Nunu & Willump", "Sion", "Gangplank", "Naafiri", "Morgana", "Ekko", "Nasus", "Cho'Gath"},
		HardMatchups: []string{"Tryndamere", "Talon", "Swain", "Kog'Maw", "Kassadin", "Aurelion Sol", "Gragas", "Katarina", "Akshan", "Hwei", "Ahri", "Singed"},
	},
	"Zyra": {
		Counters:     []string{"Zoe", "Taric", "Anivia", "Blitzcrank", "Leona", "Elise", "Thresh", "Zilean", "Alistar", "Rell", "Sona", "Shaco"},
		HardMatchups: []string{"Nautilus", "Yuumi", "Vel'Koz", "Bard", "Poppy", "Karma", "Braum", "Galio", "Pyke", "Milio", "Jarvan IV", "Nami"},
	},
}
