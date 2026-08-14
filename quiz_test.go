package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func testChamp(id, name, title, pName, pDesc string, spells ...[6]string) champDetail {
	d := champDetail{ID: id, Key: "1", Name: name, Title: title}
	d.Image.Full = id + ".png"
	d.Passive.Name = pName
	d.Passive.Description = pDesc
	d.Passive.Image.Full = id + "_P.png"
	for _, sp := range spells {
		var s struct {
			Name         string `json:"name"`
			Description  string `json:"description"`
			CooldownBurn string `json:"cooldownBurn"`
			CostBurn     string `json:"costBurn"`
			RangeBurn    string `json:"rangeBurn"`
			Image        struct {
				Full string `json:"full"`
			} `json:"image"`
		}
		s.Name, s.CooldownBurn, s.CostBurn, s.RangeBurn, s.Image.Full, s.Description = sp[0], sp[1], sp[2], sp[3], sp[4], sp[5]
		d.Spells = append(d.Spells, s)
	}
	return d
}

func testPool() []champDetail {
	return []champDetail{
		testChamp("Ahri", "Ahri", "la goupil à neuf queues", "Essence du vol", "Après 9 sbires, Ahri soigne.",
			[6]string{"Orbe de séduction", "7", "70/75/80/85/90", "880", "AhriQ.png", "Ahri envoie une orbe qui inflige des dégâts magiques puis revient."},
			[6]string{"Lucioles", "9/8/7/6/5", "50", "800", "AhriW.png", "Ahri invoque des lucioles qui pourchassent les ennemis proches."},
			[6]string{"Charme", "12/11/10/9/8", "70", "600", "AhriE.png", "Ahri envoie un baiser qui charme le premier ennemi touché."},
			[6]string{"Esprit renard", "130/105/80", "100", "450", "AhriR.png", "Ahri se précipite et tire des boules d'énergie."},
		),
		testChamp("Zed", "Zed", "le maître des ombres", "Mépris des faibles", "Les attaques contre les cibles basses PV infligent des dégâts bonus.",
			[6]string{"Shuriken tranchant", "6/5.75/5.5/5.25/5", "75", "900", "ZedQ.png", "Zed et ses ombres lancent des shurikens."},
			[6]string{"Lame vivante", "20/18/16/14/12", "40", "700", "ZedW.png", "Zed envoie une ombre qui peut être réactivée pour échanger de place."},
			[6]string{"Taillade de l'ombre", "5/4.5/4/3.5/3", "50", "290", "ZedE.png", "Zed et ses ombres tailladent les ennemis autour d'eux."},
			[6]string{"Marque de la mort", "120/110/100", "0", "625", "ZedR.png", "Zed devient incible et marque un ennemi. La marque explose."},
		),
		testChamp("Anivia", "Anivia", "la cryophénix", "Réincarnation", "Cette compétence a un délai de récupération de 4 minutes.",
			[6]string{"Souffle glacial", "12/11/10/9/8", "80/90/100/110/120", "1100", "AniviaQ.png", "Anivia lance une orbe de glace qu'elle peut faire exploser."},
			[6]string{"Cristallisation", "25", "70", "1000", "AniviaW.png", "Anivia invoque un mur de glace."},
			[6]string{"Congélation", "4", "50/60/70/80/90", "650", "AniviaE.png", "Anivia inflige plus de dégâts aux cibles étourdies ou gelées."},
			[6]string{"Tempête glaciale", "80/70/60", "75", "750", "AniviaR.png", "Anivia invoque une tempête qui ralentit et inflige des dégâts."},
		),
		testChamp("Malphite", "Malphite", "le fragment du monolithe", "Saccage granitique", "Malphite gagne un bouclier de pierre.",
			[6]string{"Éclat sismique", "8", "70/75/80/85/90", "625", "MalphiteQ.png", "Malphite lance un éclat de terre qui ralentit."},
			[6]string{"Frappe du tonnerre", "12/11/10/9/8", "30/35/40/45/50", "400", "MalphiteW.png", "Malphite renforce ses attaques."},
			[6]string{"Choc tellurique", "7", "50/55/60/65/70", "400", "MalphiteE.png", "Malphite frappe le sol autour de lui."},
			[6]string{"Force inarrêtable", "130/105/80", "100", "1000", "MalphiteR.png", "Malphite charge et projette les ennemis en l'air."},
		),
		testChamp("Caitlyn", "Caitlyn", "le shérif de Piltover", "Tir dans la tête", "Tous les quelques attaques, Caitlyn tire un tir dans la tête.",
			[6]string{"Pacificateur de Piltover", "10/9/8/7/6", "50/60/70/80/90", "1250", "CaitlynQ.png", "Caitlyn tire un projectile qui traverse les ennemis."},
			[6]string{"Piège de Yordle", "30/24/18/12/6", "20", "800", "CaitlynW.png", "Caitlyn pose un piège qui révèle et enracine."},
			[6]string{"Tir calibré", "16/14.5/13/11.5/10", "75", "700", "CaitlynE.png", "Caitlyn tire un filet et recule."},
			[6]string{" visée", "90/75/60", "100", "3500", "CaitlynR.png", "Caitlyn vise un ennemi lointain pour un tir unique."},
		),
		testChamp("Leblanc", "LeBlanc", "la menteuse", "Image miroir", "Quand LeBlanc tombe bas, elle se camoufle brièvement.",
			[6]string{"Sphère de malveillance", "7/6.5/6/5.5/5", "50", "700", "LeblancQ.png", "LeBlanc lance une sphère qui marque puis explose."},
			[6]string{"Distorsion", "18/16/14/12/10", "60/70/80/90/100", "600", "LeblancW.png", "LeBlanc bondit puis peut revenir."},
			[6]string{"Chaînes éthérées", "14/13.25/12.5/11.75/11", "50", "950", "LeblancE.png", "LeBlanc lance une chaîne qui ligote si elle reste proche."},
			[6]string{"Mimétisme", "50/40/30", "0", "600", "LeblancR.png", "LeBlanc réplique son dernier sort."},
		),
	}
}

func TestCDDistractors(t *testing.T) {
	d := cdDistractors(7, []float64{9, 12, 130})
	if len(d) < 3 {
		t.Fatalf("pas assez de distracteurs: %v", d)
	}
	for _, v := range d {
		if quizNear(v, 7) || v <= 0 {
			t.Fatalf("distracteur invalide %v", v)
		}
	}
	d2 := cdDistractors(120, []float64{100, 80})
	if len(d2) < 3 {
		t.Fatalf("ulti: %v", d2)
	}
}

func TestFinishQuizChoices(t *testing.T) {
	got := finishQuizChoices(quizChoice{Label: "7s"}, []quizChoice{
		{Label: "8s"}, {Label: "9s"}, {Label: "12s"}, {Label: "7s"}, {Label: ""},
	}, 4)
	if len(got) != 4 {
		t.Fatalf("len=%d", len(got))
	}
	seen, ids, has := map[string]bool{}, map[string]bool{}, false
	for _, c := range got {
		if c.ID == "" || ids[c.ID] {
			t.Fatalf("id: %+v", c)
		}
		ids[c.ID] = true
		if seen[c.Label] {
			t.Fatalf("doublon %s", c.Label)
		}
		seen[c.Label] = true
		if c.Label == "7s" {
			has = true
		}
	}
	if !has {
		t.Fatal("bonne réponse absente")
	}
}

func TestScrubQuizText(t *testing.T) {
	d := testPool()[0]
	s := scrubQuizText(d, "Ahri envoie une orbe. Ahri se précipite ensuite.")
	if strings.Contains(s, "Ahri") {
		t.Fatalf("fuite du nom: %s", s)
	}
	if !strings.Contains(s, "ce champion") {
		t.Fatalf("remplacement manquant: %s", s)
	}
}

func TestNameLeaksChamp(t *testing.T) {
	d := champDetail{ID: "Caitlyn", Name: "Caitlyn"}
	if !nameLeaksChamp(d, "Pacificateur de Caitlyn") {
		t.Fatal("devrait fuiter")
	}
	if nameLeaksChamp(d, "Shuriken tranchant") {
		t.Fatal("faux positif")
	}
}

func TestBuildQuizFromPool(t *testing.T) {
	pool := testPool()
	seen := map[string]bool{}
	okN := 0
	var last quizQuestion
	for i := 0; i < 40; i++ {
		q, p, ok := buildQuizFromPool(pool, seen)
		if !ok {
			seen = map[string]bool{}
			continue
		}
		okN++
		last = q
		if q.Prompt == "" || len(q.Choices) != 4 || q.Kind == "" {
			t.Fatalf("question incomplète: %+v", q)
		}
		if p.correct == "" || p.explain == "" || p.answer == "" {
			t.Fatalf("pending incomplet: %+v", p)
		}
		found := false
		for _, c := range q.Choices {
			if c.ID == p.correct {
				found = true
				if c.Label != p.answer {
					t.Fatalf("label %q != %q", c.Label, p.answer)
				}
			}
		}
		if !found {
			t.Fatalf("id correct absent des choix")
		}
		raw, _ := json.Marshal(q)
		var m map[string]any
		_ = json.Unmarshal(raw, &m)
		if _, has := m["explain"]; has {
			t.Fatal("la question ne doit pas exposer explain")
		}
		if _, has := m["correct"]; has {
			t.Fatal("la question ne doit pas exposer correct")
		}
		if strings.Contains(string(raw), p.explain) && len(p.explain) > 24 {
			t.Fatalf("explain fuité dans le JSON: %s", raw)
		}
	}
	if okN < 12 {
		t.Fatalf("trop peu de questions générées: %d", okN)
	}
	_ = last
}

func TestQuizGeneratorsHardPool(t *testing.T) {
	pool := testPool()
	gens := []struct {
		n  string
		fn func([]champDetail) *quizBuilt
	}{
		{"cdRank", genCDRank},
		{"curve", genCDCurveWho},
		{"which", genWhichSpellCD},
		{"ult", genUltCompare},
		{"name", genSpellName},
		{"who", genSpellWho},
		{"desc", genDescWho},
		{"pcd", genPassiveCD},
		{"pwho", genPassiveWho},
		{"cost", genCostRank},
	}
	for _, g := range gens {
		var b *quizBuilt
		for i := 0; i < 20 && b == nil; i++ {
			b = g.fn(pool)
		}
		if b == nil {
			t.Errorf("%s: aucune question", g.n)
			continue
		}
		if b.prompt == "" || b.correct.Label == "" || len(b.pool) < 3 {
			t.Errorf("%s incomplet: %+v", g.n, b)
		}
		_, _, ok := wrapQuiz(b, 0, 0, 0)
		if !ok {
			t.Errorf("%s: wrap a échoué", g.n)
		}
	}
}

func TestQuizAniviaPassiveFromText(t *testing.T) {
	pool := testPool()
	var b *quizBuilt
	for i := 0; i < 30 && b == nil; i++ {
		b = genPassiveCD([]champDetail{pool[2]}) // Anivia, CD 4 min = 240s
	}
	if b == nil {
		t.Fatal("passif Anivia")
	}
	if b.correct.Label != "240s" {
		t.Fatalf("attendu 240s, obtenu %s (%s)", b.correct.Label, b.explain)
	}
}
