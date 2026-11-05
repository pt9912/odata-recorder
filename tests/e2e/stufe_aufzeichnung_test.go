package e2e_test

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

const jsonKopf = "Content-Type"

func postJSON(body string) anfrage {
	return anfrage{methode: "POST", pfad: "/odata/Products", header: map[string][]string{jsonKopf: {"application/json"}}, body: body}
}

// TestE2EAufzeichnung ist die Stufe von LH-FA-02 (SPEC-037 Nr. 7): Rundlauf mit GET und zwei POST, leere
// Aufnahme, Speicher voll.
func TestE2EAufzeichnung(t *testing.T) {
	binaryPfad(t)
	t.Run("Rundlauf", rundlauf)
	t.Run("leere Aufnahme", func(t *testing.T) {
		for _, a := range []anfrage{{pfad: "/odata/Nichts"}, {methode: "POST", pfad: "/odata/Leer"}} {
			t.Run(a.methode+a.pfad, func(t *testing.T) {
				ref := neueReferenz(t)
				l := neuerLauf(t, ref)
				l.nimmAuf()
				w := l.frage(a)
				l.beendeAufnahme()
				ref.stoppe()
				l.importiere()
				rp := l.spieleAb()
				got := l.replay(rp, a, "leer")
				pruefeGleich(t, a.pfad, w, got)
				if got.status == 404 {
					t.Errorf("die leere Aufnahme ist nicht fehlgeschlagen, Replay: %d %q", got.status, got.body)
				}
			})
		}
	})
	t.Run("Speicher voll", speicherVollAufzeichnung)
}

// rundlauf: ein GET und zwei POST mit verschiedenen Bodies; der Replay fragt die POST in umgekehrter
// Reihenfolge, damit ein Match-Key ohne Body die Antworten vertauscht.
func rundlauf(t *testing.T) {
	ref := neueReferenz(t)
	l := neuerLauf(t, ref)
	l.nimmAuf()
	get := anfrage{pfad: "/odata/Products(1)"}
	pa, pb := postJSON(`{"Name":"A"}`), postJSON(`{"Name":"Bravo-lang"}`)
	wGet, wA, wB := l.frage(get), l.frage(pa), l.frage(pb)
	if string(wA.body) == string(wB.body) {
		t.Fatalf("die Fixture gibt für zwei Bodies dieselbe Antwort: %q", wA.body)
	}
	l.beendeAufnahme()
	ref.stoppe()
	l.importiere()
	rp := l.spieleAb()
	pruefeGleich(t, "POST B", wB, l.replay(rp, pb, "r1"))
	pruefeGleich(t, "POST A", wA, l.replay(rp, pa, "r1"))
	pruefeGleich(t, "GET", wGet, l.replay(rp, get, "r1"))
}

// fuelle füllt das tmpfs unter /voll mit einer Ballast-Datei bis ENOSPC; scheitert das Schreiben mit
// etwas anderem oder gar nicht, ist die Stufe rot (SPEC-037 Nr. 8).
func fuelle(t *testing.T) string {
	t.Helper()
	ballast := filepath.Join("/voll", "ballast")
	f, err := os.Create(ballast)
	if err != nil {
		t.Fatalf("Ballast-Datei: %v", err)
	}
	defer func() { _ = f.Close() }()
	block := make([]byte, 4096)
	for i := 0; i < 1024; i++ {
		if _, err := f.Write(block); err != nil {
			if !errors.Is(err, syscall.ENOSPC) {
				t.Fatalf("Ballast: %v, erwartet ENOSPC", err)
			}
			return ballast
		}
		_ = f.Sync()
	}
	t.Fatal("die Ballast-Datei erreicht ENOSPC nicht")
	return ballast
}

// volleKonfig legt DATA_DIR im tmpfs /voll an.
func volleKonfig(t *testing.T, l *lauf) {
	t.Helper()
	dir, err := os.MkdirTemp("/voll", "data")
	if err != nil {
		t.Fatalf("DATA_DIR unter /voll: %v", err)
	}
	l.dir = dir
}

// speicherVollAufzeichnung: bei vollem Speicher erhält der Client die Antwort unverändert, stderr trägt je
// Request `ereignis=ablage` mit `error_code`, und der Replay dieser Requests ist 404 mit ODR-E5001.
func speicherVollAufzeichnung(t *testing.T) {
	ref := neueReferenz(t)
	l := neuerLauf(t, ref)
	volleKonfig(t, l)
	p := l.nimmAuf()
	ballast := fuelle(t)
	arten := []anfrage{{pfad: "/odata/Products(1)"}, {pfad: "/odata/Products/$count"}}
	soll := []string{`{"@odata.context":"` + ref.url() + `/odata/$metadata#Products/$entity","Id":1,"Name":"Anvil"}`, "7"}
	for i, a := range arten {
		if w := holen(t, p.adresse, a); w.status != 200 || string(w.body) != soll[i] {
			t.Errorf("%s bei vollem Speicher: %d %q, erwartet 200 %q", a.pfad, w.status, w.body, soll[i])
		}
	}
	p.warteAufZeilen(len(arten), "ereignis=ablage", "error_code=ODR-")
	if err := os.Remove(ballast); err != nil {
		t.Fatal(err)
	}
	l.rec = p
	l.beendeAufnahme()
	ref.stoppe()
	l.importiere()
	rp := l.spieleAb()
	for _, a := range arten {
		pruefeFehler(t, "Replay von "+a.pfad, l.replay(rp, a, "voll"), 404, "ODR-E5001")
	}
}
