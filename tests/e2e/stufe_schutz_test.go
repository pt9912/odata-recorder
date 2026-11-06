package e2e_test

import (
	"os"
	"strings"
	"testing"
)

// secrets sind die drei Geheimnisse der Header: A (Authorization) erreicht den Referenzdienst, B (Cookie)
// ist ein Request-Header ohne Ziel, C (Set-Cookie) erreicht den Client.
func secrets() []string { return []string{geheimA, geheimB, geheimC} }

// schutzLauf ist ein Lauf mit dem Kontrollwert als SERVICE_ID.
func schutzLauf(t *testing.T, ref *referenz) *lauf {
	t.Helper()
	l := neuerLauf(t, ref)
	l.dienst = kontrollDienst
	return l
}

// TestE2ESchutz ist die Stufe von LH-FA-14 (SPEC-037 Nr. 7).
func TestE2ESchutz(t *testing.T) {
	binaryPfad(t)
	t.Run("Geheimnisse", geheimnisse)
	t.Run("Kontrollwert", kontrollwert)
	t.Run("Nutzlast", nutzlast)
	t.Run("JSON-Felder", jsonFelder)
	t.Run("unredigierbar", unredigierbar)
	t.Run("Speicher voll", speicherVollSchutz)
}

// ort ist eine Stelle der Suche mit ihren Funden.
type ort struct {
	name  string
	funde []fund
}

// geheimnisse: A erreicht den Referenzdienst, C den Client, und keines der drei steht in einer Datei vor
// oder nach dem Herunterfahren, in stdout oder in stderr (3 Geheimnisse mal 4 Orte).
func geheimnisse(t *testing.T) {
	ref := neueReferenz(t)
	l := schutzLauf(t, ref)
	p := l.nimmAuf()
	pflanzeKontrolle(t, l.dir)
	l.frage(anfrage{pfad: "/odata/Products"})
	l.frage(anfrage{pfad: "/odata/Products(1)"})
	for i, x := range l.aufnahme {
		if !strings.Contains(strings.Join(x.w.header.Values("Set-Cookie"), ";"), geheimC) {
			t.Errorf("Antwort %d (%s): Set-Cookie erreicht den Client ohne GEHEIM-C", i, x.a.pfad)
		}
	}
	if got := ref.mitGeheimnis.Load(); got != int64(len(l.aufnahme)) {
		t.Errorf("Requests mit GEHEIM-A am Referenzdienst %d, erwartet %d", got, len(l.aufnahme))
	}
	ref.stoppe()
	holen(t, p.adresse, anfrage{pfad: "/odata/Products(1)"})
	p.warteAufZeile("ereignis=upstream", "service_id="+kontrollDienst)
	_, vorher := durchsuche(t, l.dir, append(secrets(), werteKontrolle()...))
	l.beendeAufnahme()
	dateien, nachher := durchsuche(t, l.dir, append(secrets(), werteKontrolle()...))
	if len(dateien) < 7 {
		t.Errorf("die Suche sieht nur %d Dateien: %v", len(dateien), dateien)
	}
	strom := sucheStroeme(p, append(secrets(), werteKontrolle()...))
	pruefeKontrolle(t, "vor dem Herunterfahren", vorher, strom)
	pruefeKontrolle(t, "nach dem Herunterfahren", nachher, strom)
	orte := []ort{{"Dateien vor dem Herunterfahren", vorher}, {"Dateien nach dem Herunterfahren", nachher},
		{"stdout", filtere(strom, "stdout")}, {"stderr", filtere(strom, "stderr")}}
	for _, g := range secrets() {
		for _, o := range orte {
			if hat(o.funde, "", g) {
				t.Errorf("%s steht in %s", g, o.name)
			}
		}
	}
}

func filtere(funde []fund, ort string) (out []fund) {
	for _, f := range funde {
		if f.ort == ort {
			out = append(out, f)
		}
	}
	return out
}

// kontrollwert: die Suche findet, was da ist, in Dateien und auf stderr.
func kontrollwert(t *testing.T) {
	ref := neueReferenz(t)
	l := schutzLauf(t, ref)
	p := l.nimmAuf()
	pflanzeKontrolle(t, l.dir)
	ref.stoppe()
	holen(t, p.adresse, anfrage{pfad: "/odata/Products(1)"})
	p.warteAufZeile("service_id=" + kontrollDienst)
	l.beendeAufnahme()
	dateien, funde := durchsuche(t, l.dir, werteKontrolle())
	pruefeKontrolle(t, "Kontrollwert", funde, sucheStroeme(p, werteKontrolle()))
	if len(dateien) < len(kontrollDateien()) {
		t.Errorf("die Suche liest %d Dateien: %v", len(dateien), dateien)
	}
}

// nutzlast: ein JSON-Body ohne sensible Felder bleibt für den Replay erhalten.
func nutzlast(t *testing.T) {
	ref := neueReferenz(t)
	l := schutzLauf(t, ref)
	l.nimmAuf()
	a := postJSON(`{"Name":"Ambos","Preis":12,"Tags":["a","b"]}`)
	w := l.frage(a)
	l.beendeAufnahme()
	ref.stoppe()
	l.importiere()
	pruefeGleich(t, "POST ohne sensible Felder", w, l.replay(l.spieleAb(), a, "nutz"))
}

// unredigierbar: ein ungültiger JSON-Body lässt den Status des Clients unverändert, sein Marker steht
// nirgends in DATA_DIR, stderr trägt ODR-E2401 mit richtung=request.
func unredigierbar(t *testing.T) {
	const marker = "MARKER-UNREDIGIERBAR-7e3b"
	ref := neueReferenz(t)
	l := schutzLauf(t, ref)
	p := l.nimmAuf()
	if w := l.frage(postJSON(`{"Name": "` + marker)); w.status != 201 {
		t.Errorf("Status des Clients %d, erwartet 201 (der Referenzdienst antwortet 201)", w.status)
	}
	p.warteAufZeile("ereignis=capture", "error_code=ODR-E2401", "richtung=request")
	_, vorher := durchsuche(t, l.dir, []string{marker})
	l.beendeAufnahme()
	_, nachher := durchsuche(t, l.dir, []string{marker})
	for name, funde := range map[string][]fund{"vor dem Herunterfahren": vorher, "nach dem Herunterfahren": nachher, "Ströme": sucheStroeme(p, []string{marker})} {
		if len(funde) > 0 {
			t.Errorf("Marker %s: %v (%s)", marker, funde, name)
		}
	}
}

// speicherVollSchutz: bei vollem Speicher bleibt der Status; drei Geheimnisse (zwei Authorization, ein
// `password` im Body) stehen weder vor noch nach dem Herunterfahren in einer Datei noch in stdout und stderr;
// der Spurwert steht nach dem Herunterfahren in keiner Datei.
func speicherVollSchutz(t *testing.T) {
	const marker = "MARKER-VOLL-5a9c"
	ref := neueReferenz(t)
	l := schutzLauf(t, ref)
	volleKonfig(t, l)
	p := l.nimmAuf()
	ballast := fuelle(t)
	soll := ref.seite("1", "Anvil", true)
	const geheimD, passwort = "GEHEIM-D-2c71", "GEHEIM-E-8f40"
	w := holen(t, p.adresse, anfrage{pfad: "/odata/Products", header: map[string][]string{"X-Test-Marker": {marker}}})
	if w.status != 200 || string(w.body) != soll {
		t.Errorf("Request 1 bei vollem Speicher: %d %q, erwartet 200 mit dem vollen Body", w.status, w.body)
	}
	post := postJSON(`{"Name":"x","password":"` + passwort + `"}`)
	post.header["X-Test-Marker"] = []string{marker}
	post.header["Authorization"] = []string{"Bearer " + geheimD}
	if w := holen(t, p.adresse, post); w.status != 201 {
		t.Errorf("Request 2 bei vollem Speicher: %d %q, erwartet 201", w.status, w.body)
	}
	geheimnisse := []string{geheimA, geheimD, passwort}
	p.warteAufZeilen(2, "ereignis=ablage")
	_, vorher := durchsuche(t, l.dir, append([]string{marker}, geheimnisse...))
	for _, f := range vorher {
		if f.wert == marker {
			t.Logf("Spurwert vor dem Herunterfahren: %v", f)
		} else {
			t.Errorf("Geheimnis %s vor dem Herunterfahren in %s", f.wert, f.ort)
		}
	}
	if err := os.Remove(ballast); err != nil {
		t.Fatal(err)
	}
	l.rec = p
	l.beendeAufnahme()
	_, nachher := durchsuche(t, l.dir, append([]string{marker}, geheimnisse...))
	for _, f := range nachher {
		t.Errorf("%s liegt nach dem Herunterfahren in %s", f.wert, f.ort)
	}
	for _, f := range sucheStroeme(p, geheimnisse) {
		t.Errorf("Geheimnis %s in %s", f.wert, f.ort)
	}
}

// jsonFelder: jedes der 7 Felder der Standardmenge von REDACT_JSON_PATHS trägt in Request- und
// Antwort-Body ein Geheimnis; bei nicht vollem Speicher steht keines der 14 Werte in einer Datei vor oder
// nach dem Herunterfahren, in stdout oder stderr, ein Wert ohne Redaktion (Titel) steht im Blob.
func jsonFelder(t *testing.T) {
	ref := neueReferenz(t)
	l := schutzLauf(t, ref)
	p := l.nimmAuf()
	post := postJSON(feldBody(false))
	if w := l.frage(post); w.status != 201 {
		t.Errorf("POST mit sensiblen Feldern: %d %q, erwartet 201", w.status, w.body)
	}
	konto := l.frage(anfrage{pfad: "/odata/Konto"})
	if string(konto.body) != feldBody(true) {
		t.Errorf("der Client erhält %q, erwartet den Body des Referenzdienstes", konto.body)
	}
	if got := ref.felderErreicht.Load(); got != 1 {
		t.Errorf("POST mit allen sensiblen Werten am Referenzdienst: %d, erwartet 1", got)
	}
	var werte []string
	for _, f := range sensibleFelder() {
		werte = append(werte, f.anfrage, f.antwort)
	}
	suche := append([]string{kontrollBody}, werte...)
	_, vorher := durchsuche(t, l.dir, suche)
	l.beendeAufnahme()
	_, nachher := durchsuche(t, l.dir, suche)
	orte := []ort{{"Dateien vor dem Herunterfahren", vorher}, {"Dateien nach dem Herunterfahren", nachher},
		{"stdout", filtere(sucheStroeme(p, suche), "stdout")}, {"stderr", filtere(sucheStroeme(p, suche), "stderr")}}
	for _, f := range sensibleFelder() {
		for _, w := range []string{f.anfrage, f.antwort} {
			for _, o := range orte {
				if hat(o.funde, "", w) {
					t.Errorf("Feld %s: %s steht in %s", f.name, w, o.name)
				}
			}
		}
	}
	if !hat(nachher, "blobs", kontrollBody) {
		t.Errorf("der Kontrollwert %s steht in keinem Blob: die Suche beweist nichts", kontrollBody)
	}
}
