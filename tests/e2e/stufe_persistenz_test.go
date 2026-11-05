package e2e_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2EPersistenz ist die Stufe von LH-FA-03 (SPEC-037 Nr. 7): Neustart auf demselben DATA_DIR, Session
// ohne Request, beschädigte Datei.
func TestE2EPersistenz(t *testing.T) {
	binaryPfad(t)
	t.Run("Neustart", neustart)
	t.Run("leer", leereSession)
	t.Run("beschädigt", beschaedigteDatei)
}

// neustart: Aufnahme, SIGTERM, Übergang, zweiter Prozess auf demselben DATA_DIR; jeder der vier Requests
// der ersten Session ist im Replay gleich der Aufnahme.
func neustart(t *testing.T) {
	ref := neueReferenz(t)
	l := neuerLauf(t, ref)
	l.nimmAuf()
	for _, pfad := range []string{"/odata/Products", "/odata/Products(1)", "/odata/Products/$count"} {
		l.frage(anfrage{pfad: pfad})
	}
	l.beendeAufnahme()
	ref.stoppe()
	l.importiere()
	zweiter := l.spieleAb()
	if len(l.aufnahme) != 4 {
		t.Fatalf("%d Requests der ersten Session, erwartet 4", len(l.aufnahme))
	}
	for _, p := range l.aufnahme {
		pruefeGleich(t, p.a.pfad, p.w, l.replay(zweiter, p.a, "neustart"))
	}
}

// leereSession: ein Prozess ohne Request; `zeige` meldet exchanges=0, und der Replay dieser Session erfindet
// nichts. Der Import einer Session ohne `$metadata` ist ODR-E6004 (SPEC-127 Nr. 7); den Workspace liefert
// darum der Import einer zweiten Session auf demselben DATA_DIR, der Replay wählt die leere Session.
func leereSession(t *testing.T) {
	ref := neueReferenz(t)
	l := neuerLauf(t, ref)
	l.rec = starte(t, l.konfig("record"))
	l.rec.bereitOhneRequest()
	l.beendeAufnahme()
	leer := l.session
	out, errOut, code := uebergang(t, "zeige", l.dir, leer)
	if code != 0 || !strings.Contains(out, "exchanges=0") {
		t.Fatalf("zeige: Exit %d, stdout %q, stderr %q, erwartet exchanges=0", code, out, errOut)
	}
	l.nimmAuf()
	l.rec.beende()
	l.session = l.sessionen()[1]
	ref.stoppe()
	l.importiere()
	l.session = leer
	rp := l.spieleAb()
	for _, pfad := range []string{"/odata/Products(1)", "/odata/$metadata"} {
		pruefeFehler(t, "Replay "+pfad, l.replay(rp, anfrage{pfad: pfad}, "leer"), 404, "ODR-E5001")
	}
}

// beschaedigteDatei: DATA_DIR/odata-recorder.db ist Text oder trägt den SQLite-Kopf mit zerstörten
// Seiten; der Start endet mit Exit 1 und ODR-E3001, die Datei bleibt byte-gleich (SPEC-103 Nr. 6).
func beschaedigteDatei(t *testing.T) {
	ref := neueReferenz(t)
	l := neuerLauf(t, ref)
	l.nimmAuf()
	l.beendeAufnahme()
	echt, err := os.ReadFile(filepath.Join(l.dir, "odata-recorder.db"))
	if err != nil || len(echt) < 8192 {
		t.Fatalf("Vorlage der Ablage: %d Bytes, %v", len(echt), err)
	}
	zerstoert := bytes.Clone(echt)
	for i := 100; i < len(zerstoert); i++ {
		zerstoert[i] = 0xFF
	}
	faelle := []struct {
		name  string
		bytes []byte
	}{
		{"NOTADB", []byte(strings.Repeat("NOTADB kein SQLite, nur Text. ", 40))},
		{"CORRUPT", zerstoert},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			dir := t.TempDir()
			datei := filepath.Join(dir, "odata-recorder.db")
			if err := os.WriteFile(datei, f.bytes, 0o600); err != nil {
				t.Fatal(err)
			}
			adresse := freierPort(t)
			p := startRoh(t, adresse, konfig{modus: "record", dir: dir, upstream: ref.url()}.env(adresse))
			p.warteEnde()
			if p.exit != 1 || !strings.Contains(p.fehl.String(), "ODR-E3001") {
				t.Errorf("Exit %d, erwartet 1 mit ODR-E3001\nstderr: %s", p.exit, p.fehl)
			}
			if danach, err := os.ReadFile(datei); err != nil || !bytes.Equal(danach, f.bytes) {
				t.Errorf("die Datei ist nicht byte-gleich (%d statt %d Bytes, %v)", len(danach), len(f.bytes), err)
			}
		})
	}
}
