package maintenance_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// grosserBestand ist der Bestand, in dem jede Operation von Startup mindestens zweimal gerufen
// wird: drei alte offene Sessions in A, eine junge in B, zwei Referenzen mit Datei, zwei Dateien
// ohne Referenz. Er gehört zu MaxAge 30 d.
func grosserBestand(u *welt) {
	u.dienst(dienstA, dienstB)
	for i, id := range []string{"a1", "a2", "a3"} {
		u.sessions.fuege(id, dienstA, t0().Add(-time.Duration(40-i)*tag), model.SessionOpen)
	}
	u.sessions.fuege("b1", dienstB, t0().Add(-tag), model.SessionComplete)
	u.speicher.fest = []string{r("v1"), r("v2")}
	u.blobs.lege(r("v1"), r("v2"), r("u1"), r("u2"))
}

// startBasis ist die Folge der Aufrufe von Startup auf dem grossen Bestand ohne Fehler.
func startBasis() []string {
	return []string{
		"Check", "FindOpen", "SetStatus(a1)", "SetStatus(a2)", "SetStatus(a3)", "Dienste.List", "Sessions.List",
		"Delete(a1)", "Delete(a2)", "Delete(a3)", "Sessions.List", "ReferencedBlobs", ev("Verify", r("v1")),
		ev("Verify", r("v2")), "Blobs.List", "ReferencedBlobs", "ReferencedBlobs", ev("Remove", r("u1")), ev("Remove", r("u2")),
	}
}

func berichtMit(offene, geloescht, entfernt int) inbound.Wartungsbericht {
	b := leerBericht()
	b.OffeneBeendet, b.SessionsGeloescht, b.BlobsEntfernt = offene, geloescht, entfernt
	return b
}

// TestFehlerJeStufe: ein Fehler eines Ports beendet Startup an der Stelle; die Folge der Aufrufe
// endet mit dem gescheiterten, der Bericht trägt die Zahlen bis dahin, der Code ist ODR-E9001 (W5).
func TestFehlerJeStufe(t *testing.T) {
	basis := startBasis()
	u := neu(t, mitMaxAge(30*tag))
	grosserBestand(u)
	b, err := u.startup(context.Background())
	if err != nil || !reflect.DeepEqual(b, berichtMit(3, 3, 2)) || !slices.Equal(u.p.liste(), basis) {
		t.Fatalf("ohne Fehler: %v %+v\n%v", err, b, u.p.liste())
	}
	faelle := []struct {
		name    string
		op      string
		n       int
		bis     int
		bericht inbound.Wartungsbericht
	}{
		{"Check", "Check", 1, 0, berichtMit(0, 0, 0)},
		{"FindOpen", "FindOpen", 1, 1, berichtMit(0, 0, 0)},
		{"SetStatus zweite von drei", "SetStatus", 2, 3, berichtMit(1, 0, 0)},
		{"Dienste.List", "Dienste.List", 1, 5, berichtMit(3, 0, 0)},
		{"Sessions.List zweiter Service", "Sessions.List", 2, 10, berichtMit(3, 3, 0)},
		{"Delete zweite von drei", "Delete", 2, 8, berichtMit(3, 1, 0)},
		{"ReferencedBlobs der Prüfung", "ReferencedBlobs", 1, 11, berichtMit(3, 3, 0)},
		{"Verify zweite Referenz", "Verify", 2, 13, berichtMit(3, 3, 0)},
		{"Blobs.List", "Blobs.List", 1, 14, berichtMit(3, 3, 0)},
		{"erste Lesung des Abgleichs", "ReferencedBlobs", 2, 15, berichtMit(3, 3, 0)},
		{"zweite Lesung des Abgleichs", "ReferencedBlobs", 3, 16, berichtMit(3, 3, 0)},
		{"Remove zweiter Kandidat", "Remove", 2, 18, berichtMit(3, 3, 1)},
	}
	for _, f := range faelle {
		v := neu(t, mitMaxAge(30*tag))
		grosserBestand(v)
		v.st.scheitere(f.op, f.n, errDisk)
		b, err := v.startup(context.Background())
		code, status := codeVon(t, err)
		if code != "ODR-E9001" || status != 500 || !errors.Is(err, errDisk) {
			t.Errorf("%s: %v", f.name, err)
		}
		if got := v.p.liste(); !slices.Equal(got, basis[:f.bis+1]) {
			t.Errorf("%s: Aufrufe\n%v\nerwartet\n%v", f.name, got, basis[:f.bis+1])
		}
		if !reflect.DeepEqual(b, f.bericht) {
			t.Errorf("%s: Bericht %+v, erwartet %+v", f.name, b, f.bericht)
		}
	}
}

// bestandBereinigung ist ein Bestand ohne Session, in dem CleanupBlobs jede Operation ruft.
func bestandBereinigung(u *welt) {
	u.speicher.fest = []string{r("v1")}
	u.blobs.lege(r("v1"), r("u1"), r("u2"))
}

// TestFehlerJeStufeBereinigung: dasselbe für CleanupBlobs; die Zahl der Entfernten steht im
// Bericht (W5).
func TestFehlerJeStufeBereinigung(t *testing.T) {
	basis := []string{"FindOpen", "Blobs.List", "ReferencedBlobs", "ReferencedBlobs", ev("Remove", r("u1")), ev("Remove", r("u2"))}
	u := neu(t)
	bestandBereinigung(u)
	if b, err := u.cleanup(context.Background()); err != nil || b != (inbound.Loeschbericht{Entfernt: 2}) || !slices.Equal(u.p.liste(), basis) {
		t.Fatalf("ohne Fehler: %v %+v\n%v", err, b, u.p.liste())
	}
	faelle := []struct {
		name     string
		op       string
		n        int
		bis      int
		entfernt int
	}{
		{"FindOpen", "FindOpen", 1, 0, 0},
		{"Blobs.List", "Blobs.List", 1, 1, 0},
		{"erste Lesung", "ReferencedBlobs", 1, 2, 0},
		{"zweite Lesung", "ReferencedBlobs", 2, 3, 0},
		{"Remove zweiter Kandidat", "Remove", 2, 5, 1},
	}
	for _, f := range faelle {
		v := neu(t)
		bestandBereinigung(v)
		v.st.scheitere(f.op, f.n, errDisk)
		b, err := v.cleanup(context.Background())
		code, status := codeVon(t, err)
		if code != "ODR-E9001" || status != 500 || !errors.Is(err, errDisk) {
			t.Errorf("%s: %v", f.name, err)
		}
		if got := v.p.liste(); !slices.Equal(got, basis[:f.bis+1]) {
			t.Errorf("%s: Aufrufe\n%v\nerwartet\n%v", f.name, got, basis[:f.bis+1])
		}
		if b != (inbound.Loeschbericht{Entfernt: f.entfernt}) {
			t.Errorf("%s: Bericht %+v", f.name, b)
		}
	}
}

// TestFehlerCodes: die Abbildung der Fehler eines Ports (SPEC-123 Nr. 10): ErrBeschaedigt aus
// Check und Blobs.List sowie ErrBlobFehlt und ErrBlobBeschaedigt aus Remove sind ODR-E3001, jeder
// andere Fehler ODR-E9001 mit Status 500, auch ErrNichtGefunden und ErrUngueltigerUebergang aus
// SetStatus und ErrSessionGeschlossen aus Delete (W5). Der Test fährt Startup; CleanupBlobs
// prüft er nur für ErrBeschaedigt.
func TestFehlerCodes(t *testing.T) {
	beschaedigt := fmt.Errorf("Datei: %w", outbound.ErrBeschaedigt)
	faelle := []struct {
		name string
		op   string
		fehl error
		code string
	}{
		{"Check beschädigt", "Check", beschaedigt, "ODR-E3001"},
		{"Blobs.List beschädigt", "Blobs.List", beschaedigt, "ODR-E3001"},
		{"Remove Blob fehlt", "Remove", fmt.Errorf("Datei: %w", outbound.ErrBlobFehlt), "ODR-E3001"},
		{"Remove Blob beschädigt", "Remove", fmt.Errorf("Datei: %w", outbound.ErrBlobBeschaedigt), "ODR-E3001"},
		{"Verify weder fehlt noch beschädigt", "Verify", errDisk, "ODR-E9001"},
		{"SetStatus nicht gefunden", "SetStatus", outbound.ErrNichtGefunden, "ODR-E9001"},
		{"SetStatus ungültiger Übergang", "SetStatus", outbound.ErrUngueltigerUebergang, "ODR-E9001"},
		{"Delete Session geschlossen", "Delete", outbound.ErrSessionGeschlossen, "ODR-E9001"},
	}
	for _, f := range faelle {
		u := neu(t, mitMaxAge(30*tag))
		grosserBestand(u)
		u.st.scheitere(f.op, 1, f.fehl)
		_, err := u.startup(context.Background())
		if code, status := codeVon(t, err); code != f.code || status != 500 {
			t.Errorf("Startup, %s: %s, %d (%v)", f.name, code, status, err)
		}
	}
	for _, op := range []string{"Blobs.List", "ReferencedBlobs"} {
		u := neu(t)
		bestandBereinigung(u)
		u.st.scheitere(op, 1, beschaedigt)
		_, err := u.cleanup(context.Background())
		if code, status := codeVon(t, err); code != "ODR-E3001" || status != 500 {
			t.Errorf("CleanupBlobs, %s: %s, %d (%v)", op, code, status, err)
		}
	}
}

func startupFehler(u *welt) error {
	_, err := u.startup(context.Background())
	return err
}

func cleanupFehler(u *welt) error {
	_, err := u.cleanup(context.Background())
	return err
}

// TestReihenfolge: die Stufen laufen in der Reihenfolge Integrität, offene Sessions, Aufbewahrung,
// Blob-Abgleich; die Blobs einer gelöschten Session sind kein Befund (W4).
func TestReihenfolge(t *testing.T) {
	u := neu(t, mitMaxAge(30*tag))
	u.dienst(dienstA)
	u.sessions.fuege("s1", dienstA, t0().Add(-40*tag), model.SessionOpen)
	u.sessions.fuege("s2", dienstA, t0().Add(-35*tag), model.SessionComplete, r("b"), r("g"))
	u.sessions.fuege("s3", dienstA, t0().Add(-tag), model.SessionComplete, r("v"))
	u.blobs.lege(r("b"), r("v"), r("z"))
	u.blobs.beschaedige(r("b"))
	b, err := u.startup(context.Background())
	soll := []string{
		"Check", "FindOpen", "SetStatus(s1)", "Dienste.List", "Sessions.List", "Delete(s1)", "Delete(s2)",
		"ReferencedBlobs", ev("Verify", r("v")), "Blobs.List", "ReferencedBlobs", "ReferencedBlobs",
		ev("Remove", r("b")), ev("Remove", r("z")),
	}
	if got := u.p.liste(); err != nil || !slices.Equal(got, soll) {
		t.Errorf("Aufrufe\n%v\nerwartet\n%v (%v)", got, soll, err)
	}
	if !reflect.DeepEqual(b, berichtMit(1, 2, 2)) {
		t.Errorf("Bericht %+v", b)
	}
	if got := u.blobs.refs(); !slices.Equal(got, []string{r("v")}) {
		t.Errorf("Dateien %v", got)
	}
}

// TestIntegritaet: jedes Element der Liste von Check ist ein Befund; vorher und danach wird
// nichts geschrieben (W6).
func TestIntegritaet(t *testing.T) {
	for _, befunde := range [][]string{{"fremdschluessel verletzt", "index defekt"}, {""}} {
		u := neu(t, mitMaxAge(30*tag))
		grosserBestand(u)
		u.speicher.befunde = befunde
		b, err := u.startup(context.Background())
		code, status := codeVon(t, err)
		if code != "ODR-E3001" || status != 500 || !errors.Is(err, outbound.ErrBeschaedigt) {
			t.Errorf("%q: %v", befunde, err)
		}
		if !slices.Equal(b.Befunde, befunde) || b.OffeneBeendet != 0 {
			t.Errorf("%q: Bericht %+v", befunde, b)
		}
		if got := u.p.liste(); !slices.Equal(got, []string{"Check"}) {
			t.Errorf("%q: Aufrufe %v", befunde, got)
		}
	}
	for _, leer := range [][]string{nil, {}} {
		u := neu(t)
		u.speicher.befunde = leer
		b, err := u.startup(context.Background())
		if err != nil || !reflect.DeepEqual(b, leerBericht()) || u.st.anzahl("FindOpen") != 1 {
			t.Errorf("Liste %v: %v %+v", leer, err, b)
		}
	}
}

// TestOffeneSessions: jede offene Session wird incomplete, in der Reihenfolge von FindOpen, keine
// andere wird angefasst (W7).
func TestOffeneSessions(t *testing.T) {
	u := neu(t)
	u.sessions.fuege("s1", dienstA, t0().Add(-1*time.Hour), model.SessionOpen)
	u.sessions.fuege("s2", dienstB, t0().Add(-2*time.Hour), model.SessionOpen)
	u.sessions.fuege("s3", dienstA, t0().Add(-3*time.Hour), model.SessionComplete)
	u.sessions.fuege("s4", dienstB, t0().Add(-4*time.Hour), model.SessionIncomplete)
	b, err := u.startup(context.Background())
	if err != nil || b.OffeneBeendet != 2 {
		t.Fatalf("%v %+v", err, b)
	}
	if got := u.p.mit("SetStatus"); !slices.Equal(got, []string{"SetStatus(s2)", "SetStatus(s1)"}) {
		t.Errorf("SetStatus %v", got)
	}
	if !slices.Equal(u.sessions.ziele, []model.SessionStatus{model.SessionIncomplete, model.SessionIncomplete}) {
		t.Errorf("Ziele %v", u.sessions.ziele)
	}
	for id, soll := range map[string]model.SessionStatus{"s1": model.SessionIncomplete, "s2": model.SessionIncomplete,
		"s3": model.SessionComplete, "s4": model.SessionIncomplete} {
		if got, _ := u.sessions.status(id); got != soll {
			t.Errorf("%s: %s, erwartet %s", id, got, soll)
		}
	}
	v := neu(t)
	v.sessions.fuege("s3", dienstA, t0().Add(-tag), model.SessionComplete)
	if b, err := v.startup(context.Background()); err != nil || b.OffeneBeendet != 0 || v.st.anzahl("SetStatus") != 0 {
		t.Errorf("ohne offene Session: %v %+v", err, b)
	}
}

// TestStartupOhneSchutz: Startup stellt nichts zurück und kennt keine Karenz (W14).
func TestStartupOhneSchutz(t *testing.T) {
	u := neu(t)
	u.sessions.fuege("s1", dienstA, t0().Add(-time.Hour), model.SessionOpen)
	rn := u.commit("eben committet")
	b, err := u.startup(context.Background())
	if err != nil || b.OffeneBeendet != 1 || b.BlobsEntfernt != 1 || u.blobs.hat(rn) {
		t.Errorf("%v %+v, Blob vorhanden %v", err, b, u.blobs.hat(rn))
	}
	if st, _ := u.sessions.status("s1"); st != model.SessionIncomplete || u.st.anzahl("FindOpen") != 1 {
		t.Errorf("s1 %s, FindOpen %d", st, u.st.anzahl("FindOpen"))
	}
}

// TestBericht: die Berichte zählen je Aufruf, die Listen sind nie nil, ein zweites Startup zählt
// neu und meldet den Befund wieder (W20).
func TestBericht(t *testing.T) {
	u := neu(t, mitMaxAge(30*tag))
	b, err := u.startup(context.Background())
	if err != nil || !reflect.DeepEqual(b, leerBericht()) {
		t.Errorf("leerer Bestand: %v %+v", err, b)
	}
	if c, err := u.cleanup(context.Background()); err != nil || c != (inbound.Loeschbericht{}) {
		t.Errorf("leerer Bestand, CleanupBlobs: %v %+v", err, c)
	}

	v := neu(t, mitMaxAge(30*tag))
	v.dienst(dienstA)
	v.sessions.fuege("o1", dienstA, t0().Add(-time.Hour), model.SessionOpen)
	v.sessions.fuege("o2", dienstA, t0().Add(-2*time.Hour), model.SessionOpen)
	v.sessions.fuege("alt", dienstA, t0().Add(-40*tag), model.SessionComplete)
	v.speicher.fest = []string{r("d")}
	v.blobs.lege(r("u"))
	erster := berichtMit(2, 1, 1)
	erster.BlobsFehlend = []string{r("d")}
	zweiter := leerBericht()
	zweiter.BlobsFehlend = []string{r("d")}
	for i, soll := range []inbound.Wartungsbericht{erster, zweiter} {
		if b, err := v.startup(context.Background()); err != nil || !reflect.DeepEqual(b, soll) {
			t.Errorf("Start %d: %v %+v, erwartet %+v", i+1, err, b, soll)
		}
	}

	w := neu(t)
	w.blobs.lege(r("a"))
	for i, soll := range []int{1, 0} {
		if c, err := w.cleanup(context.Background()); err != nil || c != (inbound.Loeschbericht{Entfernt: soll}) {
			t.Errorf("CleanupBlobs %d: %v %+v", i+1, err, c)
		}
	}
}

// TestFehlerWert: Code, Status, Text und Ursache eines *Fehler, ausgeschrieben (W22); jeder andere
// Fall der Tests läuft durch dieselbe Prüfung (pruefeFehlerWert).
func TestFehlerWert(t *testing.T) {
	faelle := []struct {
		name  string
		setze func(u *welt)
		ruf   func(u *welt) error
		text  string
		ursa  error
	}{
		{"Fehler eines Ports", func(u *welt) { u.st.scheitere("Check", 1, errDisk) }, startupFehler, "ODR-E9001: Internal error", errDisk},
		{"Beschädigt", func(u *welt) { u.st.scheitere("Blobs.List", 1, outbound.ErrBeschaedigt) }, startupFehler,
			"ODR-E3001: Stored data damaged or not loadable", outbound.ErrBeschaedigt},
		{"Befund", func(u *welt) { u.speicher.befunde = []string{"x"} }, startupFehler,
			"ODR-E3001: Stored data damaged or not loadable", outbound.ErrBeschaedigt},
		{"Bereinigung", func(u *welt) { u.st.scheitere("FindOpen", 1, errDisk) }, cleanupFehler, "ODR-E9001: Internal error", errDisk},
	}
	for _, f := range faelle {
		u := neu(t)
		f.setze(u)
		err := f.ruf(u)
		var fe *inbound.Fehler
		if !errors.As(err, &fe) || fe.Status != 500 || fe.Error() != f.text || !errors.Is(err, f.ursa) || fe.Unwrap() != f.ursa {
			t.Errorf("%s: %v", f.name, err)
		}
	}
	if !model.CodeE3001.StatusErlaubt(500) || model.CodeE3001.StatusErlaubt(404) {
		t.Error("ODR-E3001 trägt genau 500")
	}
}

// TestKontext: ein Kontextfehler kehrt unverändert und ohne Code zurück, auch neben einem Fehler
// eines Ports; was bis dahin geschah, bleibt und steht im Bericht (W3).
func TestKontext(t *testing.T) {
	neuerBestand := func() *welt {
		u := neu(t, mitMaxAge(30*tag))
		grosserBestand(u)
		return u
	}
	// (a) ctx vor dem Aufruf beendet.
	u := neuerBestand()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	b, err := u.startup(ctx)
	if err != context.Canceled || !reflect.DeepEqual(b, leerBericht()) {
		t.Errorf("(a) Startup: %v %+v", err, b)
	}
	c, err := u.cleanup(ctx)
	if err != context.Canceled || c != (inbound.Loeschbericht{}) {
		t.Errorf("(a) CleanupBlobs: %v %+v", err, c)
	}
	for _, op := range []string{"SetStatus", "Delete", "Remove"} {
		if n := u.st.anzahl(op); n != 0 {
			t.Errorf("(a) %s %d-mal", op, n)
		}
	}

	// (b) ctx endet im Check: die nächste Stufe liefert den Kontextfehler.
	u = neuerBestand()
	ctx, cancel = context.WithCancel(context.Background())
	u.st.bei("Check", 1, cancel)
	if b, err = u.startup(ctx); err != context.Canceled || !reflect.DeepEqual(b, leerBericht()) || u.st.anzahl("SetStatus") != 0 {
		t.Errorf("(b) %v %+v, SetStatus %d", err, b, u.st.anzahl("SetStatus"))
	}

	// (c) ctx endet im zweiten von drei Delete.
	u = neuerBestand()
	ctx, cancel = context.WithCancel(context.Background())
	u.st.bei("Delete", 2, cancel)
	if b, err = u.startup(ctx); err != context.Canceled || !reflect.DeepEqual(b, berichtMit(3, 1, 0)) || u.st.anzahl("Delete") != 2 {
		t.Errorf("(c) %v %+v, Delete %d", err, b, u.st.anzahl("Delete"))
	}

	// (d) Delete liefert einen Fehler, der beides enthält.
	u = neuerBestand()
	u.st.scheitere("Delete", 1, errors.Join(errDisk, context.Canceled))
	if b, err = u.startup(context.Background()); !errors.Is(err, context.Canceled) || errors.As(err, new(*inbound.Fehler)) ||
		b.SessionsGeloescht != 0 || u.st.anzahl("Delete") != 1 {
		t.Errorf("(d) %v %+v, Delete %d", err, b, u.st.anzahl("Delete"))
	}

	// (e) dasselbe neben ErrReferenziert und neben ErrBlobFehlt: der Kontext gilt.
	u = neuerBestand()
	u.st.scheitere("Delete", 1, errors.Join(outbound.ErrReferenziert, context.DeadlineExceeded))
	if b, err = u.startup(context.Background()); !errors.Is(err, context.DeadlineExceeded) || errors.As(err, new(*inbound.Fehler)) ||
		b.SessionsUebersprungen != 0 || len(u.log.alle()) != 0 {
		t.Errorf("(e) Delete: %v %+v", err, b)
	}
	u = neuerBestand()
	u.st.scheitere("Verify", 1, errors.Join(outbound.ErrBlobFehlt, context.Canceled))
	if b, err = u.startup(context.Background()); !errors.Is(err, context.Canceled) || errors.As(err, new(*inbound.Fehler)) ||
		len(b.BlobsFehlend) != 0 || len(u.log.alle()) != 0 {
		t.Errorf("(e) Verify: %v %+v", err, b)
	}

	// (f) CleanupBlobs: ctx endet im zweiten von zwei Remove.
	v := neu(t)
	bestandBereinigung(v)
	ctx, cancel = context.WithCancel(context.Background())
	v.st.bei("Remove", 2, cancel)
	if c, err = v.cleanup(ctx); err != context.Canceled || c != (inbound.Loeschbericht{Entfernt: 1}) {
		t.Errorf("(f) %v %+v", err, c)
	}
}
