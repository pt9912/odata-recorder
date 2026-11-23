package support_test

import (
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// orakelArten ist die Tabelle von SPEC-234 Nr. 2 aus dem Wortlaut: Art, dann je Feld der Spalten
// service_id, session_id, workspace_id, import_id, error_code, ob es Pflicht ist.
func orakelArten() map[string][5]bool {
	return map[string][5]bool{
		"session_geloescht":      {true, true, false, false, false},
		"session_aufbewahrung":   {true, true, false, false, false},
		"workspace_geloescht":    {false, false, true, false, false},
		"import_fehlgeschlagen":  {true, false, false, true, true},
		"admin_abgewiesen":       {false, false, false, false, true},
		"session_angelegt":       {true, true, false, false, false},
		"session_beendet":        {true, true, false, false, false},
		"import_veroeffentlicht": {true, false, false, true, false},
	}
}

func orakelSpalten() [5]support.Feld {
	return [5]support.Feld{"service_id", "session_id", "workspace_id", "import_id", "error_code"}
}

// A01, Y5, S15, B14 und IM23: Schleife über 8 Arten × 5 Felder gegen das Orakel (die vier Zeilen von SPEC-234 Nr. 2, die
// fünfte, `admin_abgewiesen`, mit allein `error_code` als Pflicht, die sechste und die siebente, `session_angelegt` und
// `session_beendet`, mit `service_id` und `session_id` als Pflicht, die achte, `import_veroeffentlicht`, mit `service_id` und
// `import_id`), Zählprüfung 8 und 40; die fünfte bis achte Art stehen an ihrer Stelle.
func TestArtenGleichTabelle(t *testing.T) {
	orakel := orakelArten()
	arten := support.Arten()
	if len(arten) != 8 || len(orakel) != 8 {
		t.Fatalf("Arten() hat %d Einträge, das Orakel %d, erwartet je 8", len(arten), len(orakel))
	}
	if arten[4].Art != "admin_abgewiesen" || arten[5].Art != "session_angelegt" || arten[6].Art != "session_beendet" || arten[7].Art != "import_veroeffentlicht" {
		t.Errorf("die fünfte Zeile ist %q, die sechste %q, die siebente %q und die achte %q, erwartet admin_abgewiesen, session_angelegt, session_beendet und import_veroeffentlicht",
			arten[4].Art, arten[5].Art, arten[6].Art, arten[7].Art)
	}
	zellen := 0
	for _, s := range arten {
		soll, ok := orakel[string(s.Art)]
		if !ok {
			t.Errorf("Art %q steht nicht in der Tabelle", s.Art)
			continue
		}
		for i, f := range orakelSpalten() {
			zellen++
			if got := slices.Contains(s.Pflicht, f); got != soll[i] {
				t.Errorf("Art %q Feld %q: Pflicht %v, erwartet %v", s.Art, f, got, soll[i])
			}
		}
		// Die Pflichtfelder stehen ohne Doppelung in der Folge der Spalten.
		var folge []support.Feld
		for i, f := range orakelSpalten() {
			if soll[i] {
				folge = append(folge, f)
			}
		}
		if !slices.Equal(s.Pflicht, folge) {
			t.Errorf("Art %q: Pflicht %v, erwartet in der Folge der Spalten %v", s.Art, s.Pflicht, folge)
		}
	}
	if zellen != 40 {
		t.Errorf("geprüfte Zellen: %d, erwartet 40", zellen)
	}
}

// Der Text jeder Art ist der der Tabelle, Byte für Byte, und Arten() führt genau diese acht.
func TestArtTexte(t *testing.T) {
	soll := map[support.AuditArt]string{
		support.ArtSessionGeloescht:     "session_geloescht",
		support.ArtSessionAufbewahrung:  "session_aufbewahrung",
		support.ArtWorkspaceGeloescht:   "workspace_geloescht",
		support.ArtImportFehlgeschlagen: "import_fehlgeschlagen",
		support.ArtAdminAbgewiesen:      "admin_abgewiesen",
		support.ArtSessionAngelegt:      "session_angelegt",
		support.ArtSessionBeendet:       "session_beendet",

		support.ArtImportVeroeffentlicht: "import_veroeffentlicht",
	}
	if len(soll) != 8 {
		t.Fatalf("Orakel hat %d Arten, erwartet 8", len(soll))
	}
	for art, text := range soll {
		if string(art) != text {
			t.Errorf("Art %q, erwartet %q", art, text)
		}
		if !slices.ContainsFunc(support.Arten(), func(s support.ArtSpec) bool { return s.Art == art }) {
			t.Errorf("Arten() führt %q nicht", art)
		}
	}
}

// Jeder Aufruf liefert eine neue Kopie: eine Änderung der Antwort ändert die Tabelle nicht.
func TestArtenLiefertKopie(t *testing.T) {
	a := support.Arten()
	pflichtfelder := 0
	for i := range a {
		a[i].Art = "verändert"
		for j := range a[i].Pflicht {
			a[i].Pflicht[j] = "verändert"
			pflichtfelder++
		}
	}
	if len(a) != 8 || pflichtfelder != 2+2+1+3+1+2+2+2 {
		t.Fatalf("%d Arten, %d Pflichtfelder, erwartet 8 und 15", len(a), pflichtfelder)
	}
	b := support.Arten()
	for i, s := range b {
		if _, bekannt := orakelArten()[string(s.Art)]; !bekannt {
			t.Errorf("Arten() teilt Zustand zwischen Aufrufen: %v", s)
		}
		for j, f := range s.Pflicht {
			if f == "verändert" {
				t.Errorf("Arten()[%d].Pflicht[%d] teilt Zustand", i, j)
			}
		}
	}
}
