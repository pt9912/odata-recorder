package maintenance_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestAbgleich: der Abgleich entfernt genau die Dateien ohne Referenz, aufsteigend, nach einer
// erneuten Lesung der Referenzmenge; Startup liest dreimal, CleanupBlobs zweimal (W12).
func TestAbgleich(t *testing.T) {
	bestand := func(u *welt) {
		u.speicher.fest = []string{r("b"), r("d")}
		u.blobs.lege(r("a"), r("b"), r("c"), r("d"), r("e"))
	}
	// lesungen ist die Zahl der Lesungen der Referenzmenge und zugleich die Nummer der letzten.
	faelle := []struct {
		name     string
		ruf      func(u *welt) (int, error)
		lesungen int
	}{
		{"Startup", func(u *welt) (int, error) { b, err := u.startup(context.Background()); return b.BlobsEntfernt, err }, 3},
		{"CleanupBlobs", func(u *welt) (int, error) { b, err := u.cleanup(context.Background()); return b.Entfernt, err }, 2},
	}
	for _, f := range faelle {
		u := neu(t)
		bestand(u)
		n, err := f.ruf(u)
		soll := []string{ev("Remove", r("a")), ev("Remove", r("c")), ev("Remove", r("e"))}
		if err != nil || n != 3 || !slices.Equal(u.p.mit("Remove"), soll) {
			t.Errorf("%s: %v, %d entfernt, %v", f.name, err, n, u.p.mit("Remove"))
		}
		if got := u.blobs.refs(); !slices.Equal(got, []string{r("b"), r("d")}) {
			t.Errorf("%s: Dateien %v", f.name, got)
		}
		if got := u.st.anzahl("ReferencedBlobs"); got != f.lesungen {
			t.Errorf("%s: %d Lesungen, erwartet %d", f.name, got, f.lesungen)
		}

		v := neu(t)
		bestand(v)
		v.speicher.zusatz[f.lesungen] = []string{r("c")}
		n, err = f.ruf(v)
		soll = []string{ev("Remove", r("a")), ev("Remove", r("e"))}
		if err != nil || n != 2 || !slices.Equal(v.p.mit("Remove"), soll) || !v.blobs.hat(r("c")) {
			t.Errorf("%s, letzte Lesung mit rc: %v, %d entfernt, %v", f.name, err, n, v.p.mit("Remove"))
		}

		// Die Menge schrumpft: die erste Lesung des Abgleichs nennt rc, die letzte nicht (W12).
		s := neu(t)
		bestand(s)
		s.speicher.zusatz[f.lesungen-1] = []string{r("c")}
		n, err = f.ruf(s)
		soll = []string{ev("Remove", r("a")), ev("Remove", r("e"))}
		if err != nil || n != 2 || !slices.Equal(s.p.mit("Remove"), soll) || !s.blobs.hat(r("c")) {
			t.Errorf("%s, erste Lesung mit rc: %v, %d entfernt, %v", f.name, err, n, s.p.mit("Remove"))
		}
	}
}

// TestPruefenBlobs: eine fehlende oder beschädigte referenzierte Datei wird gemeldet und nicht
// entfernt, der Start läuft weiter (W13).
func TestPruefenBlobs(t *testing.T) {
	u := neu(t)
	u.speicher.fest = []string{r("b"), r("d"), r("f"), r("h")}
	u.blobs.lege(r("b"), r("f"), r("h"), r("u"))
	u.st.scheitere("Verify", 2, fmt.Errorf("x: %w", outbound.ErrBlobFehlt))
	u.st.scheitere("Verify", 3, fmt.Errorf("y: %w", outbound.ErrBlobBeschaedigt))
	b, err := u.startup(context.Background())
	soll := leerBericht()
	soll.BlobsFehlend, soll.BlobsBeschaedigt, soll.BlobsEntfernt = []string{r("d")}, []string{r("f")}, 1
	if err != nil || !reflect.DeepEqual(b, soll) {
		t.Errorf("%v %+v, erwartet %+v", err, b, soll)
	}
	verifiziert := []string{ev("Verify", r("b")), ev("Verify", r("d")), ev("Verify", r("f")), ev("Verify", r("h"))}
	if got := u.p.mit("Verify"); !slices.Equal(got, verifiziert) {
		t.Errorf("Verify %v, erwartet %v", got, verifiziert)
	}
	if got := u.p.mit("Remove"); !slices.Equal(got, []string{ev("Remove", r("u"))}) || !u.blobs.hat(r("f")) {
		t.Errorf("Remove %v, Datei rf vorhanden %v", got, u.blobs.hat(r("f")))
	}
	log := u.log.alle()
	if len(log) != 2 {
		t.Fatalf("Protokoll %v", log)
	}
	for i, ereignis := range []string{"blob_fehlt", "blob_beschaedigt"} {
		ref := []string{r("d"), r("f")}[i]
		if log[i].stufe != slog.LevelError || log[i].attrs["ereignis"] != ereignis || log[i].attrs["error_code"] != "ODR-E3001" ||
			log[i].attrs["blob"] != ref || len(log[i].attrs) != 3 {
			t.Errorf("Eintrag %d: %+v", i, log[i])
		}
	}
}

// TestRemoveFehler: der erste Fehler von Remove beendet den Abgleich, die Entfernten sind gezählt
// (W17).
func TestRemoveFehler(t *testing.T) {
	bestand := func() *welt {
		u := neu(t)
		u.blobs.lege(r("a"), r("b"), r("c"))
		u.st.scheitere("Remove", 2, errDisk)
		return u
	}
	pruefe := func(name string, u *welt, err error, entfernt int) {
		t.Helper()
		if code, status := codeVon(t, err); code != "ODR-E9001" || status != 500 || !errors.Is(err, errDisk) {
			t.Errorf("%s: %v", name, err)
		}
		if entfernt != 1 || !slices.Equal(u.p.mit("Remove"), []string{ev("Remove", r("a")), ev("Remove", r("b"))}) {
			t.Errorf("%s: %d entfernt, %v", name, entfernt, u.p.mit("Remove"))
		}
		if got := u.blobs.refs(); !slices.Equal(got, []string{r("b"), r("c")}) {
			t.Errorf("%s: Dateien %v", name, got)
		}
		if u.st.anzahl("SetStatus") != 0 || u.st.anzahl("Delete") != 0 {
			t.Errorf("%s: SetStatus oder Delete gerufen", name)
		}
	}
	u := bestand()
	c, err := u.cleanup(context.Background())
	pruefe("CleanupBlobs", u, err, c.Entfernt)
	if c != (inbound.Loeschbericht{Entfernt: 1}) {
		t.Errorf("CleanupBlobs: Bericht %+v", c)
	}
	v := bestand()
	b, err := v.startup(context.Background())
	pruefe("Startup", v, err, b.BlobsEntfernt)
}
